package eventbus

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultPollInterval = 2 * time.Second

type OutboxWorker struct {
	db     *pgxpool.Pool
	bus    Bus
	logger *slog.Logger
}

func NewOutboxWorker(db *pgxpool.Pool, bus Bus, logger *slog.Logger) *OutboxWorker {
	return &OutboxWorker{
		db:     db,
		bus:    bus,
		logger: logger,
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(defaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processPending(ctx)
		}
	}
}

func (w *OutboxWorker) processPending(ctx context.Context) {
	query := "SELECT id, type, payload, created_at FROM outbox_events WHERE processed = false ORDER BY id ASC LIMIT 50"
	rows, err := w.db.Query(ctx, query)
	if err != nil {
		w.logger.Error("outbox: query failed", slog.Any("error", err))
		return
	}
	defer rows.Close()

	var events []struct {
		id        int
		evtType   string
		payload   []byte
		createdAt time.Time
	}

	for rows.Next() {
		var e struct {
			id        int
			evtType   string
			payload   []byte
			createdAt time.Time
		}
		if err := rows.Scan(&e.id, &e.evtType, &e.payload, &e.createdAt); err != nil {
			w.logger.Error("outbox: scan failed", slog.Any("error", err))
			continue
		}
		events = append(events, e)
	}
	rows.Close()

	if len(events) == 0 {
		return
	}

	for _, e := range events {
		w.bus.Publish(ctx, Event{
			Type:      e.evtType,
			Payload:   e.payload,
			Timestamp: e.createdAt,
		})

		_, err = w.db.Exec(ctx, "UPDATE outbox_events SET processed = true WHERE id = $1", e.id)
		if err != nil {
			w.logger.Error("outbox: update processed failed", slog.Any("error", err))
		}
	}
}
