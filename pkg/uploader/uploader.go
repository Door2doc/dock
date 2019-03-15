package uploader

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/publysher/d2d-uploader/pkg/uploader/config"
	"github.com/publysher/d2d-uploader/pkg/uploader/db"
	"github.com/publysher/d2d-uploader/pkg/uploader/dlog"
	"github.com/publysher/d2d-uploader/pkg/uploader/rest"
)

type Uploader struct {
	Configuration *config.Configuration
	Location      *time.Location

	mu         sync.Mutex
	lastDriver string
	lastDSN    string
	db         *sql.DB
}

// Upload uses a configuration to run a query on the target database, convert the results to FHIR, and upload
// them to the door2doc integration service.
func (u *Uploader) Upload(ctx context.Context) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	// ensure DB connection
	if err := u.ensureDB(); err != nil {
		return err
	}

	// run query
	records, err := u.executeQuery(ctx)
	if err != nil {
		return err
	}

	// convert query to JSON
	vRecs, err := rest.VisitorRecordsFromDB(records, u.Location)
	if err != nil {
		return err
	}

	// upload JSON to upload service
	if err := u.upload(ctx, vRecs); err != nil {
		return err
	}

	return nil
}

func (u *Uploader) ensureDB() error {
	driver, dsn := u.Configuration.DSN()
	if driver == u.lastDriver && dsn == u.lastDSN && u.db != nil {
		return nil
	}

	if u.db != nil {
		dlog.Close(u.db)
	}

	var err error
	u.db, err = sql.Open(driver, dsn)
	if err != nil {
		return err
	}
	u.lastDriver = driver
	u.lastDSN = dsn
	return nil
}

func (u *Uploader) executeQuery(ctx context.Context) ([]db.VisitorRecord, error) {
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			dlog.Error("Failure while rolling back transaction: %v", err)
		}
	}()

	query := u.Configuration.Query()
	records, err := db.ExecuteVisitorQuery(ctx, tx, query)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return records, nil
}

func (u *Uploader) upload(ctx context.Context, recs []rest.VisitorRecord) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(recs); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "https://integration.door2doc.net/services/v1/upload/bezoeken", &buf)
	if err != nil {
		return err
	}
	user, pass := u.Configuration.Credentials()
	req.SetBasicAuth(user, pass)
	req = req.WithContext(ctx)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	var resBuf bytes.Buffer
	_, _ = io.Copy(&resBuf, res.Body)
	_ = res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.Errorf("Unexpected response: %s\n%s", res.Status, resBuf.String())
	}

	return nil
}
