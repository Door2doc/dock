package web

import (
	"html/template"
	"reflect"
	"testing"

	"github.com/publysher/d2d-uploader/pkg/uploader/config"
	"github.com/publysher/d2d-uploader/pkg/uploader/db"
)

func TestHumanize(t *testing.T) {
	for err, want := range map[error]interface{}{
		config.ErrD2DCredentialsNotConfigured:                   `Username and/or password not configured.`,
		config.D2DCredentialsStatusError{StatusCode: 404}:       `Could not verify credentials: the server returned HTTP 404. Please contact door2doc support.`,
		&config.DatabaseInvalidError{Cause: `argh`}:             `Could not connect to the database. The database driver responded with: argh.`,
		&db.SelectionError{Missing: []string{"hello", "world"}}: template.HTML(`Query is incomplete. The following columns are missing: <ul><li><code>hello</code></li><li><code>world</code></li></ul>`),
	} {
		t.Run(err.Error(), func(t *testing.T) {
			got := Humanize(err)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Humanize() == %v, got %v", want, got)
			}
		})
	}
}
