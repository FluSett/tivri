package eventbus

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

const defaultHandlerTimeout = 5 * time.Second

type MemoryEventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]Handler
	logger      *slog.Logger
	ch          chan Event
}

func NewMemoryEventBus(ctx context.Context, logger *slog.Logger) *MemoryEventBus {
	bus := &MemoryEventBus{
		subscribers: make(map[string][]Handler),
		logger:      logger,
		ch:          make(chan Event, 100),
	}

	for i := 0; i < 5; i++ {
		go bus.worker(ctx)
	}

	return bus
}

func (b *MemoryEventBus) Subscribe(eventType string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[eventType] = append(b.subscribers[eventType], handler)
}

func (b *MemoryEventBus) Publish(ctx context.Context, e Event) {
	b.ch <- e
}

func (b *MemoryEventBus) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-b.ch:
			if !ok {
				return
			}

			b.mu.RLock()
			handlers, exists := b.subscribers[e.Type]
			b.mu.RUnlock()
			if !exists {
				continue
			}

			for _, h := range handlers {
				b.executeHandler(ctx, h, e)
			}
		}
	}
}

func (b *MemoryEventBus) executeHandler(ctx context.Context, handler Handler, ev Event) {
	defer func() {
		if r := recover(); r != nil {
			b.logger.Error("eventbus: handler panic recovered",
				slog.String("event_type", ev.Type),
				slog.Any("recover", r),
			)
		}
	}()

	hCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), defaultHandlerTimeout)
	defer cancel()

	err := handler(hCtx, ev)
	if err != nil {
		b.logger.Error("eventbus: handler returned error",
			slog.String("event_type", ev.Type),
			slog.Any("error", err),
		)
	}
}
