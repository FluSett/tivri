package app

import (
	"encoding/json"
	"net/http"
	urlpkg "net/url"
)

func (a *App) handleAPILang(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	if lang != "en" && lang != "uk" && lang != "ru" {
		lang = "en"
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "lang",
		Value:    lang,
		Path:     "/",
		HttpOnly: true,
		Secure:   a.cfg.Env == "production",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   31536000,
	})

	currentURL := r.Header.Get("HX-Current-URL")
	path := "/"
	tab := "portfolio"

	if currentURL != "" {
		if parsed, err := urlpkg.Parse(currentURL); err == nil {
			path = parsed.Path
			if parsed.RawQuery != "" {
				path = path + "?" + parsed.RawQuery
			}
			if parsed.Query().Get("tab") != "" {
				tab = parsed.Query().Get("tab")
			}
		}
	}

	highQueueActive, _ := a.settingsRepo.GetHighQueue(r.Context())
	maintenanceActive, _ := a.settingsRepo.GetMaintenance(r.Context())
	pageData := PageData{
		CurrentPath:       path,
		Lang:              lang,
		T:                 a.translator.Get(lang),
		IsAdmin:           false,
		AdminTab:          tab,
		HighQueueActive:   highQueueActive,
		MaintenanceActive: maintenanceActive,
		TurnstileSiteKey:  a.cfg.TurnstileSiteKey,
		AppURL:            a.cfg.AppURL,
		ContactEmail:      a.cfg.ContactEmail,
		Nonce:             r.Header.Get("X-CSP-Nonce"),
		CloudflareInsightsToken: a.cfg.CloudflareInsightsToken,
	}

	var tmplKey string
	if len(path) >= 6 && path[:6] == "/admin" {
		if path == "/admin/login" {
			tmplKey = "login"
			pageData.IsAdmin = true
			pageData.IsAdminLogin = true
		} else {
			tmplKey = "admin"
			pageData.IsAdmin = true

			items, err := a.portfolioHandler.ListItems(r.Context())
			if err == nil {
				pageData.PortfolioItems = items
			}

			leads, err := a.leadHandler.ListLeads(r.Context())
			if err == nil {
				pageData.Leads = leads
				if raw, err := json.Marshal(leads); err == nil {
					pageData.LeadsJSON = string(raw)
				}
			}

			msgs, err := a.contactHandler.ListMessages(r.Context())
			if err == nil {
				pageData.ContactMessages = msgs
				if raw, err := json.Marshal(msgs); err == nil {
					pageData.MessagesJSON = string(raw)
				}
			}
		}
	} else {
		if maintenanceActive {
			tmplKey = "maintenance"
		} else {
			switch path {
			case "/privacy":
				tmplKey = "privacy"
			case "/terms":
				tmplKey = "terms"
			default:
				tmplKey = "home"
				items, err := a.portfolioHandler.ListItems(r.Context())
				if err == nil {
					pageData.PortfolioItems = items
				}
			}
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Push-Url", path)

	err := a.templates[tmplKey].ExecuteTemplate(w, "base.layout.html", pageData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if a.db != nil {
		if err := a.db.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"error","details":"database ping failed"}`))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
