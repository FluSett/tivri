package handlers

import (
	"crypto/sha256"
	"crypto/subtle"
	"net"
	"net/http"
	"strings"

	"tivri/internal/core/security"
	"tivri/internal/web/middleware"
	"tivri/internal/web/render"
	"tivri/internal/web/response"
)

func (h *AdminHandler) HandleAdminLogin(w http.ResponseWriter, r *http.Request) {
	baseData := middleware.GetBaseData(r.Context())
	baseData.IsAdmin = true
	baseData.IsAdminLogin = true
	baseData.PageTitle = "Admin Login"

	if r.Method == http.MethodGet {
		data := struct{ render.BaseData }{BaseData: baseData}
		if err := h.renderer.RenderPage(w, "login", data); err != nil {
			response.Error(w, r, err, http.StatusInternalServerError, "")
		}
		return
	}

	if r.Method == http.MethodPost {
		if h.securityMgr.IsLockedOut(r) {
			baseData.Error = "Too many failed attempts. Locked out for 60 seconds."
			data := struct{ render.BaseData }{BaseData: baseData}
			w.WriteHeader(http.StatusTooManyRequests)
			if err := h.renderer.RenderPage(w, "login", data); err != nil {
				response.Error(w, r, err, http.StatusInternalServerError, "")
			}
			return
		}

		if h.cfg.TurnstileSiteKey != "" {
			token := r.FormValue("cf-turnstile-response")
			var ip string
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				parts := strings.Split(forwarded, ",")
				ip = strings.TrimSpace(parts[0])
			} else {
				if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
					ip = host
				} else {
					ip = r.RemoteAddr
				}
			}
			ok, err := security.VerifyTurnstile(h.cfg.TurnstileSecretKey, token, ip)
			if err != nil || !ok {
				baseData.Error = baseData.T.Get("ValTurnstileFailed")
				data := struct{ render.BaseData }{BaseData: baseData}
				w.WriteHeader(http.StatusBadRequest)
				if err = h.renderer.RenderPage(w, "login", data); err != nil {
					response.Error(w, r, err, http.StatusInternalServerError, "")
				}
				return
			}
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		userHash := sha256.Sum256([]byte(username))
		cfgUserHash := sha256.Sum256([]byte(h.cfg.AdminUsername))
		passHash := sha256.Sum256([]byte(password))
		cfgPassHash := sha256.Sum256([]byte(h.cfg.AdminPassword))

		userMatch := subtle.ConstantTimeCompare(userHash[:], cfgUserHash[:]) == 1
		passMatch := subtle.ConstantTimeCompare(passHash[:], cfgPassHash[:]) == 1

		if !userMatch || !passMatch {
			h.securityMgr.RecordFailedAttempt(r)
			baseData.Error = "Invalid username or password"
			data := struct{ render.BaseData }{BaseData: baseData}
			w.WriteHeader(http.StatusUnauthorized)
			if err := h.renderer.RenderPage(w, "login", data); err != nil {
				response.Error(w, r, err, http.StatusInternalServerError, "")
			}
			return
		}

		h.securityMgr.RecordSuccessfulAttempt(r)
		token, err := h.securityMgr.GenerateToken(r.Context())
		if err != nil {
			response.Error(w, r, err, http.StatusInternalServerError, "")
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "admin_session",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   h.cfg.Env == "production",
			SameSite: http.SameSiteStrictMode,
			MaxAge:   86400,
		})

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

func (h *AdminHandler) HandleAdminLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.Env == "production",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}
