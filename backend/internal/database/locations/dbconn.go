package locations

import (
	"context"
	"database/sql"
)

// DBConn abstracts database operations to support both *sql.DB and *sql.Tx.
// This allows repositories to work with either a direct database connection
// or within a transaction, enabling flexible transaction management at the
// service layer while keeping repositories transaction-agnostic.
type DBConn interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}
