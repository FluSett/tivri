package lead

import (
	"context"
	"time"
)

type Lead struct {
	ID             int64
	CompanyName    string
	ProjectScope   string
	Budget         int64
	ContactEmail   string
	ContactPhone   string
	ClientStatus   string
	InternalStatus string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Repository interface {
	Save(ctx context.Context, ld *Lead) error
	List(ctx context.Context) ([]Lead, error)
	UpdateStatus(ctx context.Context, id int64, clientStatus, internalStatus string) error
}
