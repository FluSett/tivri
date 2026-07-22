package database

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"io/fs"
)

const (
	defaultMaxConnIdleTime = 15 * time.Minute
	defaultMaxConnLifetime = 1 * time.Hour
)

func Connect(ctx context.Context, dsn string, maxConns, minConns int32) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("database: parse config failed: %w", err)
	}

	config.MaxConns = maxConns
	config.MinConns = minConns
	config.MaxConnIdleTime = defaultMaxConnIdleTime
	config.MaxConnLifetime = defaultMaxConnLifetime
	config.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("database: connect failed: %w", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("database: ping failed: %w", err)
	}

	return pool, nil
}

func Migrate(dsn string, fs fs.FS) error {
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return fmt.Errorf("database: could not create iofs for migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("database: could not initialize migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("database: migration failed: %w", err)
	}

	return nil
}
