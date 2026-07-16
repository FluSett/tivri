package messaging

import (
	"context"
	"encoding/json"
	"time"
)

type ContactMessage struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Topic     string    `json:"topic"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (c ContactMessage) MarshalJSON() ([]byte, error) {
	type Alias ContactMessage
	return json.Marshal(&struct {
		Alias
		CreatedAt    int64  `json:"createdAt"`
		UpdatedAt    int64  `json:"updatedAt"`
		CreatedAtStr string `json:"createdAtStr"`
		UpdatedAtStr string `json:"updatedAtStr"`
	}{
		Alias:        Alias(c),
		CreatedAt:    c.CreatedAt.Unix(),
		UpdatedAt:    c.UpdatedAt.Unix(),
		CreatedAtStr: c.CreatedAt.Format("2006-01-02 15:04"),
		UpdatedAtStr: c.UpdatedAt.Format("2006-01-02 15:04"),
	})
}

func (c *ContactMessage) UnmarshalJSON(data []byte) error {
	type Alias ContactMessage
	aux := &struct {
		*Alias
		CreatedAt int64 `json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	c.CreatedAt = time.Unix(aux.CreatedAt, 0)
	c.UpdatedAt = time.Unix(aux.UpdatedAt, 0)
	return nil
}

type Repository interface {
	Save(ctx context.Context, msg *ContactMessage) error

	List(ctx context.Context) ([]ContactMessage, error)

	UpdateStatus(ctx context.Context, id int64, status string) error
}
