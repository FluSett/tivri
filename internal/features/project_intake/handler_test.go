package project_intake

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
	leads     []Lead
	saveErr   error
	listErr   error
	updateErr error
}

func (m *mockRepository) Save(ctx context.Context, ld *Lead) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	ld.ID = int64(len(m.leads) + 1)
	ld.CreatedAt = time.Now()
	ld.UpdatedAt = time.Now()
	m.leads = append(m.leads, *ld)
	return nil
}

func (m *mockRepository) List(ctx context.Context) ([]Lead, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.leads, nil
}

func (m *mockRepository) UpdateStatus(ctx context.Context, id int64, clientStatus, internalStatus string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	for i, l := range m.leads {
		if l.ID == id {
			m.leads[i].ClientStatus = clientStatus
			m.leads[i].InternalStatus = internalStatus
			return nil
		}
	}
	return errors.New("lead not found")
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

		req := httptest.NewRequest("GET", "/api/intake", nil)
		w := httptest.NewRecorder()
		h.Create(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", w.Code)
		}
	})

	t.Run("validation error company name too short", func(t *testing.T) {
		repo := &mockRepository{}
		bus := &mockEventBus{}
		rnd := &mockRenderer{}
		h := NewHandler(repo, bus, rnd, translator, "")

		form := url.Values{}
		form.Add("company_name", "A")
		form.Add("project_scope", "This is a sufficiently long project scope description.")
		form.Add("contact_email", "test@tivri.cc")
		form.Add("budget", "250000")

		req := httptest.NewRequest("POST", "/api/intake", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		h.Create(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("success create project applied", func(t *testing.T) {
		repo := &mockRepository{}
		bus := &mockEventBus{}
		rnd := &mockRenderer{}
		h := NewHandler(repo, bus, rnd, translator, "")

		form := url.Values{}
		form.Add("company_name", "Acme Corp")
		form.Add("project_scope", "Need a custom database migration and API wrapper setup.")
		form.Add("contact_email", "client@acme.com")
		form.Add("budget", "750000")
		form.Add("deadline_needed", "true")
		form.Add("deadline_spec", "By the end of Q3")

		req := httptest.NewRequest("POST", "/api/intake", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		h.Create(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if len(repo.leads) != 1 {
			t.Errorf("expected 1 lead saved, got %d", len(repo.leads))
		}

		lead := repo.leads[0]
		if lead.CompanyName != "Acme Corp" || lead.Budget != 750000 || !lead.DeadlineNeeded || lead.DeadlineSpec != "By the end of Q3" {
			t.Errorf("unexpected lead fields: %+v", lead)
		}

	})
}

func TestHandler_UpdateStatus(t *testing.T) {
	translator, err := i18n.NewTranslator()
	if err != nil {
		t.Fatalf("failed to load translator: %v", err)
	}

	repo := &mockRepository{
		leads: []Lead{
			{ID: 1, CompanyName: "Acme Corp", ClientStatus: "pending", InternalStatus: "new"},
		},
	}
	bus := &mockEventBus{}
	rnd := &mockRenderer{}
	h := NewHandler(repo, bus, rnd, translator, "")

	t.Run("unauthorized method", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/leads/status", nil)
		w := httptest.NewRecorder()
		h.UpdateStatus(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", w.Code)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		form := url.Values{}
		form.Add("id", "invalid")
		req := httptest.NewRequest("POST", "/admin/leads/status", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		h.UpdateStatus(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("success update status", func(t *testing.T) {
		form := url.Values{}
		form.Add("id", "1")
		form.Add("type", "client")
		form.Add("status", "active")
		req := httptest.NewRequest("POST", "/admin/leads/status", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		h.UpdateStatus(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}

		if repo.leads[0].ClientStatus != "active" {
			t.Errorf("lead status update failed: %+v", repo.leads[0])
		}
	})
}

func TestLead_MarshalJSON(t *testing.T) {
	createdAt := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 7, 16, 13, 0, 0, 0, time.UTC)
	l := Lead{
		ID:             1,
		CompanyName:    "Acme Corp",
		ProjectScope:   "Scope",
		Budget:         500000,
		ContactEmail:   "client@acme.com",
		ContactInfo:    "info",
		DeadlineNeeded: false,
		DeadlineSpec:   "",
		IsCustomBudget: false,
		ClientStatus:   "pending",
		InternalStatus: "new",
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}

	data, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("failed to marshal lead: %v", err)
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
