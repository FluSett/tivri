package core

import "context"

type OutboxEvent struct {
	Type    string
	Payload []byte
}

type OutboxRepository interface {
	Save(ctx context.Context, evt *OutboxEvent) error
}
