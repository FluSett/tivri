package portfolio

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Save(ctx context.Context, item *PortfolioItem) error {
	mediaBytes, err := json.Marshal(item.Media)
	if err != nil {
		return fmt.Errorf("portfolio: marshal media failed: %w", err)
	}

	mediaStr := string(mediaBytes)
	query := "INSERT INTO portfolio_items (title, description, budget, tech_stack, media) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	err = r.pool.QueryRow(ctx, query, item.Title, item.Description, item.Budget, item.TechStack, mediaStr).Scan(&item.ID)
	if err != nil {
		return fmt.Errorf("portfolio: save item failed: %w", err)
	}

	return nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]PortfolioItem, error) {
	query := "SELECT id, title, description, budget, tech_stack, media FROM portfolio_items ORDER BY id DESC"
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("portfolio: query items failed: %w", err)
	}
	defer rows.Close()

	var list []PortfolioItem
	for rows.Next() {
		var item PortfolioItem
		var mediaJSON string
		err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.Budget, &item.TechStack, &mediaJSON)
		if err != nil {
			return nil, fmt.Errorf("portfolio: scan item failed: %w", err)
		}

		var media []string
		if mediaJSON != "" {
			err := json.Unmarshal([]byte(mediaJSON), &media)
			if err != nil {
				return nil, fmt.Errorf("portfolio: unmarshal media failed: %w", err)
			}
		}

		item.Media = media
		list = append(list, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("portfolio: row iteration failed: %w", err)
	}

	return list, nil
}
