package app

import (
	"context"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"tivri"
	"tivri/internal/config"
	"tivri/internal/core/security"
	"tivri/internal/eventbus"
	"tivri/internal/features/messaging"
	"tivri/internal/features/portfolio"
	"tivri/internal/features/project_intake"
	"tivri/internal/i18n"
)

type mockPortfolioRepository struct {
	items []portfolio.PortfolioItem
}

func (m *mockPortfolioRepository) Save(ctx context.Context, item *portfolio.PortfolioItem) error {
	m.items = append(m.items, *item)
	return nil
}
func (m *mockPortfolioRepository) List(ctx context.Context) ([]portfolio.PortfolioItem, error) {
	return m.items, nil
}

type mockLeadRepository struct {
	leads []project_intake.Lead
}

func (m *mockLeadRepository) Save(ctx context.Context, ld *project_intake.Lead) error {
	m.leads = append(m.leads, *ld)
	return nil
}
func (m *mockLeadRepository) List(ctx context.Context) ([]project_intake.Lead, error) {
	return m.leads, nil
}
func (m *mockLeadRepository) UpdateStatus(ctx context.Context, id int64, clientStatus, internalStatus string) error {
	return nil
}

type mockContactRepository struct {
	messages []messaging.ContactMessage
}

func (m *mockContactRepository) Save(ctx context.Context, msg *messaging.ContactMessage) error {
	m.messages = append(m.messages, *msg)
	return nil
}
func (m *mockContactRepository) List(ctx context.Context) ([]messaging.ContactMessage, error) {
	return m.messages, nil
}
func (m *mockContactRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	return nil
}

type mockSettingsRepository struct {
	highQueue   bool
	maintenance bool
}

func (m *mockSettingsRepository) GetHighQueue(ctx context.Context) (bool, error) {
	return m.highQueue, nil
}
func (m *mockSettingsRepository) SetHighQueue(ctx context.Context, enabled bool) error {
	m.highQueue = enabled
	return nil
}
func (m *mockSettingsRepository) GetMaintenance(ctx context.Context) (bool, error) {
	return m.maintenance, nil
}
func (m *mockSettingsRepository) SetMaintenance(ctx context.Context, enabled bool) error {
	m.maintenance = enabled
	return nil
}

type mockEventBus struct{}

func (m *mockEventBus) Subscribe(eventType string, handler eventbus.Handler) {}
func (m *mockEventBus) Publish(ctx context.Context, e eventbus.Event)        {}

type mockRenderer struct{}

func (m *mockRenderer) ExecuteTemplate(w io.Writer, name string, data any) error {
	_, _ = w.Write([]byte("rendered"))
	return nil
}

func setupTestApp(ctx context.Context) (*App, error) {
	translator, err := i18n.NewTranslator()
	if err != nil {
		return nil, err
	}

	webUIFS, err := fs.Sub(tivri.WebFS, "web")
	if err != nil {
		return nil, err
	}

	templates, err := parseTemplates(webUIFS)
	if err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mockPortfolioRepo := &mockPortfolioRepository{}
	mockLeadRepo := &mockLeadRepository{}
	mockContactRepo := &mockContactRepository{}
	mockSettingsRepo := &mockSettingsRepository{}
	mockBus := &mockEventBus{}
	mockRnd := &mockRenderer{}

	a := &App{
		cfg: &config.Config{
			Env:           "development",
			Port:          "8080",
			AdminUsername: "admin",
			AdminPassword: "password",
			AppURL:        "http://localhost:8080",
			ContactEmail:  "contact@tivri.cc",
		},
		translator:       translator,
		templates:        templates,
		portfolioHandler: portfolio.NewHandler(mockPortfolioRepo, mockBus, mockRnd),
		leadHandler:      project_intake.NewHandler(mockLeadRepo, mockBus, mockRnd, translator, ""),
		contactHandler:   messaging.NewHandler(mockContactRepo, mockBus, mockRnd, translator, ""),
		settingsRepo:     mockSettingsRepo,
		logger:           logger,
		webFS:            webUIFS,
		securityMgr:      security.NewSecurityManager(ctx, logger, nil),
		eventBus:         mockBus,
	}

	return a, nil
}

func TestApp_Routing(t *testing.T) {
	ctx := context.Background()
	a, err := setupTestApp(ctx)
	if err != nil {
		t.Fatalf("failed to setup test app: %v", err)
	}

	router, err := a.newRouter()
	if err != nil {
		t.Fatalf("failed to initialize router: %v", err)
	}

	t.Run("GET /healthz", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/healthz", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if w.Body.String() != `{"status":"ok"}` {
			t.Errorf("expected body '{\"status\":\"ok\"}', got %q", w.Body.String())
		}
	})

	t.Run("GET /privacy and Security Headers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/privacy", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /api/lang sets language cookie with security options", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/lang?lang=uk", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		cookies := w.Result().Cookies()
		var langCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == "lang" {
				langCookie = c
				break
			}
		}

		if langCookie == nil {
			t.Fatal("expected lang cookie in response, got none")
		}

		if langCookie.Value != "uk" {
			t.Errorf("expected lang cookie value 'uk', got %s", langCookie.Value)
		}

		if langCookie.SameSite != http.SameSiteLaxMode {
			t.Errorf("expected SameSite Lax, got %v", langCookie.SameSite)
		}
	})

	t.Run("POST /admin/login failure (unauthorized)", func(t *testing.T) {
		form := url.Values{}
		form.Add("username", "wrong")
		form.Add("password", "wrong")

		req := httptest.NewRequest("POST", "/admin/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("POST /admin/login success (redirect and session cookie)", func(t *testing.T) {
		form := url.Values{}
		form.Add("username", "admin")
		form.Add("password", "password")

		req := httptest.NewRequest("POST", "/admin/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusSeeOther {
			t.Errorf("expected redirect 303, got %d", w.Code)
		}

		cookies := w.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == "admin_session" {
				sessionCookie = c
				break
			}
		}

		if sessionCookie == nil {
			t.Fatal("expected admin_session cookie, got none")
		}

		if sessionCookie.SameSite != http.SameSiteStrictMode {
			t.Errorf("expected SameSite Strict for admin session cookie, got %v", sessionCookie.SameSite)
		}
	})

	t.Run("GET /admin protected dashboard redirect", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusSeeOther {
			t.Errorf("expected redirect 303 to login, got %d", w.Code)
		}
	})
}
