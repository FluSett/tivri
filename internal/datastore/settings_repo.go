package datastore

import (
	"context"
	"fmt"
	"tivri/internal/core"
)

type SettingsRepo struct {
	store *Store
}

func NewSettingsRepo(store *Store) core.SettingsRepository {
	return &SettingsRepo{store: store}
}

func (r *SettingsRepo) GetHighQueue(ctx context.Context) (bool, error) {
	var val string
	err := r.store.QueryRow(ctx, "SELECT value FROM system_settings WHERE key = $1", "high_queue").Scan(&val)
	if err != nil {
		return false, fmt.Errorf("settings: get high_queue failed: %w", err)
	}
	return val == "true", nil
}

func (r *SettingsRepo) SetHighQueue(ctx context.Context, enabled bool) error {
	val := "false"
	if enabled {
		val = "true"
	}
	err := r.store.Exec(ctx, "INSERT INTO system_settings (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value", "high_queue", val)
	if err != nil {
		return fmt.Errorf("settings: set high_queue failed: %w", err)
	}
	return nil
}

func (r *SettingsRepo) GetMaintenance(ctx context.Context) (bool, error) {
	var val string
	err := r.store.QueryRow(ctx, "SELECT value FROM system_settings WHERE key = $1", "maintenance_mode").Scan(&val)
	if err != nil {
		return false, fmt.Errorf("settings: get maintenance failed: %w", err)
	}
	return val == "true", nil
}

func (r *SettingsRepo) SetMaintenance(ctx context.Context, enabled bool) error {
	val := "false"
	if enabled {
		val = "true"
	}
	err := r.store.Exec(ctx, "INSERT INTO system_settings (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value", "maintenance_mode", val)
	if err != nil {
		return fmt.Errorf("settings: set maintenance failed: %w", err)
	}
	return nil
}
