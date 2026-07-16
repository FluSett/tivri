package project_intake

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

func (r *PostgresRepository) Save(ctx context.Context, ld *Lead) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("project_intake: begin tx failed: %w", err)
	}
	defer tx.Rollback(ctx)

	query := "INSERT INTO intake_leads (company_name, project_scope, budget, contact_email, contact_info, deadline_needed, deadline_spec, is_custom_budget, client_status, internal_status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id"
	err = tx.QueryRow(ctx, query, ld.CompanyName, ld.ProjectScope, ld.Budget, ld.ContactEmail, ld.ContactInfo, ld.DeadlineNeeded, ld.DeadlineSpec, ld.IsCustomBudget, ld.ClientStatus, ld.InternalStatus).Scan(&ld.ID)
	if err != nil {
		return fmt.Errorf("project_intake: save failed: %w", err)
	}

	evtPayload := ProjectAppliedEvent{
		ID:             ld.ID,
		CompanyName:    ld.CompanyName,
		ProjectScope:   ld.ProjectScope,
		Budget:         ld.Budget,
		ContactEmail:   ld.ContactEmail,
		ContactInfo:    ld.ContactInfo,
		DeadlineNeeded: ld.DeadlineNeeded,
		DeadlineSpec:   ld.DeadlineSpec,
		IsCustomBudget: ld.IsCustomBudget,
		Timestamp:      ld.CreatedAt,
	}

	payloadBytes, err := json.Marshal(evtPayload)
	if err != nil {
		return fmt.Errorf("project_intake: marshal event failed: %w", err)
	}

	_, err = tx.Exec(ctx, "INSERT INTO outbox_events (type, payload) VALUES ($1, $2)", "project_intake.applied", payloadBytes)
	if err != nil {
		return fmt.Errorf("project_intake: save event failed: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("project_intake: commit failed: %w", err)
	}

	return nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]Lead, error) {
	query := "SELECT id, company_name, project_scope, budget, contact_email, contact_info, deadline_needed, deadline_spec, is_custom_budget, client_status, internal_status, created_at, updated_at FROM intake_leads ORDER BY id DESC"
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("project_intake: query leads failed: %w", err)
	}
	defer rows.Close()

	var list []Lead
	for rows.Next() {
		var ld Lead
		err := rows.Scan(&ld.ID, &ld.CompanyName, &ld.ProjectScope, &ld.Budget, &ld.ContactEmail, &ld.ContactInfo, &ld.DeadlineNeeded, &ld.DeadlineSpec, &ld.IsCustomBudget, &ld.ClientStatus, &ld.InternalStatus, &ld.CreatedAt, &ld.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("project_intake: scan lead failed: %w", err)
		}

		list = append(list, ld)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("project_intake: row iteration failed: %w", err)
	}

	return list, nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id int64, clientStatus, internalStatus string) error {
	query := "UPDATE intake_leads SET client_status = $1, internal_status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3"
	_, err := r.pool.Exec(ctx, query, clientStatus, internalStatus, id)
	if err != nil {
		return fmt.Errorf("project_intake: update status failed: %w", err)
	}

	return nil
}
