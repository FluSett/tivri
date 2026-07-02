package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"tivri/internal/domain/portfolio"
)

type SQLRepository struct {
	db         *sql.DB
	driverName string
}

func NewSQLRepository(db *sql.DB, driverName string) *SQLRepository {
	return &SQLRepository{
		db:         db,
		driverName: driverName,
	}
}

func (r *SQLRepository) Save(ctx context.Context, item *portfolio.PortfolioItem) error {
	mediaBytes, err := json.Marshal(item.Media)
	if err != nil {
		return err
	}
	mediaStr := string(mediaBytes)
	if r.driverName == "pgx" {
		query := "INSERT INTO portfolio_items (title, description, budget, tech_stack, media) VALUES ($1, $2, $3, $4, $5) RETURNING id"
		return r.db.QueryRowContext(ctx, query, item.Title, item.Description, item.Budget, item.TechStack, mediaStr).Scan(&item.ID)
	}
	query := "INSERT INTO portfolio_items (title, description, budget, tech_stack, media) VALUES (?, ?, ?, ?, ?)"
	res, err := r.db.ExecContext(ctx, query, item.Title, item.Description, item.Budget, item.TechStack, mediaStr)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	item.ID = id
	return nil
}

func (r *SQLRepository) List(ctx context.Context) ([]portfolio.PortfolioItem, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, title, description, budget, tech_stack, media FROM portfolio_items ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []portfolio.PortfolioItem
	for rows.Next() {
		var item portfolio.PortfolioItem
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
