package web

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/door2doc/d2d-uploader/pkg/uploader/config"
	"github.com/door2doc/d2d-uploader/pkg/uploader/db"
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
	case config.D2DCredentialsStatusError:
		return fmt.Sprintf(`Could not verify credentials: the server returned HTTP %d. Please contact door2doc support.`, e.StatusCode)
	case *config.DatabaseInvalidError:
		return fmt.Sprintf(`Could not connect to the database. The database driver responded with: %s.`, e.Cause)
	case *config.QueryError:
		return fmt.Sprintf(`Failed to execute query. The database responsed with: %s.`, e.Cause)
	case *db.SelectionError:
		missing := strings.Join(e.Missing, "</code></li><li><code>")
		return template.HTML(fmt.Sprintf(`Query is incomplete. The following columns are missing: <ul><li><code>%s</code></li></ul>`, missing))
	}

	return fmt.Sprintf(`Unexpected error: %v`, err.Error())
}
