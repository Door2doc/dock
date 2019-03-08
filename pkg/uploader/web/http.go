package web

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type Log interface {
	Errorf(pattern string, args ...interface{}) error
}

type mux struct {
	*http.ServeMux

	log     Log
	fs      http.FileSystem
	version string

	mu    sync.RWMutex
	err   error
	index *template.Template
}

func (m *mux) load(name string) *template.Template {
	r, err := m.fs.Open(name)
	if err != nil {
		m.err = err
		return nil
	}
	defer func() {
		if err := r.Close(); err != nil {
			_ = m.log.Errorf("Failed to close %s: %v", name, err)
		}
	}()

	text, err := ioutil.ReadAll(r)
	if err != nil {
		m.err = err
		return nil
	}

	res, err := template.New(name).Parse(string(text))
	if err != nil {
		m.err = err
		return nil
	}

	return res
}

func (m *mux) initTemplates() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.index = m.load("/index.html")
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
	res.Handle("/", res.ConfigurationHandler())
	return res, nil
}

type ConfigurationPage struct {
	Version string
}

func (m *mux) ConfigurationHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		p := ConfigurationPage{
			Version: m.version,
		}

		if err := m.index.Execute(w, p); err != nil {
			_ = m.log.Errorf("while serving index page: %v", err)
		}
	})
}
