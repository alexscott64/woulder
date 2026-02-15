package database

import "github.com/alexscott64/woulder/backend/internal/database/dberrors"

// Re-export common database errors for backwards compatibility.
// The actual implementations are in the dberrors package to avoid import cycles.
var (
	// ErrNotFound indicates that the requested record was not found
	ErrNotFound = dberrors.ErrNotFound

	// ErrConflict indicates that a conflict occurred (e.g., unique constraint violation)
	ErrConflict = dberrors.ErrConflict

	// ErrInvalidInput indicates that the provided input was invalid
	ErrInvalidInput = dberrors.ErrInvalidInput

	// ErrTransaction indicates that a transaction error occurred
	ErrTransaction = dberrors.ErrTransaction
)

// WrapNotFound converts sql.ErrNoRows to our standard ErrNotFound.
func WrapNotFound(err error) error {
	return dberrors.WrapNotFound(err)
}

// IsNotFound checks if an error represents a not-found condition
func IsNotFound(err error) bool {
	return dberrors.IsNotFound(err)
}

// IsConflict checks if an error represents a conflict condition
func IsConflict(err error) bool {
	return dberrors.IsConflict(err)
}
