package db

import (
	"errors"
	"fmt"
)

// ErrDuplicateColumns indicates that the query contains duplicate column names.
var ErrDuplicateColumnNames = errors.New("query contains duplicate column names")

// SelectionError indicates that not all required columns are present in the result set
type SelectionError struct {
	Missing []string
}

func (s *SelectionError) Error() string {
	return fmt.Sprintf("missing columns: %v", s.Missing)
}
