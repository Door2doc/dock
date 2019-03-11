package web

import (
	"reflect"
	"testing"

	"github.com/publysher/d2d-uploader/pkg/uploader/config"
)

func TestHumanize(t *testing.T) {
	for err, want := range map[error]interface{}{
		config.ErrD2DCredentialsNotConfigured:           `Username and/or password not configured.`,
		config.ErrD2DCredentialsStatus{StatusCode: 404}: `Could not verify credentials: the server returned HTTP 404. Please contact door2doc support.`,
		config.ErrDatabaseInvalid{Cause: `argh`}:        `Could not connect to the database. Driver response: argh.`,
	} {
		t.Run(err.Error(), func(t *testing.T) {
			got := Humanize(err)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Humanize() == %v, got %v", want, got)
			}
		})
	}
}
