package services

import (
	"context"
	"encoding/json"
	"fmt"
	"tivri/internal/core"
	"tivri/internal/datastore"
)

type IntakeService struct {
	store      *datastore.Store
	intakeRepo core.LeadRepository
	outboxRepo core.OutboxRepository
}

func NewIntakeService(store *datastore.Store, intakeRepo core.LeadRepository, outboxRepo core.OutboxRepository) *IntakeService {
	return &IntakeService{store: store, intakeRepo: intakeRepo, outboxRepo: outboxRepo}
}

func (s *IntakeService) Apply(ctx context.Context, lead *core.Lead) error {
	return s.store.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.intakeRepo.Save(txCtx, lead); err != nil {
			return err
		}

		payload := core.ProjectAppliedEventPayload{
			ID:             lead.ID,
			CompanyName:    lead.CompanyName,
			ProjectScope:   lead.ProjectScope,
			Budget:         lead.Budget,
			ContactEmail:   lead.ContactEmail,
			ContactInfo:    lead.ContactInfo,
			DeadlineNeeded: lead.DeadlineNeeded,
			DeadlineSpec:   lead.DeadlineSpec,
			IsCustomBudget: lead.IsCustomBudget,
			Timestamp:      lead.CreatedAt,
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("intake_service: marshal event payload failed: %w", err)
		}

		outboxEvt := &core.OutboxEvent{
			Type:    "project_intake.applied",
			Payload: payloadBytes,
		}
		if err := s.outboxRepo.Save(txCtx, outboxEvt); err != nil {
			return err
		}

		return nil
	})
}
