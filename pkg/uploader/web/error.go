package web

import (
	"fmt"

	"github.com/publysher/d2d-uploader/pkg/uploader/config"
)

// Humanize turns an error into a human-friendly error message.
func Humanize(err error) string {
	switch err {
	case config.ErrD2DCredentialsNotConfigured:
		return `Username and/or password not configured.`
	}

	return fmt.Sprintf(`Unexpected error: %v`, err.Error())
}
