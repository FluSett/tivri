package project_intake

import (
	"context"
	"time"
)

type Lead struct {
	ID             int64     `json:"id"`
	CompanyName    string    `json:"companyName"`
	ProjectScope   string    `json:"projectScope"`
	Budget         int64     `json:"budget"`
	ContactEmail   string    `json:"contactEmail"`
	ContactPhone   string    `json:"contactPhone"`
	ClientStatus   string    `json:"clientStatus"`
	InternalStatus string    `json:"internalStatus"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type Repository interface {
	Save(ctx context.Context, ld *Lead) error

	List(ctx context.Context) ([]Lead, error)

	UpdateStatus(ctx context.Context, id int64, clientStatus, internalStatus string) error
}
