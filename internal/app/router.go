package app

import (
	"encoding/json"
	"io/fs"
	"net"
	"net/http"
	urlpkg "net/url"
	"strings"
	"time"

	"tivri/internal/core/security"
	"tivri/internal/eventbus"
)

func (a *App) newRouter() (http.Handler, error) {
	mux := http.NewServeMux()
	subAssetsFS, err := fs.Sub(a.webFS, "assets")
	if err != nil {
		return nil, err
	}

	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(subAssetsFS))))

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(subAssetsFS, "favicons/favicon.png")
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(data)
	})

	mux.HandleFunc("/favicon.png", func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(subAssetsFS, "favicons/favicon.png")
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(data)
	})

	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("User-agent: *\nDisallow: /admin/\nDisallow: /admin\nAllow: /\n\nSitemap: https://tivri.cc/sitemap.xml\n"))
	})

	mux.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		now := time.Now().Format("2006-01-02")
		sitemap := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">
  <url>
    <loc>https://tivri.cc/</loc>
    <xhtml:link rel="alternate" hreflang="en" href="https://tivri.cc/?lang=en"/>
    <xhtml:link rel="alternate" hreflang="uk" href="https://tivri.cc/?lang=uk"/>
    <xhtml:link rel="alternate" hreflang="ru" href="https://tivri.cc/?lang=ru"/>
    <xhtml:link rel="alternate" hreflang="x-default" href="https://tivri.cc/"/>
    <lastmod>` + now + `</lastmod>
    <changefreq>daily</changefreq>
    <priority>1.0</priority>
  </url>
  <url>
    <loc>https://tivri.cc/?lang=en</loc>
    <xhtml:link rel="alternate" hreflang="en" href="https://tivri.cc/?lang=en"/>
    <xhtml:link rel="alternate" hreflang="uk" href="https://tivri.cc/?lang=uk"/>
    <xhtml:link rel="alternate" hreflang="ru" href="https://tivri.cc/?lang=ru"/>
    <xhtml:link rel="alternate" hreflang="x-default" href="https://tivri.cc/"/>
    <lastmod>` + now + `</lastmod>
    <changefreq>daily</changefreq>
    <priority>0.8</priority>
  </url>
  <url>
    <loc>https://tivri.cc/?lang=uk</loc>
    <xhtml:link rel="alternate" hreflang="en" href="https://tivri.cc/?lang=en"/>
    <xhtml:link rel="alternate" hreflang="uk" href="https://tivri.cc/?lang=uk"/>
    <xhtml:link rel="alternate" hreflang="ru" href="https://tivri.cc/?lang=ru"/>
    <xhtml:link rel="alternate" hreflang="x-default" href="https://tivri.cc/"/>
    <lastmod>` + now + `</lastmod>
    <changefreq>daily</changefreq>
    <priority>0.8</priority>
  </url>
  <url>
    <loc>https://tivri.cc/?lang=ru</loc>
    <xhtml:link rel="alternate" hreflang="en" href="https://tivri.cc/?lang=en"/>
    <xhtml:link rel="alternate" hreflang="uk" href="https://tivri.cc/?lang=uk"/>
    <xhtml:link rel="alternate" hreflang="ru" href="https://tivri.cc/?lang=ru"/>
    <xhtml:link rel="alternate" hreflang="x-default" href="https://tivri.cc/"/>
    <lastmod>` + now + `</lastmod>
    <changefreq>daily</changefreq>
    <priority>0.8</priority>
  </url>
