package services

import (
	"context"
	"encoding/json"
	"fmt"
	"tivri/internal/core"
	"tivri/internal/datastore"
)

type ContactService struct {
	store       *datastore.Store
	contactRepo core.ContactRepository
	outboxRepo  core.OutboxRepository
}

func NewContactService(store *datastore.Store, contactRepo core.ContactRepository, outboxRepo core.OutboxRepository) *ContactService {
	return &ContactService{store: store, contactRepo: contactRepo, outboxRepo: outboxRepo}
}

func (s *ContactService) SendMessage(ctx context.Context, msg *core.ContactMessage) error {
	return s.store.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.contactRepo.Save(txCtx, msg); err != nil {
			return err
		}

		payloadBytes, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("contact_service: marshal event failed: %w", err)
		}

		outboxEvt := &core.OutboxEvent{
			Type:    "contact.created",
			Payload: payloadBytes,
		}
		if err := s.outboxRepo.Save(txCtx, outboxEvt); err != nil {
			return err
		}

		return nil
	})
}
