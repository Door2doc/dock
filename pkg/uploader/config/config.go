// Package config manages the configuration options for the uploader.
package config

import (
	"context"
	"database/sql"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/publysher/d2d-uploader/pkg/uploader/dlog"
)

var (
	Server = "https://integration.door2doc.net/"
)

const (
	PathPing = "/services/v1/upload/ping"
)

// ValidationResult contains the results of validating the current configuration.
type ValidationResult struct {
	DatabaseConnection error
	VisitorQuery       error
	D2DConnection      error
	D2DCredentials     error
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
}

func NewConfiguration() *Configuration {
	return &Configuration{
		active:   true,
		interval: time.Minute,
		driver:   "sqlserver",
	}
}

func (c *Configuration) SetCredentials(username, password string) {
	c.username = username
	c.password = password
}

func (c *Configuration) SetDSN(driver, dsn string) {
	c.driver = driver
	c.dsn = dsn
}

func (c *Configuration) SetQuery(query string) {
	c.query = query
}

func (c *Configuration) Active() bool {
	return c.active
}

func (c *Configuration) Interval() time.Duration {
	return c.interval
}

// Validate validates the configuration and returns the results of those checks.
func (c *Configuration) Validate(ctx context.Context) *ValidationResult {
	res := &ValidationResult{}

	// configure d2d connection
	res.D2DConnection, res.D2DCredentials = c.checkConnection()
	res.DatabaseConnection, res.VisitorQuery = c.checkDatabase(ctx)

	return res
}

// Reload loads the configuration form a well-known location and updates the values accordingly.
func (c *Configuration) Reload() error {
	// todo
	return nil
}

// Save stores the latest configuration values to a well-known location.
func (c *Configuration) Save() error {
	// todo
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
		credErr = ErrD2DCredentialsStatus{StatusCode: res.StatusCode}
	}

	return
}

func (c *Configuration) checkDatabase(ctx context.Context) (connErr, queryErr error) {
	if c.query == "" {
		queryErr = ErrVisitorQueryNotConfigured
	}
	if c.dsn == "" || c.driver == "" {
		connErr = ErrDatabaseNotConfigured
		return
	}

	db, err := sql.Open(c.driver, c.dsn)
	if err != nil {
		dlog.Error("Failed to connect to database: %v", err)
		connErr = &ErrDatabaseInvalid{Cause: err.Error()}
		return
	}

	err = db.PingContext(ctx)
	if err != nil {
		dlog.Error("Failed to ping database: %v", err)
		connErr = &ErrDatabaseInvalid{Cause: err.Error()}
		return
	}
	dlog.Close(db)

	if queryErr != nil {
		return
	}

	return nil, nil
}
