package eventbus

import (
	"context"
	"time"
)

type Event struct {
	Type      string
	Payload   any
	Timestamp time.Time
}

type Handler func(ctx context.Context, e Event) error

type Bus interface {
	Publish(ctx context.Context, e Event)
	Subscribe(eventType string, handler Handler)
}
