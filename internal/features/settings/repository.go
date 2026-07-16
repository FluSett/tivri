package settings

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetHighQueue(ctx context.Context) (bool, error)
	SetHighQueue(ctx context.Context, enabled bool) error
	GetMaintenance(ctx context.Context) (bool, error)
	SetMaintenance(ctx context.Context, enabled bool) error
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetHighQueue(ctx context.Context) (bool, error) {
	var val string
	err := r.db.QueryRow(ctx, "SELECT value FROM system_settings WHERE key = $1", "high_queue").Scan(&val)
	if err != nil {
		return false, fmt.Errorf("settings: get high_queue failed: %w", err)
	}
	return val == "true", nil
}

func (r *PostgresRepository) SetHighQueue(ctx context.Context, enabled bool) error {
	val := "false"
	if enabled {
		val = "true"
	}
	_, err := r.db.Exec(ctx, "INSERT INTO system_settings (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value", "high_queue", val)
	if err != nil {
		return fmt.Errorf("settings: set high_queue failed: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetMaintenance(ctx context.Context) (bool, error) {
	var val string
	err := r.db.QueryRow(ctx, "SELECT value FROM system_settings WHERE key = $1", "maintenance_mode").Scan(&val)
	if err != nil {
		return false, fmt.Errorf("settings: get maintenance failed: %w", err)
	}
	return val == "true", nil
}

func (r *PostgresRepository) SetMaintenance(ctx context.Context, enabled bool) error {
	val := "false"
	if enabled {
		val = "true"
	}
	_, err := r.db.Exec(ctx, "INSERT INTO system_settings (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value", "maintenance_mode", val)
	if err != nil {
		return fmt.Errorf("settings: set maintenance failed: %w", err)
	}
	return nil
}
