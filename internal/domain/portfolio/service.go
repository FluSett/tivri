package portfolio

import (
	"context"
	"errors"
	"strings"
	"sync"
)

type Service struct {
	repo        Repository
	mu          sync.RWMutex
	cache       []PortfolioItem
	initialized bool
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

	s.mu.Lock()
	if s.initialized {
		s.cache = append([]PortfolioItem{*item}, s.cache...)
	}
	s.mu.Unlock()

	return item, nil
}

func (s *Service) ListPortfolioItems(ctx context.Context) ([]PortfolioItem, error) {
	s.mu.RLock()
	if s.initialized {
		items := make([]PortfolioItem, len(s.cache))
		copy(items, s.cache)
		s.mu.RUnlock()
		return items, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.initialized {
		items := make([]PortfolioItem, len(s.cache))
		copy(items, s.cache)
		return items, nil
	}

	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	s.cache = items
	s.initialized = true

	copiedItems := make([]PortfolioItem, len(items))
	copy(copiedItems, items)
	return copiedItems, nil
}
