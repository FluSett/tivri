package postgres

import (
	"context"
	"database/sql"
	"tivri/internal/domain/contact"
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

func (r *SQLRepository) Save(ctx context.Context, msg *contact.ContactMessage) error {
	var query string
	if r.driverName == "pgx" {
		query = "INSERT INTO contact_messages (email, topic, message, status) VALUES ($1, $2, $3, $4)"
	} else {
		query = "INSERT INTO contact_messages (email, topic, message, status) VALUES (?, ?, ?, ?)"
	}
	_, err := r.db.ExecContext(ctx, query, msg.Email, msg.Topic, msg.Message, msg.Status)
	return err
}

func (r *SQLRepository) List(ctx context.Context) ([]contact.ContactMessage, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, email, topic, message, status, created_at FROM contact_messages ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []contact.ContactMessage
	for rows.Next() {
		var m contact.ContactMessage
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
	var query string
	if r.driverName == "pgx" {
		query = "UPDATE contact_messages SET status = $1 WHERE id = $2"
	} else {
		query = "UPDATE contact_messages SET status = ? WHERE id = ?"
	}
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}
