package messaging

import (
	"context"
	"time"
)

type ContactMessage struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Topic     string    `json:"topic"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type Repository interface {
	Save(ctx context.Context, msg *ContactMessage) error

	List(ctx context.Context) ([]ContactMessage, error)

	UpdateStatus(ctx context.Context, id int64, status string) error
}
