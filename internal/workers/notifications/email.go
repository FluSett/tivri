package notifications

import (
	"context"
	"log/slog"

	"tivri/internal/eventbus"
)

type EmailWorker struct{}

func NewEmailWorker() *EmailWorker {
	return &EmailWorker{}
}

func (w *EmailWorker) HandleEvent(ctx context.Context, e eventbus.Event) error {
	slog.Info("Notifications Worker: Email dispatched", "event_type", e.Type)

	return nil
}
