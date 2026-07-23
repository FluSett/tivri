package datastore

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}
type roleCtxKey struct{}

type RoleType string

const (
	RolePublic RoleType = "public"
	RoleAdmin  RoleType = "admin"
	RoleSystem RoleType = "system"
)

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

func WithRole(ctx context.Context, role RoleType) context.Context {
	return context.WithValue(ctx, roleCtxKey{}, role)
}

func RoleFrom(ctx context.Context) RoleType {
	if role, ok := ctx.Value(roleCtxKey{}).(RoleType); ok && role != "" {
		return role
	}
	return RolePublic
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

	if tx, ok := reqCtx.Value(txKey{}).(pgx.Tx); ok {
		return fn(context.WithValue(reqCtx, txKey{}, tx))
	}

	tx, err := s.pool.Begin(reqCtx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		_ = tx.Rollback(reqCtx)
	}()

	role := RoleFrom(reqCtx)
	_, err = tx.Exec(reqCtx, "SELECT set_config('app.current_role', $1, true)", string(role))
	if err != nil {
		return fmt.Errorf("set rls role: %w", err)
	}

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
	return s.WithTx(ctx, func(txCtx context.Context) error {
		tx := txCtx.Value(txKey{}).(pgx.Tx)
		_, err := tx.Exec(txCtx, sql, arguments...)
		return err
	})
}

func (s *Store) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx.QueryRow(ctx, sql, args...)
	}

	return &singleRowTxWrapper{
		store: s,
		ctx:   ctx,
		sql:   sql,
		args:  args,
	}
}

type singleRowTxWrapper struct {
	store *Store
	ctx   context.Context
	sql   string
	args  []any
}

func (w *singleRowTxWrapper) Scan(dest ...any) error {
	return w.store.WithTx(w.ctx, func(txCtx context.Context) error {
		tx := txCtx.Value(txKey{}).(pgx.Tx)
		return tx.QueryRow(txCtx, w.sql, w.args...).Scan(dest...)
	})
}

func (s *Store) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx.Query(ctx, sql, args...)
	}
	return s.pool.Query(ctx, sql, args...)
}

