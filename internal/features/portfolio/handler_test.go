package portfolio

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"tivri/internal/eventbus"
)

type mockRepository struct {
	items   []PortfolioItem
	saveErr error
	listErr error
}

func (m *mockRepository) Save(ctx context.Context, item *PortfolioItem) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	item.ID = int64(len(m.items) + 1)
	m.items = append(m.items, *item)
	return nil
}

func (m *mockRepository) List(ctx context.Context) ([]PortfolioItem, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.items, nil
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

func TestHandler_ListItems(t *testing.T) {
	repo := &mockRepository{
		items: []PortfolioItem{
			{ID: 1, Title: "Project A"},
		},
	}
	bus := &mockEventBus{}
	rnd := &mockRenderer{}
	h := NewHandler(repo, bus, rnd)

	items, err := h.ListItems(context.Background())
	if err != nil {
		t.Fatalf("failed to list items: %v", err)
	}

	if len(items) != 1 || items[0].Title != "Project A" {
		t.Errorf("unexpected items listed: %+v", items)
	}
}

func TestHandler_Create(t *testing.T) {
	t.Run("invalid method", func(t *testing.T) {
		repo := &mockRepository{}
		bus := &mockEventBus{}
		rnd := &mockRenderer{}
		h := NewHandler(repo, bus, rnd)

		req := httptest.NewRequest("GET", "/admin/portfolio", nil)
		w := httptest.NewRecorder()
		h.Create(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", w.Code)
		}
	})

	t.Run("success create portfolio item no upload", func(t *testing.T) {
		repo := &mockRepository{}
		bus := &mockEventBus{}
		rnd := &mockRenderer{}
		h := NewHandler(repo, bus, rnd)

		form := url.Values{}
		form.Add("title", "Project A")
		form.Add("description", "A very cool project description that is long enough.")

		req := httptest.NewRequest("POST", "/admin/portfolio", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		h.Create(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if rnd.lastTemplate != "portfolio.html" {
			t.Errorf("expected rendered template 'portfolio.html', got '%s'", rnd.lastTemplate)
		}

		if len(repo.items) != 1 {
			t.Errorf("expected 1 item saved, got %d", len(repo.items))
		}

		if repo.items[0].Title != "Project A" {
			t.Errorf("unexpected portfolio title: %s", repo.items[0].Title)
		}
	})
}
