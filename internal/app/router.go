package app

import (
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"

	"tivri/internal/config"
	"tivri/internal/core"
	"tivri/internal/core/security"
	"tivri/internal/i18n"
	"tivri/internal/web/handlers"
	"tivri/internal/web/middleware"
	"tivri/internal/web/render"
)

func newRouter(
	cfg *config.Config,
	logger *slog.Logger,
	webFS fs.FS,
	securityMgr *security.SecurityManager,
	settingsRepo core.SettingsRepository,
	translator *i18n.Translator,
	renderer *render.Renderer,
	publicHandler *handlers.PublicHandler,
	adminHandler *handlers.AdminHandler,
) (http.Handler, error) {
	mux := http.NewServeMux()
	subAssetsFS, err := fs.Sub(webFS, "assets")
	if err != nil {
		return nil, err
	}

	assetHandler := http.StripPrefix("/assets/", http.FileServer(http.FS(subAssetsFS)))
	uploadHandler := http.StripPrefix("/assets/uploads/", http.FileServer(http.Dir("web/assets/uploads")))

	mux.Handle("GET /assets/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=2592000, no-transform")
		if strings.HasPrefix(r.URL.Path, "/assets/uploads/") {
			uploadHandler.ServeHTTP(w, r)
			return
		}
		assetHandler.ServeHTTP(w, r)
	}))

	serveStaticFile := func(path, contentType string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			data, err := fs.ReadFile(subAssetsFS, path)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", contentType)
			w.Write(data)
		}
	}

	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(subAssetsFS, "favicons/favicon.ico")
		if err != nil {
			data, _ = fs.ReadFile(subAssetsFS, "favicons/favicon.png")
		}
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(data)
	})
	mux.HandleFunc("GET /favicon.png", serveStaticFile("favicons/favicon.png", "image/png"))
	mux.HandleFunc("GET /favicon.svg", serveStaticFile("favicons/favicon.svg", "image/svg+xml"))
	mux.HandleFunc("GET /apple-touch-icon.png", serveStaticFile("favicons/apple-touch-icon.png", "image/png"))

	uiMux := http.NewServeMux()

	uiMux.HandleFunc("GET /robots.txt", publicHandler.HandleRobots)
	uiMux.HandleFunc("GET /sitemap.xml", publicHandler.HandleSitemap)
	uiMux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	uiMux.HandleFunc("GET /privacy", publicHandler.HandlePrivacy)
	uiMux.HandleFunc("GET /terms", publicHandler.HandleTerms)
	uiMux.HandleFunc("GET /", publicHandler.HandleHome)
	uiMux.HandleFunc("GET /api/lang", publicHandler.HandleAPILang)

	uiMux.HandleFunc("GET /admin/login", adminHandler.HandleAdminLogin)
	uiMux.HandleFunc("POST /admin/login", adminHandler.HandleAdminLogin)
	uiMux.HandleFunc("GET /admin/logout", adminHandler.HandleAdminLogout)
	uiMux.HandleFunc("POST /admin/logout", adminHandler.HandleAdminLogout)

	adminAuth := func(h http.HandlerFunc) http.HandlerFunc {
		return securityMgr.CookieAuth(cfg.AdminUsername, cfg.AdminPassword, h)
	}

	uiMux.HandleFunc("GET /admin", adminAuth(adminHandler.HandleAdminDashboard))
	uiMux.HandleFunc("GET /admin/{tab}", adminAuth(adminHandler.HandleAdminDashboard))
	uiMux.HandleFunc("GET /admin/api/leads", adminAuth(adminHandler.HandleAdminLeadsPartial))
	uiMux.HandleFunc("GET /admin/api/messages", adminAuth(adminHandler.HandleAdminMessagesPartial))
	uiMux.HandleFunc("POST /admin/portfolio", adminAuth(adminHandler.HandlePortfolioCreate))
	uiMux.HandleFunc("POST /admin/leads/status", adminAuth(adminHandler.HandleLeadUpdateStatus))
	uiMux.HandleFunc("POST /admin/messages/status", adminAuth(adminHandler.HandleContactUpdateStatus))
	uiMux.HandleFunc("POST /admin/settings/high-queue", adminAuth(adminHandler.HandleAdminSettingsHighQueue))
	uiMux.HandleFunc("POST /admin/settings/maintenance", adminAuth(adminHandler.HandleAdminSettingsMaintenance))

	mux.HandleFunc("POST /api/intake", publicHandler.HandleIntakeCreate)
	mux.HandleFunc("POST /api/contact", publicHandler.HandleContactCreate)

	uiHandler := middleware.MaintenanceMiddleware(settingsRepo, renderer)(uiMux)
	uiHandler = middleware.BaseDataMiddleware(cfg, translator)(uiHandler)

	csrfMiddleware := csrf.Protect(
		cfg.CSRFAuthKey,
		csrf.Secure(cfg.Env == "production"),
		csrf.Path("/"),
		csrf.TrustedOrigins([]string{
			"localhost:8080",
			"127.0.0.1:8080",
			strings.TrimPrefix(strings.TrimPrefix(cfg.AppURL, "https://"), "http://"),
		}),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Warn("CSRF token validation failed",
				slog.String("path", r.URL.Path),
				slog.Any("reason", csrf.FailureReason(r)),
			)
			http.Error(w, "Forbidden - CSRF token invalid", http.StatusForbidden)
		})),
	)

	if cfg.Env != "test" {
		uiHandler = csrfMiddleware(uiHandler)
	}

	mux.Handle("/", uiHandler)

	return security.StructuredLogger(logger)(mux), nil
}
