package web

import (
	"fmt"
	"html/template"

	"github.com/publysher/d2d-uploader/pkg/uploader/config"
)

// Humanize turns an error into a human-friendly error message.
func Humanize(err error) interface{} {
	switch err {
	case config.ErrD2DCredentialsNotConfigured:
		return `Username and/or password not configured.`
	case config.ErrD2DCredentialsInvalid:
		return `Invalid username/password.`
	case config.ErrD2DConnectionFailed:
		return template.HTML(`Unable to connect to <a href="https://integration.door2doc.net">integration.door2doc.net</a>. 
			Please make sure the firewall allows outgoing connections to this server.`)
	case config.ErrVisitorQueryNotConfigured:
		return `Visitor query not configured.`
	case config.ErrDatabaseNotConfigured:
		return `Database connection not configured.`
	}

	switch e := err.(type) {
	case config.ErrD2DCredentialsStatus:
		return fmt.Sprintf(`Could not verify credentials: the server returned HTTP %d. Please contact door2doc support.`, e.StatusCode)
	case *config.ErrDatabaseInvalid:
		return fmt.Sprintf(`Could not connect to the database. Driver response: %s.`, e.Cause)
	}

	return fmt.Sprintf(`Unexpected error: %v`, err.Error())
}
