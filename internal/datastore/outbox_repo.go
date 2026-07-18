package datastore

import (
	"context"
	"fmt"
	"tivri/internal/core"
)

type OutboxRepo struct {
	store *Store
}

func NewOutboxRepo(store *Store) core.OutboxRepository {
	return &OutboxRepo{store: store}
}

func (r *OutboxRepo) Save(ctx context.Context, evt *core.OutboxEvent) error {
	err := r.store.Exec(ctx, "INSERT INTO outbox_events (type, payload) VALUES ($1, $2)", evt.Type, evt.Payload)
	if err != nil {
		return fmt.Errorf("outbox: save failed: %w", err)
	}
	return nil
}
