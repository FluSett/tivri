package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"

	"tivri/internal/config"
	"tivri/internal/core"
	"tivri/internal/core/security"
	"tivri/internal/i18n"
	"tivri/internal/web/render"
)

type contextKey string

const baseDataKey = contextKey("baseData")

func BaseDataMiddleware(cfg *config.Config, translator *i18n.Translator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lang := security.ResolveLocale(r)
			t := translator.Get(lang)

			baseData := render.BaseData{
				CurrentPath:             r.URL.Path,
				Lang:                    lang,
				T:                       t,
				TurnstileSiteKey:        cfg.TurnstileSiteKey,
				AppURL:                  cfg.AppURL,
				ContactEmail:            cfg.ContactEmail,
				Nonce:                   r.Header.Get("X-CSP-Nonce"),
				CloudflareInsightsToken: cfg.CloudflareInsightsToken,
				CSRFToken:               csrf.TemplateField(r),
			}

			ctx := context.WithValue(r.Context(), baseDataKey, baseData)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetBaseData(ctx context.Context) render.BaseData {
	if data, ok := ctx.Value(baseDataKey).(render.BaseData); ok {
		return data
	}
	return render.BaseData{}
}

func MaintenanceMiddleware(settingsRepo core.SettingsRepository, renderer *render.Renderer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if strings.HasPrefix(path, "/assets/") || path == "/healthz" || strings.HasPrefix(path, "/admin") {
				next.ServeHTTP(w, r)
				return
			}

			active, _ := settingsRepo.GetMaintenance(r.Context())
			if active {
				if path == "/api/lang" {
					next.ServeHTTP(w, r)
					return
				}

				w.WriteHeader(http.StatusServiceUnavailable)
				baseData := GetBaseData(r.Context())

				data := struct {
					render.BaseData
					MaintenanceActive bool
				}{
					BaseData:          baseData,
					MaintenanceActive: true,
				}

				_ = renderer.RenderPage(w, "maintenance", data)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
