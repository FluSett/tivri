package portfolio

import (
	"context"
)

type PortfolioItem struct {
	ID          int64
	Title       string
	Description string
	Budget      int64
	TechStack   string
	Media       []string
}

type Repository interface {
	Save(ctx context.Context, item *PortfolioItem) error
	List(ctx context.Context) ([]PortfolioItem, error)
}
