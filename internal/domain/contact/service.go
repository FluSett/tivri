package contact

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateMessage(ctx context.Context, email, topic, message string) (*ContactMessage, error) {
	email = strings.TrimSpace(email)
	topic = strings.TrimSpace(topic)
	message = strings.TrimSpace(message)

	if len(email) < 5 || !strings.Contains(email, "@") {
		return nil, errors.New("invalid email")
	}

	if len(topic) < 3 {
		return nil, errors.New("topic too short")
	}

	if len(message) < 10 {
		return nil, errors.New("message too short")
	}

	msg := &ContactMessage{
		Email:     email,
		Topic:     topic,
		Message:   message,
		Status:    "new",
		CreatedAt: time.Now(),
	}

	if err := s.repo.Save(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *Service) ListMessages(ctx context.Context) ([]ContactMessage, error) {
	return s.repo.List(ctx)
}

func (s *Service) UpdateMessageStatus(ctx context.Context, id int64, status string) error {
	return s.repo.UpdateStatus(ctx, id, status)
}
