package db

import "fmt"

// QueryError indicates a general error while executing the query.
type QueryError struct {
	Cause string
}

func (e *QueryError) Error() string {
	return fmt.Sprintf("general query error: %s", e.Cause)
}

// SelectionError indicates that not all required columns are present in the result set
type SelectionError struct {
	Missing []string
}

func (s *SelectionError) Error() string {
	return fmt.Sprintf("missing columns: %v", s.Missing)
}
