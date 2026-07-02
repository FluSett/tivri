package postgres

import (
	"context"
	"database/sql"
	"tivri/internal/domain/lead"
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

func (r *SQLRepository) Save(ctx context.Context, ld *lead.Lead) error {
	var query string
	if r.driverName == "pgx" {
		query = "INSERT INTO intake_leads (company_name, project_scope, budget, contact_email, contact_phone, client_status, internal_status) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	} else {
		query = "INSERT INTO intake_leads (company_name, project_scope, budget, contact_email, contact_phone, client_status, internal_status) VALUES (?, ?, ?, ?, ?, ?, ?)"
	}
	_, err := r.db.ExecContext(ctx, query, ld.CompanyName, ld.ProjectScope, ld.Budget, ld.ContactEmail, ld.ContactPhone, ld.ClientStatus, ld.InternalStatus)
	return err
}

func (r *SQLRepository) List(ctx context.Context) ([]lead.Lead, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, company_name, project_scope, budget, contact_email, contact_phone, client_status, internal_status, created_at, updated_at FROM intake_leads ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []lead.Lead
	for rows.Next() {
		var ld lead.Lead
		if err := rows.Scan(&ld.ID, &ld.CompanyName, &ld.ProjectScope, &ld.Budget, &ld.ContactEmail, &ld.ContactPhone, &ld.ClientStatus, &ld.InternalStatus, &ld.CreatedAt, &ld.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, ld)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *SQLRepository) UpdateStatus(ctx context.Context, id int64, clientStatus, internalStatus string) error {
	var query string
	if r.driverName == "pgx" {
		query = "UPDATE intake_leads SET client_status = $1, internal_status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3"
	} else {
		query = "UPDATE intake_leads SET client_status = ?, internal_status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	}
	_, err := r.db.ExecContext(ctx, query, clientStatus, internalStatus, id)
	return err
}
