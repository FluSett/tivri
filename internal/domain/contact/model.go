package contact

import (
	"context"
	"time"
)

type ContactMessage struct {
	ID        int64
	Email     string
	Topic     string
	Message   string
	Status    string
	CreatedAt time.Time
}

type Repository interface {
	Save(ctx context.Context, msg *ContactMessage) error
	List(ctx context.Context) ([]ContactMessage, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
}
