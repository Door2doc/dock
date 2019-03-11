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
	Username string
	// password to connect to the d2d upload service
	Password string
	// Driver for the database connection
	Driver string
	// DSN for the database connection to retrieve visitor information from
	DSN string
	// Query to execute to retrieve visitor information
	Query string
	// Set to true if the service should be active
	Active bool
	// Pause between runs
	Interval time.Duration
}

// Load loads the configuration from a well-known location. It does not give an error if the configuration
// does not exist.
func Load(ctx context.Context) (*Configuration, error) {
	// todo
	return &Configuration{
		Active:   true,
		Interval: time.Minute,
		Driver:   "sqlserver",
	}, nil
}

// Validate validates the configuration and returns the results of those checks.
func (c *Configuration) Validate(ctx context.Context) *ValidationResult {
	res := &ValidationResult{}

	// configure d2d connection
	res.D2DConnection, res.D2DCredentials = c.checkConnection()
	res.DatabaseConnection, res.VisitorQuery = c.checkDatabase(ctx)

	return res
}

// Save stores the latest configuration values to a well-known location.
func (c *Configuration) Save() error {
	// todo
	return nil
}

// Refresh ensures that the configuration is the latest version saved.
func (c *Configuration) Refresh() error {
	// todo
	return nil
}

func (c *Configuration) checkConnection() (connErr error, credErr error) {
	if c.Username == "" || c.Password == "" {
		credErr = ErrD2DCredentialsNotConfigured
	}

	req, err := http.NewRequest(http.MethodGet, Server, nil)
	if err != nil {
		dlog.Error("Failed to initialize connection to %s: %v", Server, err)
		connErr = err
		return
	}
	req.URL.Path = PathPing
	req.SetBasicAuth(c.Username, c.Password)

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
	if c.Query == "" {
		queryErr = ErrVisitorQueryNotConfigured
	}
	if c.DSN == "" || c.Driver == "" {
		connErr = ErrDatabaseNotConfigured
		return
	}

	db, err := sql.Open(c.Driver, c.DSN)
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
