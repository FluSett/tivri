package notifications

import (
	"context"
	"fmt"

	"tivri/internal/eventbus"
)

type EmailWorker struct{}

func NewEmailWorker() *EmailWorker {
	return &EmailWorker{}
}

func (w *EmailWorker) HandleEvent(ctx context.Context, e eventbus.Event) error {
	fmt.Printf("Notifications Worker: Email dispatched for event type %q\n", e.Type)

	return nil
}