</urlset>`
		_, _ = w.Write([]byte(sitemap))
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := a.db.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"error","details":"database ping failed"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

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
		maintenanceActive, _ := a.getMaintenanceSetting(r.Context())
		pageData := PageData{
			Lang:              lang,
			T:                 a.translator.Get(lang),
			IsAdmin:           false,
			AdminTab:          tab,
			HighQueueActive:   highQueueActive,
			MaintenanceActive: maintenanceActive,
			TurnstileSiteKey:  a.cfg.TurnstileSiteKey,
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
			if maintenanceActive {
				tmplKey = "maintenance"
			} else {
				tmplKey = "home"
				items, err := a.portfolioHandler.ListItems(r.Context())
				if err == nil {
					pageData.PortfolioItems = items
				}
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
				Lang:             lang,
				T:                a.translator.Get(lang),
				IsAdmin:          true,
				IsAdminLogin:     true,
				TurnstileSiteKey: a.cfg.TurnstileSiteKey,
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
					Lang:             lang,
					T:                a.translator.Get(lang),
					IsAdmin:          true,
					IsAdminLogin:     true,
					Error:            "Too many failed attempts. Locked out for 60 seconds.",
					TurnstileSiteKey: a.cfg.TurnstileSiteKey,
				}

				w.WriteHeader(http.StatusTooManyRequests)
				err = a.templates["login"].ExecuteTemplate(w, "base.layout.html", data)
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

			if username != a.cfg.AdminUsername || password != a.cfg.AdminPassword {
				a.securityMgr.RecordFailedAttempt(r)
				lang := security.ResolveLocale(r)
				data := PageData{
					Lang:             lang,
					T:                a.translator.Get(lang),
					IsAdmin:          true,
					IsAdminLogin:     true,
					Error:            "Invalid username or password",
					TurnstileSiteKey: a.cfg.TurnstileSiteKey,
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

	mux.HandleFunc("/admin/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     "admin_session",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1,
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
			maintenanceActive, _ := a.getMaintenanceSetting(r.Context())
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

			a.eventBus.Publish(r.Context(), eventbus.Event{
				Type:      "settings.high_queue_changed",
				Payload:   enabled,
				Timestamp: time.Now(),
			})
			w.WriteHeader(http.StatusOK)
		})(w, r)
	})

	mux.HandleFunc("/admin/settings/maintenance", func(w http.ResponseWriter, r *http.Request) {
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
			enabled := r.FormValue("maintenance") == "true" || r.FormValue("maintenance") == "on" || r.FormValue("maintenance") == "1"
			err = a.setMaintenanceSetting(r.Context(), enabled)
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
		})(w, r)
	})

	mux.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
		lang := security.ResolveLocale(r)
		data := PageData{
			CurrentPath:      "/privacy",
			Lang:             lang,
			T:                a.translator.Get(lang),
			TurnstileSiteKey: a.cfg.TurnstileSiteKey,
		}
		err = a.templates["privacy"].ExecuteTemplate(w, "base.layout.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/terms", func(w http.ResponseWriter, r *http.Request) {
		lang := security.ResolveLocale(r)
		data := PageData{
			CurrentPath:      "/terms",
			Lang:             lang,
			T:                a.translator.Get(lang),
			TurnstileSiteKey: a.cfg.TurnstileSiteKey,
		}
		err = a.templates["terms"].ExecuteTemplate(w, "base.layout.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			lang := security.ResolveLocale(r)
			data := PageData{
				CurrentPath:      r.URL.Path,
				Lang:             lang,
				T:                a.translator.Get(lang),
				TurnstileSiteKey: a.cfg.TurnstileSiteKey,
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
			CurrentPath:     "/",
			Lang:            lang,
			T:               a.translator.Get(lang),
			PortfolioItems:  items,
			HighQueueActive: highQueueActive,
			TurnstileSiteKey: a.cfg.TurnstileSiteKey,
		}

		err = a.templates["home"].ExecuteTemplate(w, "base.layout.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	return security.StructuredLogger(a.logger)(a.maintenanceMiddleware(mux)), nil
}

func (a *App) maintenanceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, "/assets/") || path == "/healthz" || strings.HasPrefix(path, "/admin") {
			next.ServeHTTP(w, r)
			return
		}

		active, err := a.getMaintenanceSetting(r.Context())
		if err != nil {
			a.logger.Error("failed to retrieve maintenance mode setting", "error", err)
		}

		if active {
			if path == "/api/lang" {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)

			lang := security.ResolveLocale(r)
			data := PageData{
				Lang:              lang,
				T:                 a.translator.Get(lang),
				MaintenanceActive: true,
				TurnstileSiteKey:  a.cfg.TurnstileSiteKey,
			}

			err = a.templates["maintenance"].ExecuteTemplate(w, "base.layout.html", data)
			if err != nil {
				a.logger.Error("failed to render maintenance template", "error", err)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}
