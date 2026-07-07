package portfolio

import (
	"context"
)

type PortfolioItem struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Budget      int64    `json:"budget"`
	TechStack   string   `json:"techStack"`
	Media       []string `json:"media"`
}

type Repository interface {
	Save(ctx context.Context, item *PortfolioItem) error

	List(ctx context.Context) ([]PortfolioItem, error)
}
