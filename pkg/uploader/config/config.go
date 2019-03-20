package config

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/publysher/d2d-uploader/pkg/uploader/db"
	"github.com/publysher/d2d-uploader/pkg/uploader/dlog"
	"github.com/shibukawa/configdir"
)

var (
	Server     = "https://integration.door2doc.net/"
	configDirs = configdir.New("door2doc", "Upload Service")
)

const (
	PathPing = "/services/v1/upload/ping"

	config = "door2doc.json"
)

// ValidationResult contains the results of validating the current configuration.
type ValidationResult struct {
	DatabaseConnection error
	VisitorQuery       error
	D2DConnection      error
	D2DCredentials     error

	QueryDuration time.Duration
	QueryResults  []db.VisitorRecord
}

// IsValid returns true if all possible validation errors are nil.
func (v *ValidationResult) IsValid() bool {
	return v.DatabaseConnection == nil &&
		v.VisitorQuery == nil &&
		v.D2DConnection == nil &&
		v.D2DCredentials == nil
}

// Configuration contains the configuration options for the service.
type Configuration struct {
	mu sync.RWMutex

	// username to connect to the d2d upload service
	username string
	// password to connect to the d2d upload service
	password string
	// Driver for the database connection
	driver string
	// DSN for the database connection to retrieve visitor information from
	dsn string
	// Query to execute to retrieve visitor information
	query string
	// Set to true if the service should be active
	active bool
	// Pause between runs
	interval time.Duration

	// results of the last call to UpdateValidation
	validationResult *ValidationResult
}

func NewConfiguration() *Configuration {
	return &Configuration{
		active:   true,
		interval: time.Minute,
		driver:   "sqlserver",
	}
}

// Credentials returns the door2doc credentials stored in the configuration.
func (c *Configuration) Credentials() (string, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.username, c.password
}

func (c *Configuration) SetCredentials(username, password string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.username = username
	c.password = password
}

// DSN returns the database driver and DSN stored in the configuration.
func (c *Configuration) DSN() (driver, dsn string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.driver, c.dsn
}

func (c *Configuration) SetDSN(driver, dsn string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.driver = driver
	c.dsn = dsn
}

// Query returns the visitor query stored in the configuration.
func (c *Configuration) Query() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.query
}

func (c *Configuration) SetQuery(query string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.query = query
}

func (c *Configuration) Active() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.active
}

func (c *Configuration) Interval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.interval
}

// UpdateValidation validates the configuration and returns the results of those checks.
func (c *Configuration) UpdateValidation(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	res := &ValidationResult{}

	// configure d2d connection
	res.D2DConnection, res.D2DCredentials = c.checkConnection()
	res.QueryDuration, res.QueryResults, res.DatabaseConnection, res.VisitorQuery = c.checkDatabase(ctx)

	c.validationResult = res
	c.active = c.validationResult.IsValid()
}

// Validate returns the result of the last validation.
func (c *Configuration) Validate() *ValidationResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.validationResult
}

// Reload loads the configuration form a well-known location and updates the values accordingly.
func (c *Configuration) Reload() error {
	folders := configDirs.QueryFolders(configdir.System)
	if len(folders) == 0 {
		return nil
	}
	bs, err := folders[0].ReadFile(config)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if err := json.Unmarshal(bs, &c); err != nil {
		return err
	}
	return nil
}

// Save stores the latest configuration values to a well-known location.
func (c *Configuration) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	bs, err := json.Marshal(c)
	if err != nil {
		return err
	}

	folders := configDirs.QueryFolders(configdir.System)
	if len(folders) == 0 {
		return errors.New("failed to find configuration folder")
	}
	if err := folders[0].WriteFile(config, bs); err != nil {
		return errors.Wrap(err, "while writing configuration file")
	}
	dlog.Info("Updated %s/%s", folders[0].Path, config)
	return nil
}

func (c *Configuration) MarshalJSON() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	vars := map[string]string{
		"username": c.username,
		"password": c.password,
		"driver":   c.driver,
		"dsn":      c.dsn,
		"query":    c.query,
	}
	return json.Marshal(vars)
}

func (c *Configuration) UnmarshalJSON(v []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	vars := make(map[string]string)
	if err := json.Unmarshal(v, &vars); err != nil {
		return err
	}

	c.username = vars["username"]
	c.password = vars["password"]
	c.driver = vars["driver"]
	c.dsn = vars["dsn"]
	c.query = vars["query"]

	return nil
}

func (c *Configuration) checkConnection() (connErr error, credErr error) {
	if c.username == "" || c.password == "" {
		credErr = ErrD2DCredentialsNotConfigured
	}

	req, err := http.NewRequest(http.MethodGet, Server, nil)
	if err != nil {
		dlog.Error("Failed to initialize connection to %s: %v", Server, err)
		connErr = err
		return
	}
	req.URL.Path = PathPing
	req.SetBasicAuth(c.username, c.password)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		dlog.Error("Failed to connect to %s: %v", Server, err)
		connErr = ErrD2DConnectionFailed
		return
	}
	_, err = io.Copy(ioutil.Discard, res.Body)
	if err != nil {
		dlog.Error("Failed to drain response: %v", err)
		connErr = err
		return
	}
	dlog.Close(res.Body)

	if credErr != nil {
		return
	}

	switch res.StatusCode {
	case http.StatusOK:
		credErr = nil
	case http.StatusUnauthorized:
		credErr = ErrD2DCredentialsInvalid
	default:
		credErr = D2DCredentialsStatusError{StatusCode: res.StatusCode}
	}

	return
}

func (c *Configuration) checkDatabase(ctx context.Context) (queryDuration time.Duration, queryResult []db.VisitorRecord, connErr, queryErr error) {
	if c.query == "" {
		queryErr = ErrVisitorQueryNotConfigured
	}
	if c.dsn == "" || c.driver == "" {
		connErr = ErrDatabaseNotConfigured
		return
	}

	conn, err := sql.Open(c.driver, c.dsn)
	if err != nil {
		dlog.Error("Failed to connect to database: %v", err)
		connErr = &DatabaseInvalidError{Cause: err.Error()}
		return
	}

	err = conn.PingContext(ctx)
	if err != nil {
		dlog.Error("Failed to ping database: %v", err)
		connErr = &DatabaseInvalidError{Cause: err.Error()}
		return
	}
	defer func() {
		dlog.Close(conn)
	}()

	if queryErr != nil {
		return
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		dlog.Error("Failed to start read-only transaction: %v", err)
		connErr = &DatabaseInvalidError{Cause: err.Error()}
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			dlog.Error("Failed to roll back transaction: %v", err)
		}
	}()

	queryStart := time.Now()
	records, err := db.ExecuteVisitorQuery(ctx, tx, c.query)
	if err != nil {
		queryErr = err
		return
	}
	queryDuration = time.Since(queryStart)

	max := 10
	if max > len(records) {
		max = len(records)
	}
	queryResult = records[:max]

	return
}
