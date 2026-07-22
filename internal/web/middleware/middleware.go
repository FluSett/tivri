package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/csrf"

	"tivri/internal/config"
	"tivri/internal/core"
	"tivri/internal/core/security"
	"tivri/internal/i18n"
	"tivri/internal/web/render"
)

type contextKey string

const (
	baseDataKey  = contextKey("baseData")
	requestIDKey = contextKey("requestID")
	cspNonceKey  = contextKey("cspNonce")
)

type ipRateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
}

var defaultRateLimiter = &ipRateLimiter{
	requests: make(map[string][]time.Time),
}

func BaseDataMiddleware(cfg *config.Config, translator *i18n.Translator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lang := security.ResolveLocale(r)
			t := translator.Get(lang)

			nonce, _ := r.Context().Value(cspNonceKey).(string)
			if nonce == "" {
				nonce = r.Header.Get("X-CSP-Nonce")
			}

			baseData := render.BaseData{
				CurrentPath:             r.URL.Path,
				Lang:                    lang,
				T:                       t,
				TurnstileSiteKey:        cfg.TurnstileSiteKey,
				AppURL:                  cfg.AppURL,
				ContactEmail:            cfg.ContactEmail,
				Nonce:                   nonce,
				CloudflareInsightsToken: cfg.CloudflareInsightsToken,
				CSRFToken:               csrf.TemplateField(r),
				CSRFTokenVal:            csrf.Token(r),
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

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			reqID = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-ID", reqID)
		ctx := context.WithValue(r.Context(), requestIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return reqID
	}
	return ""
}

func CSPNonceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		nonce := base64.StdEncoding.EncodeToString(b)
		w.Header().Set("X-CSP-Nonce", nonce)
		ctx := context.WithValue(r.Context(), cspNonceKey, nonce)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RateLimiterMiddleware(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				parts := strings.Split(forwarded, ",")
				ip = strings.TrimSpace(parts[0])
			} else if host, _, err := net.SplitHostPort(ip); err == nil {
				ip = host
			}

			now := time.Now()
			defaultRateLimiter.mu.Lock()
			timestamps := defaultRateLimiter.requests[ip]
			var valid []time.Time
			for _, t := range timestamps {
				if now.Sub(t) < window {
					valid = append(valid, t)
				}
			}

			if len(valid) >= maxRequests {
				defaultRateLimiter.requests[ip] = valid
				defaultRateLimiter.mu.Unlock()
				w.Header().Set("Retry-After", "60")
				http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
				return
			}

			valid = append(valid, now)
			defaultRateLimiter.requests[ip] = valid
			defaultRateLimiter.mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
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

func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}
