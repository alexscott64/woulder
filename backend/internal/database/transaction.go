package database

import (
	"context"
	"database/sql"
	"fmt"
)

// TxFunc is a function that performs operations within a transaction.
// It receives a transaction that can be used to create transactional repositories.
type TxFunc func(tx *sql.Tx) error

// WithTransaction executes fn within a database transaction.
// If fn returns an error, the transaction is rolled back.
// If fn returns nil, the transaction is committed.
//
// Example usage:
//
//	err := database.WithTransaction(ctx, db, func(tx *sql.Tx) error {
//	    weatherRepo := weather.NewPostgresRepository(tx)
//	    return weatherRepo.SaveWeatherData(ctx, data)
//	})
func WithTransaction(ctx context.Context, db *sql.DB, fn TxFunc) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrTransaction, err)
	}

	// Defer rollback in case of panic or error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw panic after rollback
		}
	}()

	// Execute the function
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: commit failed: %v", ErrTransaction, err)
	}

	return nil
}

// WithTransactionOptions executes fn within a transaction with specific options.
// This is useful when you need to control transaction isolation level or read-only mode.
func WithTransactionOptions(ctx context.Context, db *sql.DB, opts *sql.TxOptions, fn TxFunc) error {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrTransaction, err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: commit failed: %v", ErrTransaction, err)
	}

	return nil
}
