package datastore

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

const defaultQueryTimeout = 3 * time.Second

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

func ensureTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, defaultQueryTimeout)
}

func (s *Store) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	reqCtx, cancel := ensureTimeout(ctx)
	defer cancel()

	tx, err := s.pool.Begin(reqCtx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		_ = tx.Rollback(reqCtx)
	}()

	txCtx := context.WithValue(reqCtx, txKey{}, tx)

	if err := fn(txCtx); err != nil {
		return err
	}

	if err := tx.Commit(reqCtx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (s *Store) Exec(ctx context.Context, sql string, arguments ...any) error {
	reqCtx, cancel := ensureTimeout(ctx)
	defer cancel()

	if tx, ok := reqCtx.Value(txKey{}).(pgx.Tx); ok {
		_, err := tx.Exec(reqCtx, sql, arguments...)
		return err
	}
	_, err := s.pool.Exec(reqCtx, sql, arguments...)
	return err
}

func (s *Store) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx.QueryRow(ctx, sql, args...)
	}
	return s.pool.QueryRow(ctx, sql, args...)
}

func (s *Store) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	reqCtx, cancel := ensureTimeout(ctx)
	defer cancel()

	if tx, ok := reqCtx.Value(txKey{}).(pgx.Tx); ok {
		return tx.Query(reqCtx, sql, args...)
	}
	return s.pool.Query(reqCtx, sql, args...)
}
