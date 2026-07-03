package contact

import (
	"context"
	"database/sql"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Save(ctx context.Context, msg *ContactMessage) error {
	query := "INSERT INTO contact_messages (email, topic, message, status) VALUES ($1, $2, $3, $4)"
	_, err := r.db.ExecContext(ctx, query, msg.Email, msg.Topic, msg.Message, msg.Status)

	return err
}

func (r *SQLRepository) List(ctx context.Context) ([]ContactMessage, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, email, topic, message, status, created_at FROM contact_messages ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ContactMessage

	for rows.Next() {
		var m ContactMessage
		if err := rows.Scan(&m.ID, &m.Email, &m.Topic, &m.Message, &m.Status, &m.CreatedAt); err != nil {
			return nil, err
		}

		list = append(list, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *SQLRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := "UPDATE contact_messages SET status = $1 WHERE id = $2"
	_, err := r.db.ExecContext(ctx, query, status, id)

	return err
}
