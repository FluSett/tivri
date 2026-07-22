package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tivri/internal/i18n"
	"tivri/internal/web/middleware"
	"tivri/internal/web/render"
)

func setupTestPublicHandler(t *testing.T) *PublicHandler {
	t.Helper()

	translator, err := i18n.NewTranslator()
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}

	renderer := &render.Renderer{}

	return NewPublicHandler(
		renderer,
		translator,
		nil,
		nil,
		nil,
		nil,
		nil,
		"",
		false,
	)
}

func TestHandleHealthz(t *testing.T) {
	handler := setupTestPublicHandler(t)

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	handler.HandleHealthz(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json content type, got %q", resp.Header.Get("Content-Type"))
	}
}

func TestHandleAPILang(t *testing.T) {
	handler := setupTestPublicHandler(t)

	req := httptest.NewRequest("GET", "/api/lang?lang=uk", nil)
	req.Header.Set("HX-Current-URL", "http://localhost:8080/admin/leads?page=1")
	w := httptest.NewRecorder()

	handler.HandleAPILang(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %d", resp.StatusCode)
	}

	hxLoc := resp.Header.Get("HX-Location")
	if hxLoc != "http://localhost:8080/admin/leads?page=1" {
		t.Errorf("expected HX-Location header, got %q", hxLoc)
	}

	cookies := resp.Cookies()
	var langCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "lang" {
			langCookie = c
			break
		}
	}

	if langCookie == nil {
		t.Fatal("expected lang cookie to be set")
	}

	if langCookie.Value != "uk" {
		t.Errorf("expected lang cookie value 'uk', got %q", langCookie.Value)
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	securedHandler := middleware.SecurityHeadersMiddleware(nextHandler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	securedHandler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("missing or invalid X-Content-Type-Options header")
	}
	if resp.Header.Get("X-Frame-Options") != "DENY" {
		t.Errorf("missing or invalid X-Frame-Options header")
	}
	if resp.Header.Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Errorf("missing or invalid Referrer-Policy header")
	}
	if !strings.Contains(resp.Header.Get("Permissions-Policy"), "camera=()") {
		t.Errorf("missing or invalid Permissions-Policy header")
	}
}
