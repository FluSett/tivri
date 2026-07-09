package messaging

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Save(ctx context.Context, msg *ContactMessage) error {
	query := "INSERT INTO contact_messages (email, topic, message, status) VALUES ($1, $2, $3, $4) RETURNING id"
	err := r.pool.QueryRow(ctx, query, msg.Email, msg.Topic, msg.Message, msg.Status).Scan(&msg.ID)
	if err != nil {
		return fmt.Errorf("messaging: save failed: %w", err)
	}

	return nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]ContactMessage, error) {
	query := "SELECT id, email, topic, message, status, created_at, updated_at FROM contact_messages ORDER BY id DESC"
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("messaging: query failed: %w", err)
	}
	defer rows.Close()

	var list []ContactMessage
	for rows.Next() {
		var m ContactMessage
		err := rows.Scan(&m.ID, &m.Email, &m.Topic, &m.Message, &m.Status, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("messaging: scan failed: %w", err)
		}

		list = append(list, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("messaging: row iteration failed: %w", err)
	}

	return list, nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := "UPDATE contact_messages SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2"
	_, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("messaging: update status failed: %w", err)
	}

	return nil
}
