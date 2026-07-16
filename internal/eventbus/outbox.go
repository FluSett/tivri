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

	cleanupTicker := time.NewTicker(24 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processPending(ctx)
		case <-cleanupTicker.C:
			w.cleanupProcessed(ctx)
		}
	}
}

func (w *OutboxWorker) cleanupProcessed(ctx context.Context) {
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	res, err := w.db.Exec(ctx, "DELETE FROM outbox_events WHERE processed = true AND created_at < $1", sevenDaysAgo)
	if err != nil {
		w.logger.Error("outbox: cleanup failed", slog.Any("error", err))
		return
	}
	if res.RowsAffected() > 0 {
		w.logger.Info("outbox: cleanup removed old events", slog.Int64("count", res.RowsAffected()))
	}
}

func (w *OutboxWorker) processPending(ctx context.Context) {
	tx, err := w.db.Begin(ctx)
	if err != nil {
		w.logger.Error("outbox: begin transaction failed", slog.Any("error", err))
		return
	}
	defer tx.Rollback(ctx)

	query := "SELECT id, type, payload, created_at FROM outbox_events WHERE processed = false ORDER BY id ASC LIMIT 50 FOR UPDATE SKIP LOCKED"
	rows, err := tx.Query(ctx, query)
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
		_, err = tx.Exec(ctx, "UPDATE outbox_events SET processed = true WHERE id = $1", e.id)
		if err != nil {
			w.logger.Error("outbox: update processed failed", slog.Any("error", err))
			return
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		w.logger.Error("outbox: commit transaction failed", slog.Any("error", err))
		return
	}

	for _, e := range events {
		w.bus.Publish(ctx, Event{
			Type:      e.evtType,
			Payload:   e.payload,
			Timestamp: e.createdAt,
		})
	}
}
