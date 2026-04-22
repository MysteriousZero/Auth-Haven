package repository

import (
	"context"
	"database/sql"

	"auth-haven/internal/domain/interfaces"
)

type txManager struct {
	db *sql.DB
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *sql.DB) interfaces.TransactionManager {
	return &txManager{db: db}
}

// WithTransaction executes the given function within a database transaction.
func (m *txManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// If already in a transaction, just execute the function
	if _, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return fn(ctx)
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Create a new context with the transaction injected
	txCtx := context.WithValue(ctx, txKey{}, tx)

	err = fn(txCtx)
	if err != nil {
		// Rollback ignores any rollback errors to return the original execution error
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
