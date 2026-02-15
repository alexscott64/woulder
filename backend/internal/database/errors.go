package database

import (
	"database/sql"
	"errors"
)

// Common database errors that repositories can return.
// These provide a consistent error interface across all repositories
// and allow service layer to handle errors uniformly.
var (
	// ErrNotFound indicates that the requested record was not found
	ErrNotFound = errors.New("database: record not found")

	// ErrConflict indicates that a conflict occurred (e.g., unique constraint violation)
	ErrConflict = errors.New("database: conflict occurred")

	// ErrInvalidInput indicates that the provided input was invalid
	ErrInvalidInput = errors.New("database: invalid input")

	// ErrTransaction indicates that a transaction error occurred
	ErrTransaction = errors.New("database: transaction error")
)

// WrapNotFound converts sql.ErrNoRows to our standard ErrNotFound.
// This provides a clean abstraction layer between database-specific errors
// and application-level errors.
func WrapNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// IsNotFound checks if an error represents a not-found condition
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound) || errors.Is(err, sql.ErrNoRows)
}

// IsConflict checks if an error represents a conflict condition
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}
