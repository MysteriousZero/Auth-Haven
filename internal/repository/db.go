package repository

import (
	"context"
	"database/sql"
)

// DBTX is an interface that encompasses both *sql.DB and *sql.Tx
// This allows repositories to seamlessly execute queries within or outside a transaction.
type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type txKey struct{}

// getDB extracts a transaction from context if it exists, otherwise returns the default DB.
func getDB(ctx context.Context, defaultDB *sql.DB) DBTX {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return defaultDB
}
