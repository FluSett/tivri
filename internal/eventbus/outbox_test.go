package eventbus

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"tivri/internal/core/database"
)

type mockBus struct {
	published []Event
}

func (m *mockBus) Subscribe(eventType string, handler Handler) {}
func (m *mockBus) Publish(ctx context.Context, e Event) {
	m.published = append(m.published, e)
}

func TestOutboxWorker(t *testing.T) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		t.Skip("Skipping outbox worker integration test: DB_DSN not set")
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, dsn, 10, 10)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	_, _ = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS outbox_events (
			id SERIAL PRIMARY KEY,
			type TEXT NOT NULL,
			payload JSONB NOT NULL,
			processed BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	_, _ = pool.Exec(ctx, "DELETE FROM outbox_events")

	_, err = pool.Exec(ctx, "INSERT INTO outbox_events (type, payload) VALUES ($1, $2)", "test.outbox.event", `{"data":"test"}`)
	if err != nil {
		t.Fatalf("failed to insert test outbox event: %v", err)
	}

	bus := &mockBus{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	worker := NewOutboxWorker(pool, bus, logger)

	worker.processPending(ctx)

	if len(bus.published) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(bus.published))
	}

	if bus.published[0].Type != "test.outbox.event" {
		t.Errorf("expected type test.outbox.event, got %s", bus.published[0].Type)
	}

	var processed bool
	err = pool.QueryRow(ctx, "SELECT processed FROM outbox_events LIMIT 1").Scan(&processed)
	if err != nil {
		t.Fatalf("failed to scan processed status: %v", err)
	}
	if !processed {
		t.Error("expected outbox event to be marked as processed")
	}
}
