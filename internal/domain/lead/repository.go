package lead

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

func (r *SQLRepository) Save(ctx context.Context, ld *Lead) error {
	query := "INSERT INTO intake_leads (company_name, project_scope, budget, contact_email, contact_phone, client_status, internal_status) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	_, err := r.db.ExecContext(ctx, query, ld.CompanyName, ld.ProjectScope, ld.Budget, ld.ContactEmail, ld.ContactPhone, ld.ClientStatus, ld.InternalStatus)

	return err
}

func (r *SQLRepository) List(ctx context.Context) ([]Lead, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, company_name, project_scope, budget, contact_email, contact_phone, client_status, internal_status, created_at, updated_at FROM intake_leads ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Lead

	for rows.Next() {
		var ld Lead
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
	query := "UPDATE intake_leads SET client_status = $1, internal_status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3"
	_, err := r.db.ExecContext(ctx, query, clientStatus, internalStatus, id)

	return err
}
