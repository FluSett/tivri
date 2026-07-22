package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"tivri/internal/config"
)

func TestHandleAdminLogout(t *testing.T) {
	cfg := &config.Config{Env: "development"}
	handler := &AdminHandler{cfg: cfg}

	req := httptest.NewRequest("GET", "/admin/logout", nil)
	w := httptest.NewRecorder()

	handler.HandleAdminLogout(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("expected status SeeOther (303), got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location != "/admin/login" {
		t.Errorf("expected redirect to /admin/login, got %q", location)
	}

	cookies := resp.Cookies()
	var adminCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "admin_session" {
			adminCookie = c
			break
		}
	}

	if adminCookie == nil {
		t.Fatal("expected admin_session cookie in response")
	}

	if adminCookie.MaxAge != -1 {
		t.Errorf("expected admin_session cookie MaxAge -1, got %d", adminCookie.MaxAge)
	}
}
