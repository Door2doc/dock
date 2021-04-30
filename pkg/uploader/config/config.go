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

	"github.com/door2doc/d2d-uploader/pkg/uploader/db"
	"github.com/door2doc/d2d-uploader/pkg/uploader/dlog"
	"github.com/door2doc/d2d-uploader/pkg/uploader/rest"
	"github.com/pkg/errors"
	"github.com/shibukawa/configdir"
)

var (
	Server     = "https://integration.door2doc.net/"
	configDirs = configdir.New("door2doc", "Upload Service")
)

const (
	PathPing            = "/services/v3/upload/ping"
	PathUpload          = "/services/v3/upload/bezoeken"
	DBValidationTimeout = 5 * time.Second

	config = "door2doc.json"
)

// ValidationResult contains the results of validating the current configuration.
type ValidationResult struct {
	DatabaseConnection error
	VisitorQuery       error
	D2DConnection      error
	D2DCredentials     error

	QueryDuration time.Duration
	QueryResults  db.VisitorRecords

	Access error
}

// IsValid returns true if all fatal validation errors are nil.
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
	// database connection data
	connection db.ConnectionData
	// Query to execute to retrieve visitor information
	query string
	// Set to true if the service should be active
	active bool
	// Pause between runs
	interval time.Duration
	// proxy server to use for all HTTP requests
	proxy string

	// username to access the web interface
	accessUsername string
	// password to access the web interface
	accessPassword string

	// results of the last call to UpdateValidation
	validationResult *ValidationResult
}

func NewConfiguration() *Configuration {
	return &Configuration{
		active:   true,
		interval: time.Minute,
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

func (c *Configuration) Proxy() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.proxy
}

func (c *Configuration) SetProxy(proxy string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.proxy = proxy
}

// Connection returns the connection data stored in the configuration.
func (c *Configuration) Connection() db.ConnectionData {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connection
}

func (c *Configuration) SetConnection(cd db.ConnectionData) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connection = cd
}

func (c *Configuration) SetDSN(driver, dsn string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_ = c.connection.UnmarshalText([]byte(dsn))
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

func (c *Configuration) AccessCredentials() (username, password string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessUsername, c.accessPassword
}

func (c *Configuration) SetAccessCredentials(username, password string) {
	c.mu.Lock()
	c.mu.Unlock()

	if username == "" || password == "" {
		username = ""
		password = ""
	}
	c.accessUsername = username
	c.accessPassword = password
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

	// check d2d connection
	connCtx, timeout := context.WithTimeout(ctx, DBValidationTimeout)
	defer timeout()

	res.D2DConnection, res.D2DCredentials = c.checkConnection(connCtx)

	if c.accessUsername == "" && c.accessPassword == "" {
		res.Access = ErrAccessNotConfigured
	}

	// check db connection
	dbCtx, timeout := context.WithTimeout(ctx, DBValidationTimeout)
	defer timeout()
	res.QueryDuration, res.QueryResults, res.DatabaseConnection, res.VisitorQuery = c.checkDatabase(dbCtx)
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

	bs, err := json.MarshalIndent(c, "", "  ")
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

type persistentConfig struct {
	Username       string            `json:"username"`
	Password       string            `json:"password"`
	Proxy          string            `json:"proxy"`
	Dsn            db.ConnectionData `json:"dsn"`
	Query          string            `json:"query"`
	AccessUsername string            `json:"accessUsername"`
	AccessPassword string            `json:"accessPassword"`
}

func (c *Configuration) MarshalJSON() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	vars := persistentConfig{
		Username:       c.username,
		Password:       c.password,
		Proxy:          c.proxy,
		Dsn:            c.connection,
		Query:          c.query,
		AccessUsername: c.accessUsername,
		AccessPassword: c.accessPassword,
	}
	return json.Marshal(vars)
}

func (c *Configuration) UnmarshalJSON(v []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	vars := &persistentConfig{}
	if err := json.Unmarshal(v, &vars); err != nil {
		return err
	}

	c.username = vars.Username
	c.password = vars.Password
	c.proxy = vars.Proxy
	c.connection = vars.Dsn
	c.query = vars.Query
	c.accessUsername = vars.AccessUsername
	c.accessPassword = vars.AccessPassword

	return nil
}

func (c *Configuration) checkConnection(ctx context.Context) (connErr error, credErr error) {
	if c.username == "" || c.password == "" {
		credErr = ErrD2DCredentialsNotConfigured
	}

	req, err := http.NewRequest(http.MethodGet, Server, nil)
	if err != nil {
		dlog.Error("Failed to initialize connection to %s: %v", Server, err)
		return err, credErr
	}
	req.URL.Path = PathPing
	req.SetBasicAuth(c.username, c.password)

	res, err := rest.Do(ctx, c.proxy, req)
	if err != nil {
		dlog.Error("Failed to connect to %s: %v", Server, err)
		return ErrD2DConnectionFailed, credErr
	}
	_, err = io.Copy(ioutil.Discard, res.Body)
	if err != nil {
		dlog.Error("Failed to drain response: %v", err)
		return err, credErr
	}
	dlog.Close(res.Body)

	if credErr != nil {
		return nil, credErr
	}

	switch res.StatusCode {
	case http.StatusOK:
		dlog.Info("Ping successful")
		return nil, nil
	case http.StatusUnauthorized:
		dlog.Info("Ping not authorized")
		return nil, ErrD2DCredentialsInvalid
	default:
		dlog.Info("Ping failed")
		return nil, D2DCredentialsStatusError{StatusCode: res.StatusCode}
	}
}

func (c *Configuration) checkDatabase(ctx context.Context) (queryDuration time.Duration, queryResult []db.VisitorRecord, connErr, queryErr error) {
	if c.query == "" {
		queryErr = ErrVisitorQueryNotConfigured
	}

	if !c.connection.IsValid() {
		connErr = ErrDatabaseNotConfigured
		return
	}

	conn, err := sql.Open(c.connection.Driver, c.connection.DSN())
	if err != nil {
		dlog.Error("Failed to connect to database: %v", err)
		connErr = &DatabaseInvalidError{Cause: err.Error()}
		return
	}

	err = conn.PingContext(ctx)
	if err != nil {
		dlog.Error("Failed to ping database %s: %v", c.connection.DSN(), err)
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
	_, errIsSelection := err.(*db.SelectionError)

	switch {
	case err == nil:
	case errIsSelection:
		queryErr = err
		return
	case err == context.DeadlineExceeded:
		queryErr = err
		return
	default:
		queryErr = &QueryError{Cause: err.Error()}
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

func (c *Configuration) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	req.SetBasicAuth(c.username, c.password)
	return rest.Do(ctx, c.proxy, req)
}
