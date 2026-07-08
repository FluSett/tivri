package project_intake

import (
	"context"
	"encoding/json"
	"time"
)

type Lead struct {
	ID             int64     `json:"id"`
	CompanyName    string    `json:"companyName"`
	ProjectScope   string    `json:"projectScope"`
	Budget         int64     `json:"budget"`
	ContactEmail   string    `json:"contactEmail"`
	ContactInfo    string    `json:"contactInfo"`
	DeadlineNeeded bool      `json:"deadlineNeeded"`
	DeadlineSpec   string    `json:"deadlineSpec"`
	IsCustomBudget bool      `json:"isCustomBudget"`
	ClientStatus   string    `json:"clientStatus"`
	InternalStatus string    `json:"internalStatus"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (l Lead) MarshalJSON() ([]byte, error) {
	type Alias Lead
	return json.Marshal(&struct {
		Alias
		CreatedAtStr string `json:"createdAtStr"`
		UpdatedAtStr string `json:"updatedAtStr"`
	}{
		Alias:        Alias(l),
		CreatedAtStr: l.CreatedAt.Format("2006-01-02 15:04"),
		UpdatedAtStr: l.UpdatedAt.Format("2006-01-02 15:04"),
	})
}

type Repository interface {
	Save(ctx context.Context, ld *Lead) error

	List(ctx context.Context) ([]Lead, error)

	UpdateStatus(ctx context.Context, id int64, clientStatus, internalStatus string) error
}
