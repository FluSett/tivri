package settings

import (
	"context"
	"os"
	"testing"

	"tivri/internal/core/database"
)

func TestRepository(t *testing.T) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		t.Skip("Skipping settings repository integration test: DB_DSN not set")
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Ensure clean slate
	_, _ = pool.Exec(ctx, "CREATE TABLE IF NOT EXISTS system_settings (key TEXT PRIMARY KEY, value TEXT)")
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM system_settings WHERE key IN ('high_queue', 'maintenance_mode')")
	}()

	repo := NewRepository(pool)

	// Set High Queue
	err = repo.SetHighQueue(ctx, true)
	if err != nil {
		t.Fatalf("failed to set high queue: %v", err)
	}

	val, err := repo.GetHighQueue(ctx)
	if err != nil {
		t.Fatalf("failed to get high queue: %v", err)
	}
	if !val {
		t.Errorf("expected high queue to be true")
	}

	// Disable High Queue
	err = repo.SetHighQueue(ctx, false)
	if err != nil {
		t.Fatalf("failed to disable high queue: %v", err)
	}

	val, err = repo.GetHighQueue(ctx)
	if err != nil {
		t.Fatalf("failed to get disabled high queue: %v", err)
	}
	if val {
		t.Errorf("expected high queue to be false")
	}

	// Set Maintenance
	err = repo.SetMaintenance(ctx, true)
	if err != nil {
		t.Fatalf("failed to set maintenance: %v", err)
	}

	val, err = repo.GetMaintenance(ctx)
	if err != nil {
		t.Fatalf("failed to get maintenance: %v", err)
	}
	if !val {
		t.Errorf("expected maintenance to be true")
	}
}
