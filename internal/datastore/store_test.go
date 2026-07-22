package datastore

import (
	"context"
	"os"
	"testing"

	"tivri/internal/core/database"
)

func TestStore_Connect(t *testing.T) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		t.Skip("Skipping datastore integration test: DB_DSN not set")
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, dsn, 5, 2)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	store := NewStore(pool)
	if store.Pool() == nil {
		t.Fatal("expected non-nil pool from store")
	}
}
