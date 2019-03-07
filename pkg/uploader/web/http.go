package web

import (
	"html/template"
	"io/ioutil"
	"net/http"
)

// NewServeMux generates the toplevel http mux for managing the service.
func NewServeMux() http.Handler {
	res := http.NewServeMux()
	res.Handle("/assets/", http.FileServer(FS(false)))
	res.Handle("/", ConfigurationHandler())
	return res
}

func load(name string) *template.Template {
	r, err := FS(false).Open(name)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	text, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}

	return template.Must(template.New(name).Parse(string(text)))
	// return nil
}

func ConfigurationHandler() http.Handler {
	tmpl := load("/index.html")
	_ = tmpl

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	})
}
