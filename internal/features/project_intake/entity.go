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
		CreatedAt    int64  `json:"createdAt"`
		UpdatedAt    int64  `json:"updatedAt"`
		CreatedAtStr string `json:"createdAtStr"`
		UpdatedAtStr string `json:"updatedAtStr"`
	}{
		Alias:        Alias(l),
		CreatedAt:    l.CreatedAt.Unix(),
		UpdatedAt:    l.UpdatedAt.Unix(),
		CreatedAtStr: l.CreatedAt.Format("2006-01-02 15:04"),
		UpdatedAtStr: l.UpdatedAt.Format("2006-01-02 15:04"),
	})
}

func (l *Lead) UnmarshalJSON(data []byte) error {
	type Alias Lead
	aux := &struct {
		*Alias
		CreatedAt int64 `json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
	}{
		Alias: (*Alias)(l),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	l.CreatedAt = time.Unix(aux.CreatedAt, 0)
	l.UpdatedAt = time.Unix(aux.UpdatedAt, 0)
	return nil
}

type Repository interface {
	Save(ctx context.Context, ld *Lead) error

	List(ctx context.Context) ([]Lead, error)

	UpdateStatus(ctx context.Context, id int64, clientStatus, internalStatus string) error
}
