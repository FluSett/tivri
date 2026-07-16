package database

import (
	"context"
	"testing"
)

func TestConnect_InvalidDSN(t *testing.T) {
	ctx := context.Background()

	// Parse failure testing
	_, err := Connect(ctx, "invalid connection string format")
	if err == nil {
		t.Error("expected connection parsing error for invalid DSN, got nil")
	}

	// Ping failure testing on valid URL structure but unreachable server
	_, err = Connect(ctx, "postgres://postgres:wrong_password@localhost:2345/non_existent_db?sslmode=disable")
	if err == nil {
		t.Error("expected connection ping error for unreachable server, got nil")
	}
}
