package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"tivri/internal/eventbus"
	"tivri/internal/i18n"
)

type mockRepository struct {
	messages  []ContactMessage
	saveErr   error
	listErr   error
	updateErr error
}

func (m *mockRepository) Save(ctx context.Context, msg *ContactMessage) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	msg.ID = int64(len(m.messages) + 1)
	msg.CreatedAt = time.Now()
	msg.UpdatedAt = time.Now()
	m.messages = append(m.messages, *msg)
	return nil
}

func (m *mockRepository) List(ctx context.Context) ([]ContactMessage, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.messages, nil
}

func (m *mockRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	for i, msg := range m.messages {
		if msg.ID == id {
			m.messages[i].Status = status
			return nil
		}
	}
	return errors.New("message not found")
}

type mockEventBus struct {
	events []eventbus.Event
}

func (m *mockEventBus) Subscribe(eventType string, handler eventbus.Handler) {}
func (m *mockEventBus) Publish(ctx context.Context, e eventbus.Event) {
	m.events = append(m.events, e)
}

type mockRenderer struct {
	lastTemplate string
	lastData     any
	renderErr    error
}

func (m *mockRenderer) ExecuteTemplate(w io.Writer, name string, data any) error {
	if m.renderErr != nil {
		return m.renderErr
	}
	m.lastTemplate = name
	m.lastData = data
	_, _ = w.Write([]byte("rendered"))
	return nil
}

func TestHandler_Create(t *testing.T) {
	translator, err := i18n.NewTranslator()
	if err != nil {
		t.Fatalf("failed to load translator: %v", err)
	}

	t.Run("invalid method", func(t *testing.T) {
		repo := &mockRepository{}
		bus := &mockEventBus{}
		rnd := &mockRenderer{}
		h := NewHandler(repo, bus, rnd, translator, "")

		req := httptest.NewRequest("GET", "/api/contact", nil)
		w := httptest.NewRecorder()
		h.Create(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", w.Code)
		}
	})

	t.Run("validation error email invalid", func(t *testing.T) {
		repo := &mockRepository{}
		bus := &mockEventBus{}
		rnd := &mockRenderer{}
		h := NewHandler(repo, bus, rnd, translator, "")

		form := url.Values{}
		form.Add("email", "invalid-email")
		form.Add("topic", "Hello")
		form.Add("message", "This is a sufficiently long message description.")

		req := httptest.NewRequest("POST", "/api/contact", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		h.Create(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("success create message", func(t *testing.T) {
		repo := &mockRepository{}
		bus := &mockEventBus{}
		rnd := &mockRenderer{}
		h := NewHandler(repo, bus, rnd, translator, "")

		form := url.Values{}
		form.Add("email", "client@domain.com")
		form.Add("topic", "General Query")
		form.Add("message", "We would love to hire you to build our new core dashboard system.")

		req := httptest.NewRequest("POST", "/api/contact", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		h.Create(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if len(repo.messages) != 1 {
			t.Errorf("expected 1 message saved, got %d", len(repo.messages))
		}

		msg := repo.messages[0]
		if msg.Email != "client@domain.com" || msg.Topic != "General Query" || !strings.Contains(msg.Message, "dashboard system") {
			t.Errorf("unexpected message fields: %+v", msg)
		}
	})
}

func TestHandler_UpdateStatus(t *testing.T) {
	translator, err := i18n.NewTranslator()
	if err != nil {
		t.Fatalf("failed to load translator: %v", err)
	}

	repo := &mockRepository{
		messages: []ContactMessage{
			{ID: 1, Email: "client@domain.com", Status: "unread"},
		},
	}
	bus := &mockEventBus{}
	rnd := &mockRenderer{}
	h := NewHandler(repo, bus, rnd, translator, "")

	t.Run("unauthorized method", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/messages/status", nil)
		w := httptest.NewRecorder()
		h.UpdateStatus(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", w.Code)
		}
	})

	t.Run("success update status", func(t *testing.T) {
		form := url.Values{}
		form.Add("id", "1")
		form.Add("status", "answered")
		req := httptest.NewRequest("POST", "/admin/messages/status", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		h.UpdateStatus(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}

		if repo.messages[0].Status != "answered" {
			t.Errorf("message status update failed: %+v", repo.messages[0])
		}
	})
}

func TestContactMessage_MarshalJSON(t *testing.T) {
	createdAt := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 7, 16, 13, 0, 0, 0, time.UTC)
	c := ContactMessage{
		ID:        1,
		Email:     "client@domain.com",
		Topic:     "Topic",
		Message:   "Message text",
		Status:    "unread",
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("failed to marshal contact message: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if val, ok := m["createdAt"].(float64); !ok || int64(val) != createdAt.Unix() {
		t.Errorf("expected createdAt to be Unix timestamp %d, got %v", createdAt.Unix(), m["createdAt"])
	}

	if val, ok := m["updatedAt"].(float64); !ok || int64(val) != updatedAt.Unix() {
		t.Errorf("expected updatedAt to be Unix timestamp %d, got %v", updatedAt.Unix(), m["updatedAt"])
	}

	if val, ok := m["createdAtStr"].(string); !ok || val != "2026-07-16 12:00" {
		t.Errorf("expected createdAtStr to be '2026-07-16 12:00', got %v", m["createdAtStr"])
	}
}
