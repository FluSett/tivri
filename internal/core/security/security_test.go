package security

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"tivri/internal/core/database"
)

func TestResolveLocale(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		headers        map[string]string
		cookieVal      string
		expectedLocale string
	}{
		{
			name:           "query parameter lang overrides all",
			url:            "/?lang=uk",
			headers:        map[string]string{"Accept-Language": "en-US"},
			cookieVal:      "en",
			expectedLocale: "uk",
		},
		{
			name:           "cookie takes precedence over Accept-Language",
			url:            "/privacy",
			headers:        map[string]string{"Accept-Language": "uk-UA"},
			cookieVal:      "ru",
			expectedLocale: "ru",
		},
		{
			name:           "Accept-Language uk fallback",
			url:            "/privacy",
			headers:        map[string]string{"Accept-Language": "uk-UA,uk;q=0.9,en;q=0.8"},
			expectedLocale: "uk",
		},
		{
			name:           "Accept-Language ru fallback",
			url:            "/privacy",
			headers:        map[string]string{"Accept-Language": "ru-RU,ru;q=0.9"},
			expectedLocale: "ru",
		},
		{
			name:           "Accept-Language default fallback en",
			url:            "/privacy",
			headers:        map[string]string{"Accept-Language": "fr-FR"},
			expectedLocale: "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			if tt.cookieVal != "" {
				req.AddCookie(&http.Cookie{Name: "lang", Value: tt.cookieVal})
			}

			result := ResolveLocale(req)
			if result != tt.expectedLocale {
				t.Errorf("expected locale %s, got %s", tt.expectedLocale, result)
			}
		})
	}
}

func TestSecurityManagerLockouts(t *testing.T) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		t.Skip("Skipping integration test requiring live database connection")
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, dsn, 10, 10)
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	defer pool.Close()

	_, _ = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS admin_login_attempts (
			ip TEXT PRIMARY KEY,
			attempts INTEGER DEFAULT 0,
			last_attempt TIMESTAMP WITH TIME ZONE
		);
		CREATE TABLE IF NOT EXISTS admin_sessions (
			token TEXT PRIMARY KEY,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	_, _ = pool.Exec(ctx, "DELETE FROM admin_login_attempts")
	_, _ = pool.Exec(ctx, "DELETE FROM admin_sessions")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	sm := NewSecurityManager(ctx, logger, pool)

	req := httptest.NewRequest("POST", "/admin/login", nil)
	req.RemoteAddr = "127.0.0.1:1234"

	if sm.IsLockedOut(req) {
		t.Error("expected not locked out initially")
	}

	for i := 0; i < 5; i++ {
		sm.RecordFailedAttempt(req)
	}

	if !sm.IsLockedOut(req) {
		t.Error("expected locked out after 5 failures")
	}

	sm.RecordSuccessfulAttempt(req)
	if sm.IsLockedOut(req) {
		t.Error("expected lockout reset after success")
	}

	token, err := sm.GenerateToken(ctx)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if len(token) == 0 {
		t.Error("expected non-empty token")
	}
}
