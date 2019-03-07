// Package config manages the configuration options for the uploader.
package config

import (
	"context"
)

// ValidationResult contains the results of a single step in validating the current configuration.
type ValidationResult struct {
	// todo
}

// Configuration contains the configuration options for the service.
type Configuration struct {
	Username string // username to connect to the d2d upload service
	Password string // password to connect to the d2d upload service
	DSN      string // DSN for the database connection to retrieve visitor information from
	Query    string // Query to execute to retrieve visitor information
}

// Load loads the configuration from a well-known location. It does not give an error if the configuration
// does not exist.
func Load(ctx context.Context) (*Configuration, error) {
	panic("todo")
}

// Validate validates the configuration and returns the results of those checks.
func (c *Configuration) Validate(ctx context.Context) []ValidationResult {
	panic("todo")
}

// Save stores the latest configuration values to a well-known location.
func (c *Configuration) Save() error {
	panic("todo")
}
