package portfolio

import (
	"context"
	"database/sql"
	"encoding/json"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Save(ctx context.Context, item *PortfolioItem) error {
	mediaBytes, err := json.Marshal(item.Media)
	if err != nil {
		return err
	}
	mediaStr := string(mediaBytes)
	query := "INSERT INTO portfolio_items (title, description, budget, tech_stack, media) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	return r.db.QueryRowContext(ctx, query, item.Title, item.Description, item.Budget, item.TechStack, mediaStr).Scan(&item.ID)
}

func (r *SQLRepository) List(ctx context.Context) ([]PortfolioItem, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, title, description, budget, tech_stack, media FROM portfolio_items ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []PortfolioItem
	for rows.Next() {
		var item PortfolioItem
		var mediaJSON string
		if err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.Budget, &item.TechStack, &mediaJSON); err != nil {
			return nil, err
		}
		var media []string
		if mediaJSON != "" {
			if err := json.Unmarshal([]byte(mediaJSON), &media); err != nil {
				return nil, err
			}
		}
		item.Media = media
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}
