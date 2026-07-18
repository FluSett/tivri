package database

import (
	"context"
	"testing"
)

func TestConnect_InvalidDSN(t *testing.T) {
	ctx := context.Background()

	_, err := Connect(ctx, "invalid connection string format", 10, 10)
	if err == nil {
		t.Error("expected connection parsing error for invalid DSN, got nil")
	}

	_, err = Connect(ctx, "postgres://postgres:wrong_password@localhost:2345/non_existent_db?sslmode=disable", 10, 10)
	if err == nil {
		t.Error("expected connection ping error for unreachable server, got nil")
	}
}
