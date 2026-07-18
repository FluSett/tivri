package datastore

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

// WithTx executes the provided function within a transaction.
// It injects the pgx.Tx into the context so repositories can use it.
func (s *Store) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txCtx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(txCtx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// Exec takes a context and executes the query on the transaction if it exists in the context,
// otherwise it falls back to the pool.
func (s *Store) Exec(ctx context.Context, sql string, arguments ...any) error {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		_, err := tx.Exec(ctx, sql, arguments...)
		return err
	}
	_, err := s.pool.Exec(ctx, sql, arguments...)
	return err
}

// QueryRow runs the query on the transaction if it exists in the context, otherwise on the pool.
func (s *Store) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx.QueryRow(ctx, sql, args...)
	}
	return s.pool.QueryRow(ctx, sql, args...)
}
