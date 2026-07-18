package core

import "context"

type SettingsRepository interface {
	GetHighQueue(ctx context.Context) (bool, error)
	SetHighQueue(ctx context.Context, enabled bool) error
	GetMaintenance(ctx context.Context) (bool, error)
	SetMaintenance(ctx context.Context, enabled bool) error
}
