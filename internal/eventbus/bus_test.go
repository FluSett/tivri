package eventbus

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"
)

func TestMemoryEventBus(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	bus := NewMemoryEventBus(context.Background(), logger)

	var wg sync.WaitGroup

	wg.Add(1)

	var received Event

	bus.Subscribe("test.event", func(ctx context.Context, e Event) error {
		received = e

		wg.Done()

		return nil
	})

	payload := "hello world"

	bus.Publish(context.Background(), Event{
		Type:      "test.event",
		Payload:   payload,
		Timestamp: time.Now(),
	})

	wg.Wait()

	if received.Type != "test.event" {
		t.Errorf("expected type test.event, got %s", received.Type)
	}

	if received.Payload.(string) != payload {
		t.Errorf("expected payload %s, got %v", payload, received.Payload)
	}
}

func TestMemoryEventBus_HandlerError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	bus := NewMemoryEventBus(context.Background(), logger)

	var wg sync.WaitGroup

	wg.Add(1)

	bus.Subscribe("test.fail", func(ctx context.Context, e Event) error {
		defer wg.Done()

		return errors.New("expected failure")
	})

	bus.Publish(context.Background(), Event{
		Type:      "test.fail",
		Timestamp: time.Now(),
	})

	wg.Wait()
}
