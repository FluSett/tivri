package notifications

import (
	"context"
	"testing"
	"time"

	"tivri/internal/eventbus"
)

func TestEmailWorker_HandleEvent(t *testing.T) {
	worker := NewEmailWorker()
	event := eventbus.Event{
		Type:      "contact.created",
		Timestamp: time.Now(),
	}

	err := worker.HandleEvent(context.Background(), event)
	if err != nil {
		t.Errorf("expected no error from EmailWorker, got %v", err)
	}
}
