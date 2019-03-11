package web

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/publysher/d2d-uploader/pkg/uploader/config"
	"github.com/publysher/d2d-uploader/pkg/uploader/dlog"
)

const (
	pathUpload   = "/upload"
	pathDatabase = "/database"
	pathQuery    = "/query"
)

type ServeMux struct {
	*http.ServeMux

	fs      http.FileSystem
	version string
	cfg     *config.Configuration

	mu       sync.RWMutex
	err      error
	database *template.Template
	query    *template.Template
	status   *template.Template
	upload   *template.Template
}

func (m *ServeMux) load(templates ...string) *template.Template {
	res := template.New(templates[0])
	res = res.Funcs(template.FuncMap{
		"humanize": Humanize,
	})

	for _, name := range templates {
		err := func() error {
			r, err := m.fs.Open(name)
			if err != nil {
				return err
			}
			defer func() {
				if err := r.Close(); err != nil {
					dlog.Error("Failed to close %s: %v", name, err)
				}
			}()

			text, err := ioutil.ReadAll(r)
			if err != nil {
				return err
			}

			res, err = res.Parse(string(text))
			if err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			m.err = err
			return nil
		}
	}

	return res
}

func (m *ServeMux) initTemplates() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.database = m.load("/database.html", "/_layout.html")
	m.query = m.load("/query.html", "/_layout.html")
	m.status = m.load("/status.html", "/_layout.html")
	m.upload = m.load("/upload.html", "/_layout.html")
}

func runTemplate(w http.ResponseWriter, tmpl *template.Template, data interface{}) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		dlog.Error("Error while writing template %s: %v", tmpl.Name(), err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(buf.Bytes()); err != nil {
		dlog.Error("Error while writing response: %v", err)
	}
}

// NewServeMux generates the toplevel http mux for managing the service.
func NewServeMux(dev bool, version string, cfg *config.Configuration) (*ServeMux, error) {
	res := &ServeMux{
		ServeMux: http.NewServeMux(),
		fs:       FS(dev),
		version:  version,
		cfg:      cfg,
	}

	res.initTemplates()
	if res.err != nil {
		return nil, res.err
	}

	if dev {
		go func() {
			for {
				select {
				case <-time.After(time.Second):
					res.initTemplates()
				}
			}
		}()
	}

	res.Handle("/assets/", http.FileServer(res.fs))
	res.Handle("/", res.StatusHandler())
	res.Handle(pathDatabase, res.DatabaseHandler())
	res.Handle(pathQuery, res.QueryHandler())
	res.Handle(pathUpload, res.UploadHandler())
	return res, nil
}

type Page struct {
	Version       string
	Path          string
	Problems      map[string]bool
	GlobalError   error
	Validation    *config.ValidationResult
	Configuration *config.Configuration
}

func (m *ServeMux) page(ctx context.Context, path string) *Page {
	p := &Page{
		Version: m.version,
		Path:    path,
	}

	p.Configuration = m.cfg
	p.Validation = p.Configuration.Validate()
	p.Problems = map[string]bool{
		"Database": p.Validation.DatabaseConnection != nil,
		"Query":    p.Validation.VisitorQuery != nil,
		"Upload":   p.Validation.D2DCredentials != nil,
	}

	return p
}

type StatusPage struct {
	*Page
}

func (m *ServeMux) StatusHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		runTemplate(w, m.status, StatusPage{
			Page: m.page(r.Context(), r.URL.Path),
		})
	})
}

type DatabasePage struct {
	*Page

	Driver string
	DSN    string
	Error  error
}

func (m *ServeMux) DatabaseHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		if r.Method == http.MethodPost {
			m.cfg.SetDSN(r.FormValue("driver"), r.FormValue("dsn"))
			m.cfg.UpdateValidation(r.Context())
			if err := m.cfg.Save(); err != nil {
				dlog.Error("While saving credentials: %v", err)
			}

			w.Header().Set("Location", pathDatabase)
			w.WriteHeader(http.StatusFound)
			return
		}

		driver, dsn := m.cfg.DSN()
		runTemplate(w, m.database, DatabasePage{
			Page:   m.page(r.Context(), r.URL.Path),
			Driver: driver,
			DSN:    dsn,
			Error:  m.cfg.Validate().DatabaseConnection,
		})
	})
}

type UploadPage struct {
	*Page

	Username string
	Password string
	Error    error
}

func (m *ServeMux) UploadHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		if r.Method == http.MethodPost {
			m.cfg.SetCredentials(r.FormValue("username"), r.FormValue("password"))
			m.cfg.UpdateValidation(r.Context())
			if err := m.cfg.Save(); err != nil {
				dlog.Error("While saving credentials: %v", err)
			}

			w.Header().Set("Location", pathUpload)
			w.WriteHeader(http.StatusFound)
			return
		}

		username, password := m.cfg.Credentials()
		err := m.cfg.Validate().D2DCredentials

		runTemplate(w, m.upload, UploadPage{
			Page:     m.page(r.Context(), r.URL.Path),
			Username: username,
			Password: password,
			Error:    err,
		})
	})
}

type QueryPage struct {
	*Page
	Query string
	Error error
}

func (m *ServeMux) QueryHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		if r.Method == http.MethodPost {
			m.cfg.SetQuery(r.FormValue("query"))
			m.cfg.UpdateValidation(r.Context())
			if err := m.cfg.Save(); err != nil {
				dlog.Error("While saving query: %v", err)
			}

			w.Header().Set("Location", pathQuery)
			w.WriteHeader(http.StatusFound)
			return
		}

		runTemplate(w, m.query, QueryPage{
			Page:  m.page(r.Context(), r.URL.Path),
			Query: m.cfg.Query(),
			Error: m.cfg.Validate().VisitorQuery,
		})
	})
}
