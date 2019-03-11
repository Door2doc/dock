// Package config manages the configuration options for the uploader.
package config

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/publysher/d2d-uploader/pkg/uploader/dlog"
)

var (
	ErrDatabaseNotConfigured       = errors.New("database connection not configured")
	ErrVisitorQueryNotConfigured   = errors.New("visitor query not configured")
	ErrD2DConnectionFailed         = errors.New("connection failed")
	ErrD2DCredentialsNotConfigured = errors.New("credentials not configured")
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
	// todo
	return false
}

// Configuration contains the configuration options for the service.
type Configuration struct {
	// function used to check accessibility of the upload service.
	checkAccess func() error

	// username to connect to the d2d upload service
	Username string
	// password to connect to the d2d upload service
	Password string
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
	}, nil
}

// Validate validates the configuration and returns the results of those checks.
func (c *Configuration) Validate(ctx context.Context) *ValidationResult {
	res := &ValidationResult{}

	res.D2DConnection = c.validateConnection(ctx)
	res.D2DCredentials = ErrD2DCredentialsNotConfigured
	res.VisitorQuery = ErrVisitorQueryNotConfigured
	res.DatabaseConnection = ErrDatabaseNotConfigured

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

func (c *Configuration) validateConnection(ctx context.Context) error {
	check := c.checkAccess
	if check == nil {
		check = checkConnection
	}

	err := check()
	if err != nil {
		return ErrD2DConnectionFailed
	}
	return nil
}

func checkConnection() error {
	server := "https://integration.door2doc.net"
	res, err := http.Get(server)
	if err != nil {
		dlog.Error("Failed to connect to %s: %v", server, err)
		return err
	}
	_, err = io.Copy(ioutil.Discard, res.Body)
	if err != nil {
		dlog.Error("Failed to drain response: %v", err)
		return err
	}

	dlog.Close(res.Body)
	return nil
}
