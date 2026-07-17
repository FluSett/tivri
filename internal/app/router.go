package app

import (
	"io/fs"
	"net/http"
	"strings"

	"tivri/internal/core/security"
)

func (a *App) newRouter() (http.Handler, error) {
	mux := http.NewServeMux()
	subAssetsFS, err := fs.Sub(a.webFS, "assets")
	if err != nil {
		return nil, err
	}

	// Static Assets
	assetHandler := http.StripPrefix("/assets/", http.FileServer(http.FS(subAssetsFS)))
	mux.Handle("GET /assets/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=2592000, no-transform") // 30 days
		assetHandler.ServeHTTP(w, r)
	}))

	// Favicons
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(subAssetsFS, "favicons/favicon.ico")
		if err != nil {
			data, err = fs.ReadFile(subAssetsFS, "favicons/favicon.png")
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}
		w.Header().Set("Content-Type", "image/x-icon")
		_, _ = w.Write(data)
	})

	mux.HandleFunc("GET /favicon.png", func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(subAssetsFS, "favicons/favicon.png")
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(data)
	})

	mux.HandleFunc("GET /favicon.svg", func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(subAssetsFS, "favicons/favicon.svg")
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		_, _ = w.Write(data)
	})

	mux.HandleFunc("GET /apple-touch-icon.png", func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(subAssetsFS, "favicons/apple-touch-icon.png")
		if err != nil {
			data, err = fs.ReadFile(subAssetsFS, "favicons/favicon.png")
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(data)
	})

	// Public Routes
	mux.HandleFunc("GET /robots.txt", a.handleRobots)
	mux.HandleFunc("GET /sitemap.xml", a.handleSitemap)
	mux.HandleFunc("GET /healthz", a.handleHealthz)
	mux.HandleFunc("GET /privacy", a.handlePrivacy)
	mux.HandleFunc("GET /terms", a.handleTerms)
	mux.HandleFunc("GET /", a.handleHome)

	// API Routes
	mux.HandleFunc("GET /api/lang", a.handleAPILang)
	mux.HandleFunc("POST /api/intake", a.leadHandler.Create)
	mux.HandleFunc("POST /api/contact", a.contactHandler.Create)

	// Admin Routes (Login/Logout)
	mux.HandleFunc("GET /admin/login", a.handleAdminLogin)
	mux.HandleFunc("POST /admin/login", a.handleAdminLogin)
	mux.HandleFunc("GET /admin/logout", a.handleAdminLogout)
	mux.HandleFunc("POST /admin/logout", a.handleAdminLogout)

	// Admin Routes (Protected)
	adminAuth := func(h http.HandlerFunc) http.HandlerFunc {
		return a.securityMgr.CookieAuth(a.cfg.AdminUsername, a.cfg.AdminPassword, h)
	}

	mux.HandleFunc("GET /admin", adminAuth(a.handleAdminDashboard))
	mux.HandleFunc("POST /admin/portfolio", adminAuth(a.portfolioHandler.Create))
	mux.HandleFunc("POST /admin/leads/status", adminAuth(a.leadHandler.UpdateStatus))
	mux.HandleFunc("POST /admin/messages/status", adminAuth(a.contactHandler.UpdateStatus))
	mux.HandleFunc("POST /admin/settings/high-queue", adminAuth(a.handleAdminSettingsHighQueue))
	mux.HandleFunc("POST /admin/settings/maintenance", adminAuth(a.handleAdminSettingsMaintenance))

	return security.StructuredLogger(a.logger)(a.maintenanceMiddleware(mux)), nil
}

func (a *App) maintenanceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, "/assets/") || path == "/healthz" || strings.HasPrefix(path, "/admin") {
			next.ServeHTTP(w, r)
			return
		}

		active, err := a.settingsRepo.GetMaintenance(r.Context())
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
				Lang:                    lang,
				T:                       a.translator.Get(lang),
				MaintenanceActive:       true,
				TurnstileSiteKey:        a.cfg.TurnstileSiteKey,
				Nonce:                   r.Header.Get("X-CSP-Nonce"),
				CloudflareInsightsToken: a.cfg.CloudflareInsightsToken,
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
