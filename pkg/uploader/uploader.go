package uploader

import (
	"context"

	"github.com/publysher/d2d-uploader/pkg/uploader/config"
)

type Log interface {
	Infof(pattern string, args ...interface{}) error
}

// Upload uses a configuration to run a query on the target database, convert the results to FHIR, and upload
// them to the door2doc integration service.
func Upload(ctx context.Context, log Log, c *config.Configuration) error {
	// open database connection: TODO

	// run query: TODO

	// convert query to JSON: TODO

	// upload JSON to upload service: TODO
	_ = log.Infof("TODO...")

	return nil
}
