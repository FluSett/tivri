package datastore

import (
	"context"
	"fmt"
	"tivri/internal/core"
)

type IntakeRepo struct {
	store *Store
}

func NewIntakeRepo(store *Store) core.LeadRepository {
	return &IntakeRepo{store: store}
}

func (r *IntakeRepo) Save(ctx context.Context, ld *core.Lead) error {
	query := "INSERT INTO intake_leads (company_name, service_type, project_scope, existing_url, tech_stack, budget, contact_email, contact_info, deadline_needed, deadline_spec, client_status, internal_status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id"
	err := r.store.QueryRow(ctx, query, ld.CompanyName, ld.ServiceType, ld.ProjectScope, ld.ExistingURL, ld.TechStack, ld.Budget, ld.ContactEmail, ld.ContactInfo, ld.DeadlineNeeded, ld.DeadlineSpec, ld.ClientStatus, ld.InternalStatus).Scan(&ld.ID)
	if err != nil {
		return fmt.Errorf("project_intake: save failed: %w", err)
	}
	return nil
}

func (r *IntakeRepo) Get(ctx context.Context, id int64) (*core.Lead, error) {
	query := "SELECT id, company_name, service_type, project_scope, existing_url, tech_stack, budget, contact_email, contact_info, deadline_needed, deadline_spec, client_status, internal_status, created_at, updated_at FROM intake_leads WHERE id = $1"
	var ld core.Lead
	err := r.store.QueryRow(ctx, query, id).Scan(&ld.ID, &ld.CompanyName, &ld.ServiceType, &ld.ProjectScope, &ld.ExistingURL, &ld.TechStack, &ld.Budget, &ld.ContactEmail, &ld.ContactInfo, &ld.DeadlineNeeded, &ld.DeadlineSpec, &ld.ClientStatus, &ld.InternalStatus, &ld.CreatedAt, &ld.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("project_intake: get failed: %w", err)
	}
	return &ld, nil
}

func (r *IntakeRepo) List(ctx context.Context, params core.LeadListParams) (core.PaginatedLeads, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	whereClause := "WHERE 1=1"
	var args []interface{}
	argID := 1

	if params.ClientStatus != "" && params.ClientStatus != "all" {
		whereClause += fmt.Sprintf(" AND client_status = $%d", argID)
		args = append(args, params.ClientStatus)
		argID++
	}

	if params.InternalStatus != "" && params.InternalStatus != "all" {
		whereClause += fmt.Sprintf(" AND internal_status = $%d", argID)
		args = append(args, params.InternalStatus)
		argID++
	}

	if params.ServiceType != "" && params.ServiceType != "all" {
		whereClause += fmt.Sprintf(" AND service_type = $%d", argID)
		args = append(args, params.ServiceType)
		argID++
	}

	if params.SearchQuery != "" {
		whereClause += fmt.Sprintf(" AND (company_name ILIKE $%d OR contact_email ILIKE $%d OR project_scope ILIKE $%d)", argID, argID+1, argID+2)
		args = append(args, "%"+params.SearchQuery+"%", "%"+params.SearchQuery+"%", "%"+params.SearchQuery+"%")
		argID += 3
	}

	orderBy := "ORDER BY created_at DESC"
	switch params.SortBy {
	case "date_asc":
		orderBy = "ORDER BY created_at ASC"
	case "date_desc":
		orderBy = "ORDER BY created_at DESC"
	case "budget_desc":
		orderBy = "ORDER BY budget DESC"
	case "budget_asc":
		orderBy = "ORDER BY budget ASC"
	case "name_asc":
		orderBy = "ORDER BY company_name ASC"
	case "name_desc":
		orderBy = "ORDER BY company_name DESC"
	}

	offset := (params.Page - 1) * params.PageSize
	query := fmt.Sprintf("SELECT id, company_name, service_type, project_scope, existing_url, tech_stack, budget, contact_email, contact_info, deadline_needed, deadline_spec, client_status, internal_status, created_at, updated_at, COUNT(*) OVER() AS full_count FROM intake_leads %s %s LIMIT $%d OFFSET $%d", whereClause, orderBy, argID, argID+1)
	args = append(args, params.PageSize, offset)

	var list []core.Lead
	var totalItems int

	err := r.store.WithTx(ctx, func(txCtx context.Context) error {
		rows, err := r.store.Query(txCtx, query, args...)
		if err != nil {
			return fmt.Errorf("project_intake: query leads failed: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var ld core.Lead
			var fullCount int
			if err := rows.Scan(&ld.ID, &ld.CompanyName, &ld.ServiceType, &ld.ProjectScope, &ld.ExistingURL, &ld.TechStack, &ld.Budget, &ld.ContactEmail, &ld.ContactInfo, &ld.DeadlineNeeded, &ld.DeadlineSpec, &ld.ClientStatus, &ld.InternalStatus, &ld.CreatedAt, &ld.UpdatedAt, &fullCount); err != nil {
				return fmt.Errorf("project_intake: scan lead failed: %w", err)
			}
			totalItems = fullCount
			list = append(list, ld)
		}

		return rows.Err()
	})
	if err != nil {
		return core.PaginatedLeads{}, err
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages = totalItems / params.PageSize
		if totalItems%params.PageSize > 0 {
			totalPages++
		}
	}

	return core.PaginatedLeads{
		Items:      list,
		TotalItems: totalItems,
		TotalPages: totalPages,
		Page:       params.Page,
		PageSize:   params.PageSize,
		Params:     params,
	}, nil
}

func (r *IntakeRepo) UpdateStatus(ctx context.Context, id int64, clientStatus, internalStatus string) error {
	query := "UPDATE intake_leads SET client_status = $1, internal_status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3"
	if err := r.store.Exec(ctx, query, clientStatus, internalStatus, id); err != nil {
		return fmt.Errorf("project_intake: update status failed: %w", err)
	}
	return nil
}
