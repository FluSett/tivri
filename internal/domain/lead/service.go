package lead

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateLead(ctx context.Context, companyName, projectScope string, budget int64, contactEmail, contactPhone string) (*Lead, error) {
	companyName = strings.TrimSpace(companyName)
	projectScope = strings.TrimSpace(projectScope)
	contactEmail = strings.TrimSpace(contactEmail)
	contactPhone = strings.TrimSpace(contactPhone)

	if len(companyName) < 2 {
		return nil, errors.New("company name too short")
	}

	if len(projectScope) < 20 {
		return nil, errors.New("project scope too short")
	}

	if budget < 100 {
		return nil, errors.New("invalid budget")
	}

	if len(contactEmail) < 5 || !strings.Contains(contactEmail, "@") {
		return nil, errors.New("invalid email")
	}

	ld := &Lead{
		CompanyName:    companyName,
		ProjectScope:   projectScope,
		Budget:         budget,
		ContactEmail:   contactEmail,
		ContactPhone:   contactPhone,
		ClientStatus:   "pending",
		InternalStatus: "pending",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.Save(ctx, ld); err != nil {
		return nil, err
	}

	return ld, nil
}

func (s *Service) ListLeads(ctx context.Context) ([]Lead, error) {
	return s.repo.List(ctx)
}

func (s *Service) UpdateLeadStatus(ctx context.Context, id int64, clientStatus, internalStatus string) error {
	return s.repo.UpdateStatus(ctx, id, clientStatus, internalStatus)
}
