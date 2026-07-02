package portfolio

import (
	"context"
	"errors"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreatePortfolioItem(ctx context.Context, title, description string, budget int64, techStack string, media []string) (*PortfolioItem, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	techStack = strings.TrimSpace(techStack)

	if len(title) == 0 {
		return nil, errors.New("title cannot be empty")
	}

	if len(description) == 0 {
		return nil, errors.New("description cannot be empty")
	}

	item := &PortfolioItem{
		Title:       title,
		Description: description,
		Budget:      budget,
		TechStack:   techStack,
		Media:       media,
	}

	if err := s.repo.Save(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *Service) ListPortfolioItems(ctx context.Context) ([]PortfolioItem, error) {
	return s.repo.List(ctx)
}
