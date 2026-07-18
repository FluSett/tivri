package services

import (
	"context"
	"time"
	"tivri/internal/core"
	"tivri/internal/eventbus"
)

type PortfolioService struct {
	repo     core.PortfolioRepository
	eventBus eventbus.Bus
}

func NewPortfolioService(repo core.PortfolioRepository, eventBus eventbus.Bus) *PortfolioService {
	return &PortfolioService{repo: repo, eventBus: eventBus}
}

func (s *PortfolioService) SaveItem(ctx context.Context, item *core.PortfolioItem) error {
	if err := s.repo.Save(ctx, item); err != nil {
		return err
	}

	s.eventBus.Publish(ctx, eventbus.Event{
		Type:      "portfolio.created",
		Payload:   item.ID,
		Timestamp: time.Now(),
	})

	return nil
}
