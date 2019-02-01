package postgres

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

var (
	// ErrNoTransaction should be returned if no transaction on ctx is active
	ErrNoTransaction = errors.New("no transaction active")
)

// Transactor is used to finalize an active transaction.
type Transactor interface {
	// Commit commits any changes made during the transaction. On error a
	// caller is expected to clean up any resources which would have relied
	// on data mutated as part of this transaction.
	Commit() error

	// Rollback rolls back any changes made during the transaction.
	Rollback() error
}

// transactionKey is used for the context.
type transactionKey struct{}

// TransactionHandler manages transactions.
type TransactionHandler struct {
	db *sqlx.DB
}

// NewTransactionHandler creates a transactionHandler.
func NewTransactionHandler(db *sqlx.DB) *TransactionHandler {
	return &TransactionHandler{
		db: db,
	}
}

// TransactionContext creates a new transaction context.
func (r *TransactionHandler) TransactionContext(ctx context.Context) (context.Context, Transactor, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	ctx = context.WithValue(ctx, transactionKey{}, tx)
	return ctx, tx, nil
}

// Transaction extract the transaction from the context.
func Transaction(ctx context.Context) (*sqlx.Tx, bool) {
	tx, ok := ctx.Value(transactionKey{}).(*sqlx.Tx)
	return tx, ok
}

// NamedStmt uses the context and converts the named stmt to an
// transaction stmt if possible.
func NamedStmt(ctx context.Context, stmt *sqlx.NamedStmt) *sqlx.NamedStmt {
	if tx, ok := Transaction(ctx); ok {
		return tx.NamedStmt(stmt)
	}
	return stmt
}

// Stmt uses the context and converts the stmt to an
// transaction stmt if possible.
func Stmt(ctx context.Context, stmt *sqlx.Stmt) *sqlx.Stmt {
	if tx, ok := Transaction(ctx); ok {
		return tx.Stmtx(stmt)
	}
	return stmt
}
