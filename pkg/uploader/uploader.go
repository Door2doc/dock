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
func (u *Uploader) Upload(ctx context.Context) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if err := u.upload(ctx, config.PathVisitorUpload, u.executeVisitorQuery); err != nil {
		dlog.Error("While processing visitor upload: %v", err)
	}

	if u.Configuration.RadiologieQuery() != "" {
		if err := u.upload(ctx, config.PathRadiologieUpload, u.executeRadiologieQuery); err != nil {
			dlog.Error("While processing radiologie upload: %v", err)
		}
	}
	if u.Configuration.LabQuery() != "" {
		if err := u.upload(ctx, config.PathLabUpload, u.executeLabQuery); err != nil {
			dlog.Error("While processing lab upload: %v", err)
		}
	}
	if u.Configuration.ConsultQuery() != "" {
		if err := u.upload(ctx, config.PathConsultUpload, u.executeConsultQuery); err != nil {
			dlog.Error("While processing consult upload: %v", err)
		}
	}
}

type queryFunc func(ctx context.Context) (interface{}, int, error)

func (u *Uploader) upload(ctx context.Context, path string, q queryFunc) error {
	evt := u.History.NewEvent(path)

	// ensure DB connection
	if err := u.ensureDB(); err != nil {
		evt.Error = err
		return err
	}

	// run query
	start := time.Now()
	vRecs, size, err := q(ctx)
	if err != nil {
		evt.Error = err
		return err
	}
	evt.QueryDuration = time.Since(start)
	evt.Size = size

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(vRecs); err != nil {
		return err
	}
	evt.JSON = buf.String()

	// upload JSON to upload service
	start = time.Now()
	if err := u.UploadJSON(ctx, buf, path, false); err != nil {
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

func (u *Uploader) executeVisitorQuery(ctx context.Context) (interface{}, int, error) {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			dlog.Error("Failure while rolling back transaction: %v", err)
		}
	}()

	records, err := db.ExecuteVisitorQuery(ctx, tx, u.Configuration.VisitorQuery())
	if err != nil {
		return nil, 0, err
	}
	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}

	// convert query to JSON
	vRecs, err := rest.VisitorRecordsFromDB(records, u.Location)
	if err != nil {
		return nil, 0, err
	}
	return vRecs, len(vRecs), nil
}

func (u *Uploader) executeRadiologieQuery(ctx context.Context) (interface{}, int, error) {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			dlog.Error("Failure while rolling back transaction: %v", err)
		}
	}()

	records, err := db.ExecuteRadiologieQuery(ctx, tx, u.Configuration.RadiologieQuery())
	if err != nil {
		return nil, 0, err
	}
	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}

	// convert query to JSON
	vRecs, err := rest.RadiologieRecordsFromDB(records, u.Location)
	if err != nil {
		return nil, 0, err
	}
	return vRecs, len(vRecs), nil
}

func (u *Uploader) executeLabQuery(ctx context.Context) (interface{}, int, error) {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			dlog.Error("Failure while rolling back transaction: %v", err)
		}
	}()

	records, err := db.ExecuteLabQuery(ctx, tx, u.Configuration.LabQuery())
	if err != nil {
		return nil, 0, err
	}
	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}

	// convert query to JSON
	vRecs, err := rest.LabRecordsFromDB(records, u.Location)
	if err != nil {
		return nil, 0, err
	}
	return vRecs, len(vRecs), nil
}

func (u *Uploader) executeConsultQuery(ctx context.Context) (interface{}, int, error) {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			dlog.Error("Failure while rolling back transaction: %v", err)
		}
	}()

	records, err := db.ExecuteConsultQuery(ctx, tx, u.Configuration.ConsultQuery())
	if err != nil {
		return nil, 0, err
	}
	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}

	// convert query to JSON
	vRecs, err := rest.ConsultRecordsFromDB(records, u.Location)
	if err != nil {
		return nil, 0, err
	}
	return vRecs, len(vRecs), nil
}

func (u *Uploader) UploadJSON(ctx context.Context, json *bytes.Buffer, path string, importMode bool) error {
	req, err := http.NewRequest(http.MethodPost, config.Server, json)
	if err != nil {
		return err
	}
	req.URL.Path = path

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "close")

	if importMode {
		req.URL.RawQuery = "import=true"
	}

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
