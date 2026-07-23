package datastore

import (
	"context"
	"fmt"
	"tivri/internal/core"
)

type ContactRepo struct {
	store *Store
}

func NewContactRepo(store *Store) core.ContactRepository {
	return &ContactRepo{store: store}
}

func (r *ContactRepo) Save(ctx context.Context, msg *core.ContactMessage) error {
	query := "INSERT INTO contact_messages (email, topic, message, status) VALUES ($1, $2, $3, $4) RETURNING id"
	if err := r.store.QueryRow(ctx, query, msg.Email, msg.Topic, msg.Message, msg.Status).Scan(&msg.ID); err != nil {
		return fmt.Errorf("messaging: save failed: %w", err)
	}
	return nil
}

func (r *ContactRepo) List(ctx context.Context, params core.MessageListParams) (core.PaginatedMessages, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	whereClause := "WHERE 1=1"
	var args []interface{}
	argID := 1

	if params.Status != "" && params.Status != "all" {
		whereClause += fmt.Sprintf(" AND status = $%d", argID)
		args = append(args, params.Status)
		argID++
	}

	if params.SearchQuery != "" {
		whereClause += fmt.Sprintf(" AND (email ILIKE $%d OR topic ILIKE $%d)", argID, argID+1)
		args = append(args, "%"+params.SearchQuery+"%", "%"+params.SearchQuery+"%")
		argID += 2
	}

	orderBy := "ORDER BY created_at DESC"
	switch params.SortBy {
	case "date_asc":
		orderBy = "ORDER BY created_at ASC"
	case "date_desc":
		orderBy = "ORDER BY created_at DESC"
	case "email_asc":
		orderBy = "ORDER BY email ASC"
	case "email_desc":
		orderBy = "ORDER BY email DESC"
	}

	offset := (params.Page - 1) * params.PageSize
	query := fmt.Sprintf("SELECT id, email, topic, message, status, created_at, updated_at, COUNT(*) OVER() AS full_count FROM contact_messages %s %s LIMIT $%d OFFSET $%d", whereClause, orderBy, argID, argID+1)
	args = append(args, params.PageSize, offset)

	var list []core.ContactMessage
	var totalItems int

	err := r.store.WithTx(ctx, func(txCtx context.Context) error {
		rows, err := r.store.Query(txCtx, query, args...)
		if err != nil {
			return fmt.Errorf("messaging: query failed: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var m core.ContactMessage
			var fullCount int
			if err := rows.Scan(&m.ID, &m.Email, &m.Topic, &m.Message, &m.Status, &m.CreatedAt, &m.UpdatedAt, &fullCount); err != nil {
				return fmt.Errorf("messaging: scan failed: %w", err)
			}
			totalItems = fullCount
			list = append(list, m)
		}

		return rows.Err()
	})
	if err != nil {
		return core.PaginatedMessages{}, err
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages = totalItems / params.PageSize
		if totalItems%params.PageSize > 0 {
			totalPages++
		}
	}

	return core.PaginatedMessages{
		Items:      list,
		TotalItems: totalItems,
		TotalPages: totalPages,
		Page:       params.Page,
		PageSize:   params.PageSize,
		Params:     params,
	}, nil
}

func (r *ContactRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := "UPDATE contact_messages SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2"
	if err := r.store.Exec(ctx, query, status, id); err != nil {
		return fmt.Errorf("messaging: update status failed: %w", err)
	}
	return nil
}
