package config

import (
	"errors"
	"fmt"
)

var (
	ErrDatabaseNotConfigured       = errors.New("database connection not configured")
	ErrQueryNotConfigured          = errors.New("query not configured")
	ErrD2DConnectionFailed         = errors.New("connection failed")
	ErrD2DCredentialsNotConfigured = errors.New("credentials not configured")
	ErrD2DCredentialsInvalid       = errors.New("credentials invalid")
	ErrAccessNotConfigured         = errors.New("access credentials have not been configured")
	ErrInvalidTimeout              = errors.New("invalid query timeout")
)

// D2DCredentialsStatusError indicates a general error while connecting to the door2doc cloud.
type D2DCredentialsStatusError struct {
	StatusCode int
}

func (err D2DCredentialsStatusError) Error() string {
	return fmt.Sprintf("failed to check credentials: %d", err.StatusCode)
}

// DatabaseInvalid indicates a general error while connecting to the database.
type DatabaseInvalidError struct {
	Cause string
}

func (err DatabaseInvalidError) Error() string {
	return fmt.Sprintf("database connection failed: %s", err.Cause)
}

// QueryError indicates a general error while executing the query.
type QueryError struct {
	Cause string
}

func (e *QueryError) Error() string {
	return fmt.Sprintf("general query error: %s", e.Cause)
}
