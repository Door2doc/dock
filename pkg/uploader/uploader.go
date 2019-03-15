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
	"github.com/publysher/d2d-uploader/pkg/uploader/history"
	"github.com/publysher/d2d-uploader/pkg/uploader/rest"
)

type Uploader struct {
	Configuration *config.Configuration
	Location      *time.Location
	History       *history.History

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

	evt := u.History.NewEvent()

	// ensure DB connection
	if err := u.ensureDB(); err != nil {
		evt.Error = err
		return err
	}

	// run query
	start := time.Now()
	records, err := u.executeQuery(ctx)
	if err != nil {
		evt.Error = err
		return err
	}
	evt.QueryDuration = time.Since(start)

	// convert query to JSON
	vRecs, err := rest.VisitorRecordsFromDB(records, u.Location)
	if err != nil {
		evt.Error = err
		return err
	}

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(vRecs); err != nil {
		return err
	}
	evt.JSON = buf.String()

	// upload JSON to upload service
	start = time.Now()
	if err := u.upload(ctx, buf); err != nil {
		evt.Error = err
		return err
	}
	evt.UploadDuration = time.Since(start)

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

func (u *Uploader) upload(ctx context.Context, json *bytes.Buffer) error {
	req, err := http.NewRequest(http.MethodPost, "https://integration.door2doc.net/services/v1/upload/bezoeken", json)
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
