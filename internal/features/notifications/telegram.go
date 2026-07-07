package notifications

import (
	"context"
	"fmt"

	"tivri/internal/eventbus"
)

type TelegramWorker struct{}

func NewTelegramWorker() *TelegramWorker {
	return &TelegramWorker{}
}

func (w *TelegramWorker) HandleEvent(ctx context.Context, e eventbus.Event) error {
	fmt.Printf("Notifications Worker: Telegram message dispatched for event type %q\n", e.Type)

	return nil
}
