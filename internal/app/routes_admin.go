package app

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"tivri/internal/core/security"
	"tivri/internal/eventbus"
)

func (a *App) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		lang := security.ResolveLocale(r)
		data := PageData{
			Lang:             lang,
			T:                a.translator.Get(lang),
			IsAdmin:          true,
			IsAdminLogin:     true,
			TurnstileSiteKey: a.cfg.TurnstileSiteKey,
			AppURL:           a.cfg.AppURL,
			ContactEmail:     a.cfg.ContactEmail,
		}

		err := a.templates["login"].ExecuteTemplate(w, "base.layout.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		if a.securityMgr.IsLockedOut(r) {
			lang := security.ResolveLocale(r)
			data := PageData{
				Lang:             lang,
				T:                a.translator.Get(lang),
				IsAdmin:          true,
				IsAdminLogin:     true,
				Error:            "Too many failed attempts. Locked out for 60 seconds.",
				TurnstileSiteKey: a.cfg.TurnstileSiteKey,
				AppURL:           a.cfg.AppURL,
				ContactEmail:     a.cfg.ContactEmail,
			}

			w.WriteHeader(http.StatusTooManyRequests)
			err := a.templates["login"].ExecuteTemplate(w, "base.layout.html", data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if a.cfg.TurnstileSiteKey != "" {
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
			ok, err := security.VerifyTurnstile(a.cfg.TurnstileSecretKey, token, ip)
			if err != nil || !ok {
				lang := security.ResolveLocale(r)
				data := PageData{
					Lang:             lang,
					T:                a.translator.Get(lang),
					IsAdmin:          true,
					IsAdminLogin:     true,
					Error:            a.translator.Get(lang).Get("ValTurnstileFailed"),
					TurnstileSiteKey: a.cfg.TurnstileSiteKey,
					AppURL:           a.cfg.AppURL,
					ContactEmail:     a.cfg.ContactEmail,
				}
				w.WriteHeader(http.StatusBadRequest)
				err = a.templates["login"].ExecuteTemplate(w, "base.layout.html", data)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		userHash := sha256.Sum256([]byte(username))
		cfgUserHash := sha256.Sum256([]byte(a.cfg.AdminUsername))
		passHash := sha256.Sum256([]byte(password))
		cfgPassHash := sha256.Sum256([]byte(a.cfg.AdminPassword))

		userMatch := subtle.ConstantTimeCompare(userHash[:], cfgUserHash[:]) == 1
		passMatch := subtle.ConstantTimeCompare(passHash[:], cfgPassHash[:]) == 1

		if !userMatch || !passMatch {
			a.securityMgr.RecordFailedAttempt(r)
			lang := security.ResolveLocale(r)
			data := PageData{
				Lang:             lang,
				T:                a.translator.Get(lang),
				IsAdmin:          true,
				IsAdminLogin:     true,
				Error:            "Invalid username or password",
				TurnstileSiteKey: a.cfg.TurnstileSiteKey,
				AppURL:           a.cfg.AppURL,
				ContactEmail:     a.cfg.ContactEmail,
			}

			w.WriteHeader(http.StatusUnauthorized)
			err := a.templates["login"].ExecuteTemplate(w, "base.layout.html", data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		a.securityMgr.RecordSuccessfulAttempt(r)
		token, err := a.securityMgr.GenerateToken(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "admin_session",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   a.cfg.Env == "production",
			SameSite: http.SameSiteStrictMode,
			MaxAge:   86400,
		})

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

func (a *App) handleAdminLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   a.cfg.Env == "production",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

func (a *App) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	lang := security.ResolveLocale(r)
	items, err := a.portfolioHandler.ListItems(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	leads, err := a.leadHandler.ListLeads(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	msgs, err := a.contactHandler.ListMessages(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tab := r.URL.Query().Get("tab")
	if tab == "" {
		tab = "portfolio"
	}

	var leadsJSON, msgsJSON string
	if raw, err := json.Marshal(leads); err == nil {
		leadsJSON = string(raw)
	}

	if raw, err := json.Marshal(msgs); err == nil {
		msgsJSON = string(raw)
	}

	highQueueActive, _ := a.settingsRepo.GetHighQueue(r.Context())
	maintenanceActive, _ := a.settingsRepo.GetMaintenance(r.Context())
	data := PageData{
		Lang:              lang,
		T:                 a.translator.Get(lang),
		PortfolioItems:    items,
		Leads:             leads,
		ContactMessages:   msgs,
		LeadsJSON:         leadsJSON,
		MessagesJSON:      msgsJSON,
		IsAdmin:           true,
		AdminTab:          tab,
		HighQueueActive:   highQueueActive,
		MaintenanceActive: maintenanceActive,
		TurnstileSiteKey:  a.cfg.TurnstileSiteKey,
		AppURL:            a.cfg.AppURL,
		ContactEmail:      a.cfg.ContactEmail,
	}

	err = a.templates["admin"].ExecuteTemplate(w, "base.layout.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleAdminSettingsHighQueue(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	enabled := r.FormValue("high_queue") == "true" || r.FormValue("high_queue") == "on" || r.FormValue("high_queue") == "1"
	err = a.settingsRepo.SetHighQueue(r.Context(), enabled)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.eventBus.Publish(r.Context(), eventbus.Event{
		Type:      "settings.high_queue_changed",
		Payload:   enabled,
		Timestamp: time.Now(),
	})
	w.WriteHeader(http.StatusOK)
}

func (a *App) handleAdminSettingsMaintenance(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	enabled := r.FormValue("maintenance") == "true" || r.FormValue("maintenance") == "on" || r.FormValue("maintenance") == "1"
	err = a.settingsRepo.SetMaintenance(r.Context(), enabled)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.eventBus.Publish(r.Context(), eventbus.Event{
		Type:      "settings.maintenance_changed",
		Payload:   enabled,
		Timestamp: time.Now(),
	})
	w.WriteHeader(http.StatusOK)
}
