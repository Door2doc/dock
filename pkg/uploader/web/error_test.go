package web

import (
	"reflect"
	"testing"

	"github.com/publysher/d2d-uploader/pkg/uploader/config"
)

func TestHumanize(t *testing.T) {
	for err, want := range map[error]interface{}{
		config.ErrD2DCredentialsNotConfigured: `Username and/or password not configured.`,
	} {
		t.Run(err.Error(), func(t *testing.T) {
			got := Humanize(err)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Humanize() == %v, got %v", want, got)
			}
		})
	}
}
