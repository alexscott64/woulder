package analytics

import (
	"context"
	"database/sql"
)

// DBConn abstracts database operations to support both *sql.DB and *sql.Tx.
type DBConn interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}
