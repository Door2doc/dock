package web

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/publysher/d2d-uploader/pkg/uploader/config"
)

func TestTemplatesDontFail(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewConfiguration()
	m, err := NewServeMux(false, "testing", cfg)
	if err != nil {
		t.Fatal(err)
	}

	for name, test := range map[string]struct {
		Template *template.Template
		Page     interface{}
	}{
		"status": {
			Template: m.status,
			Page: StatusPage{
				Page: m.page(ctx, "/"),
			},
		},
		"query": {
			Template: m.query,
			Page: QueryPage{
				Page: m.page(ctx, "/"),
			},
		},
		"database": {
			Template: m.database,
			Page: DatabasePage{
				Page: m.page(ctx, "/"),
			},
		},
		"upload": {
			Template: m.upload,
			Page: UploadPage{
				Page: m.page(ctx, "/"),
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			runTemplate(w, test.Template, test.Page)

			res := w.Result()
			if res.StatusCode != http.StatusOK {
				t.Errorf("GET == 200, got %d", res.StatusCode)
			}
		})
	}
}
