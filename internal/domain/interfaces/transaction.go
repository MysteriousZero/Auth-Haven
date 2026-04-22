package interfaces

import "context"

// TransactionManager defines an interface for managing database transactions
type TransactionManager interface {
	// WithTransaction executes the given function within a database transaction.
	// If the function returns an error, the transaction is rolled back.
	// Otherwise, it is committed.
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
