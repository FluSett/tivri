package datastore

import (
	"context"
	"encoding/json"
	"fmt"
	"tivri/internal/core"
)

type PortfolioRepo struct {
	store *Store
}

func NewPortfolioRepo(store *Store) core.PortfolioRepository {
	return &PortfolioRepo{store: store}
}

func (r *PortfolioRepo) Save(ctx context.Context, item *core.PortfolioItem) error {
	mediaBytes, err := json.Marshal(item.Media)
	if err != nil {
		return fmt.Errorf("portfolio: marshal media failed: %w", err)
	}

	query := "INSERT INTO portfolio_items (title, description, budget, tech_stack, media) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err = r.store.QueryRow(ctx, query, item.Title, item.Description, item.Budget, item.TechStack, string(mediaBytes)).Scan(&item.ID)
	if err != nil {
		return fmt.Errorf("portfolio: save item failed: %w", err)
	}

	return nil
}

func (r *PortfolioRepo) List(ctx context.Context) ([]core.PortfolioItem, error) {
	query := "SELECT id, title, description, budget, tech_stack, media FROM portfolio_items ORDER BY id DESC"
	rows, err := r.store.Pool().Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("portfolio: query items failed: %w", err)
	}
	defer rows.Close()

	var list []core.PortfolioItem
	for rows.Next() {
		var item core.PortfolioItem
		var mediaJSON string
		if err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.Budget, &item.TechStack, &mediaJSON); err != nil {
			return nil, fmt.Errorf("portfolio: scan item failed: %w", err)
		}
		if mediaJSON != "" {
			if err := json.Unmarshal([]byte(mediaJSON), &item.Media); err != nil {
				return nil, fmt.Errorf("portfolio: unmarshal media failed: %w", err)
			}
		}
		list = append(list, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("portfolio: row iteration failed: %w", err)
	}

	return list, nil
}
