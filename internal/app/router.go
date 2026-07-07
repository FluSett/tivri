package app

import (
	"encoding/json"
	"io/fs"
	"net/http"
	urlpkg "net/url"
	"strings"

	"tivri/internal/core/security"
)

func (a *App) newRouter() (http.Handler, error) {
	mux := http.NewServeMux()
	subAssetsFS, err := fs.Sub(a.webFS, "assets")
	if err != nil {
		return nil, err
	}

	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(subAssetsFS))))

	mux.HandleFunc("/api/lang", func(w http.ResponseWriter, r *http.Request) {
		lang := r.URL.Query().Get("lang")
		if lang != "en" && lang != "uk" && lang != "ru" {
			lang = "en"
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "lang",
			Value:    lang,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   31536000,
		})

		currentURL := r.Header.Get("HX-Current-URL")
		path := "/"
		tab := "portfolio"

		if currentURL != "" {
			if parsed, err := urlpkg.Parse(currentURL); err == nil {
				path = parsed.Path
				if parsed.Query().Get("tab") != "" {
					tab = parsed.Query().Get("tab")
				}
			}
		}

		highQueueActive, _ := a.getHighQueueSetting(r.Context())
		pageData := PageData{
			Lang:            lang,
			T:               a.translator.Get(lang),
			IsAdmin:         false,
			AdminTab:        tab,
			HighQueueActive: highQueueActive,
		}

		var tmplKey string
		if strings.HasPrefix(path, "/admin") {
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
			tmplKey = "home"
			items, err := a.portfolioHandler.ListItems(r.Context())
			if err == nil {
				pageData.PortfolioItems = items
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("HX-Push-Url", path)

		err = a.templates[tmplKey].ExecuteTemplate(w, "base.layout.html", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/api/intake", a.leadHandler.Create)
	mux.HandleFunc("/api/contact", a.contactHandler.Create)

	mux.HandleFunc("/admin/portfolio", func(w http.ResponseWriter, r *http.Request) {
		a.securityMgr.CookieAuth(a.cfg.AdminUsername, a.cfg.AdminPassword, a.portfolioHandler.Create)(w, r)
	})

	mux.HandleFunc("/admin/leads/status", func(w http.ResponseWriter, r *http.Request) {
		a.securityMgr.CookieAuth(a.cfg.AdminUsername, a.cfg.AdminPassword, a.leadHandler.UpdateStatus)(w, r)
	})

	mux.HandleFunc("/admin/messages/status", func(w http.ResponseWriter, r *http.Request) {
		a.securityMgr.CookieAuth(a.cfg.AdminUsername, a.cfg.AdminPassword, a.contactHandler.UpdateStatus)(w, r)
	})

	mux.HandleFunc("/admin/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			lang := security.ResolveLocale(r)
			data := PageData{
				Lang:         lang,
				T:            a.translator.Get(lang),
				IsAdmin:      true,
				IsAdminLogin: true,
			}

			err = a.templates["login"].ExecuteTemplate(w, "base.layout.html", data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			return
		}

		if r.Method == http.MethodPost {
			if a.securityMgr.IsLockedOut(r) {
				lang := security.ResolveLocale(r)
				data := PageData{
					Lang:         lang,
					T:            a.translator.Get(lang),
					IsAdmin:      true,
					IsAdminLogin: true,
					Error:        "Too many failed attempts. Locked out for 60 seconds.",
				}

				w.WriteHeader(http.StatusTooManyRequests)
				err = a.templates["login"].ExecuteTemplate(w, "base.layout.html", data)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}

				return
			}

			username := r.FormValue("username")
			password := r.FormValue("password")

			if username != a.cfg.AdminUsername || password != a.cfg.AdminPassword {
				a.securityMgr.RecordFailedAttempt(r)
				lang := security.ResolveLocale(r)
				data := PageData{
					Lang:         lang,
					T:            a.translator.Get(lang),
					IsAdmin:      true,
					IsAdminLogin: true,
					Error:        "Invalid username or password",
				}

				w.WriteHeader(http.StatusUnauthorized)
				err = a.templates["login"].ExecuteTemplate(w, "base.layout.html", data)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}

				return
			}

			a.securityMgr.RecordSuccessfulAttempt(r)
			token, err := a.securityMgr.GenerateToken()
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
	})

	mux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		a.securityMgr.CookieAuth(a.cfg.AdminUsername, a.cfg.AdminPassword, func(w http.ResponseWriter, r *http.Request) {
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

			highQueueActive, _ := a.getHighQueueSetting(r.Context())
			data := PageData{
				Lang:            lang,
				T:               a.translator.Get(lang),
				PortfolioItems:  items,
				Leads:           leads,
				ContactMessages: msgs,
				LeadsJSON:       leadsJSON,
				MessagesJSON:    msgsJSON,
				IsAdmin:         true,
				AdminTab:        tab,
				HighQueueActive: highQueueActive,
			}

			err = a.templates["admin"].ExecuteTemplate(w, "base.layout.html", data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})(w, r)
	})

	mux.HandleFunc("/admin/settings/high-queue", func(w http.ResponseWriter, r *http.Request) {
		a.securityMgr.CookieAuth(a.cfg.AdminUsername, a.cfg.AdminPassword, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			err := r.ParseForm()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			enabled := r.FormValue("high_queue") == "true" || r.FormValue("high_queue") == "on" || r.FormValue("high_queue") == "1"
			err = a.setHighQueueSetting(r.Context(), enabled)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})(w, r)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			lang := security.ResolveLocale(r)
			data := PageData{
				Lang: lang,
				T:    a.translator.Get(lang),
			}

			err = a.templates["notFound"].ExecuteTemplate(w, "base.layout.html", data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			return
		}

		lang := security.ResolveLocale(r)
		items, err := a.portfolioHandler.ListItems(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		highQueueActive, _ := a.getHighQueueSetting(r.Context())
		data := PageData{
			Lang:            lang,
			T:               a.translator.Get(lang),
			PortfolioItems:  items,
			HighQueueActive: highQueueActive,
		}

		err = a.templates["home"].ExecuteTemplate(w, "base.layout.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	return security.StructuredLogger(a.logger)(mux), nil
}
