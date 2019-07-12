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

	"github.com/door2doc/d2d-uploader/pkg/uploader/config"
	"github.com/door2doc/d2d-uploader/pkg/uploader/db"
	"github.com/door2doc/d2d-uploader/pkg/uploader/dlog"
	"github.com/door2doc/d2d-uploader/pkg/uploader/history"
	"github.com/door2doc/d2d-uploader/pkg/uploader/rest"
	"github.com/pkg/errors"
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

// Upload uses a configuration to run a query on the target database, convert the results to JSON, and upload
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
	evt.Size = len(vRecs)

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
	conn := u.Configuration.Connection()
	driver, dsn := conn.Driver, conn.DSN()
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
	tx, err := u.db.BeginTx(ctx, nil)
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
	req, err := http.NewRequest(http.MethodPost, config.Server, json)
	if err != nil {
		return err
	}
	req.URL.Path = config.PathUpload

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "close")

	res, err := u.Configuration.Do(ctx, req)
	if err != nil {
		return err
	}

	var resBuf bytes.Buffer
	_ = res.Header.Write(&resBuf)
	_, _ = resBuf.WriteRune('\n')

	_, _ = io.Copy(&resBuf, res.Body)
	_ = res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.Errorf("Unexpected response: %s\n%s", res.Status, resBuf.String())
	}

	return nil
}
