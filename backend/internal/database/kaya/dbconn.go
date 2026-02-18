package kaya

import (
	"context"
	"database/sql"
)

// DBConn represents a database connection interface compatible with *sql.DB and *sql.Tx.
type DBConn interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}
