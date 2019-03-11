package web

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/publysher/d2d-uploader/pkg/uploader/config"
)

type Log interface {
	Errorf(pattern string, args ...interface{}) error
}

type mux struct {
	*http.ServeMux

	log     Log
	fs      http.FileSystem
	version string

	mu       sync.RWMutex
	err      error
	database *template.Template
	query    *template.Template
	status   *template.Template
	upload   *template.Template
}

var (
	urlRe      = regexp.MustCompile(`https://[a-zA-Z0-9/.]+`)
	sentenceRe = regexp.MustCompile(`(?:^|\. )[a-z]`)
)

func sentence(s interface{}) template.HTML {
	base := fmt.Sprintf("%s", s)
	base = urlRe.ReplaceAllStringFunc(base, func(v string) string {
		return fmt.Sprintf("<a href=\"%s\">%s</a>", v, v)
	})
	if !strings.HasPrefix(base, ".") {
		base = base + "."
	}
	base = strings.Replace(base, ";", ".", -1)
	base = sentenceRe.ReplaceAllStringFunc(base, func(v string) string {
		return strings.ToUpper(v)
	})

	return template.HTML(base)
}

func (m *mux) load(templates ...string) *template.Template {
	res := template.New(templates[0])
	res = res.Funcs(template.FuncMap{
		"sentence": sentence,
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
					_ = m.log.Errorf("Failed to close %s: %v", name, err)
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

func (m *mux) initTemplates() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.database = m.load("/database.html", "/_layout.html")
	m.query = m.load("/query.html", "/_layout.html")
	m.status = m.load("/status.html", "/_layout.html")
	m.upload = m.load("/upload.html", "/_layout.html")
}

// NewServeMux generates the toplevel http mux for managing the service.
func NewServeMux(l Log, dev bool, version string) (http.Handler, error) {
	res := &mux{
		ServeMux: http.NewServeMux(),
		log:      l,
		fs:       FS(dev),
		version:  version,
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
	res.Handle("/database", res.DatabaseHandler())
	res.Handle("/query", res.QueryHandler())
	res.Handle("/upload", res.UploadHandler())
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

func (m *mux) page(r *http.Request) *Page {
	p := &Page{
		Version: m.version,
		Path:    r.URL.Path,
	}

	p.Configuration, p.GlobalError = config.Load(r.Context())
	if p.GlobalError != nil {
		return p
	}

	p.Validation = p.Configuration.Validate(r.Context())
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

func (m *mux) StatusHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		p := &StatusPage{
			Page: m.page(r),
		}

		if err := m.status.Execute(w, p); err != nil {
			_ = m.log.Errorf("while serving index page: %v", err)
		}
	})
}

type DatabasePage struct {
	*Page
}

func (m *mux) DatabaseHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		p := DatabasePage{
			Page: m.page(r),
		}

		if err := m.database.Execute(w, p); err != nil {
			_ = m.log.Errorf("while serving index page: %v", err)
		}
	})
}

type UploadPage struct {
	*Page
}

func (m *mux) UploadHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		p := UploadPage{
			Page: m.page(r),
		}

		if err := m.upload.Execute(w, p); err != nil {
			_ = m.log.Errorf("while serving index page: %v", err)
		}
	})
}

type QueryPage struct {
	*Page
}

func (m *mux) QueryHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		p := QueryPage{
			Page: m.page(r),
		}

		if err := m.query.Execute(w, p); err != nil {
			_ = m.log.Errorf("while serving index page: %v", err)
		}
	})
}
