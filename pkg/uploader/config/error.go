package config

import (
	"fmt"

	"github.com/pkg/errors"
)

var (
	ErrDatabaseNotConfigured       = errors.New("database connection not configured")
	ErrVisitorQueryNotConfigured   = errors.New("visitor query not configured")
	ErrD2DConnectionFailed         = errors.New("connection failed")
	ErrD2DCredentialsNotConfigured = errors.New("credentials not configured")
	ErrD2DCredentialsInvalid       = errors.New("credentials invalid")
)

type D2DCredentialsStatusError struct {
	StatusCode int
}

func (err D2DCredentialsStatusError) Error() string {
	return fmt.Sprintf("failed to check credentials: %d", err.StatusCode)
}

type DatabaseInvalidError struct {
	Cause string
}

func (err DatabaseInvalidError) Error() string {
	return fmt.Sprintf("database connection failed: %s", err.Cause)
}
