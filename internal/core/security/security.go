package security

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"tivri/internal/datastore"

	"github.com/jackc/pgx/v5/pgxpool"
)

const bruteForceLockoutDuration = 60 * time.Second
const cleanupLoopInterval = 5 * time.Minute
const sessionExpirationDuration = 24 * time.Hour

type SecurityManager struct {
	mu          sync.RWMutex
	failedCount map[string]int
	lockoutTime map[string]time.Time
	db          *pgxpool.Pool
	logger      *slog.Logger
}

func NewSecurityManager(ctx context.Context, logger *slog.Logger, db *pgxpool.Pool) *SecurityManager {
	sm := &SecurityManager{
		failedCount: make(map[string]int),
		lockoutTime: make(map[string]time.Time),
		db:          db,
		logger:      logger,
	}

	go sm.cleanupLoop(ctx)
	return sm
}

func (sm *SecurityManager) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(cleanupLoopInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sm.mu.Lock()
			now := time.Now()

			for ip, until := range sm.lockoutTime {
				if now.Before(until) {
					continue
				}

				delete(sm.lockoutTime, ip)
				delete(sm.failedCount, ip)
			}
			sm.mu.Unlock()

			if sm.db != nil {
				tx, txErr := sm.db.Begin(ctx)
				if txErr == nil {
					_, _ = tx.Exec(ctx, "SELECT set_config('app.current_role', 'system', true)")
					_, err := tx.Exec(ctx, "DELETE FROM admin_sessions WHERE expires_at < $1", now)
					if err != nil {
						sm.logger.Error("security: failed to cleanup expired sessions", slog.Any("error", err))
					}
					_ = tx.Commit(ctx)
				}
			}
		}
	}
}

func (sm *SecurityManager) getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

func (sm *SecurityManager) IsLockedOut(r *http.Request) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	ip := sm.getIP(r)
	until, ok := sm.lockoutTime[ip]
	if ok && time.Now().Before(until) {
		return true
	}

	return false
}

func (sm *SecurityManager) RecordFailedAttempt(r *http.Request) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	ip := sm.getIP(r)
	sm.failedCount[ip]++
	if sm.failedCount[ip] >= 3 {
		sm.lockoutTime[ip] = time.Now().Add(bruteForceLockoutDuration)
	}
}

func (sm *SecurityManager) RecordSuccessfulAttempt(r *http.Request) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	ip := sm.getIP(r)
	delete(sm.failedCount, ip)
	delete(sm.lockoutTime, ip)
}

func (sm *SecurityManager) GenerateToken(ctx context.Context) (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("security: token generation failed: %w", err)
	}

	token := hex.EncodeToString(b)
	expiresAt := time.Now().Add(sessionExpirationDuration)

	if sm.db != nil {
		tx, txErr := sm.db.Begin(ctx)
		if txErr != nil {
			return "", fmt.Errorf("security: failed to begin transaction for token: %w", txErr)
		}
		defer tx.Rollback(ctx)

		_, _ = tx.Exec(ctx, "SELECT set_config('app.current_role', 'admin', true)")
		_, err = tx.Exec(ctx, "INSERT INTO admin_sessions (token, expires_at) VALUES ($1, $2)", token, expiresAt)
		if err != nil {
			return "", fmt.Errorf("security: failed to store session: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return "", fmt.Errorf("security: failed to commit session token: %w", err)
		}
	}

	return token, nil
}

func (sm *SecurityManager) CookieAuth(adminUsername, adminPassword string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if sm.IsLockedOut(r) {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		cookie, err := r.Cookie("admin_session")
		if err != nil {
			if r.Method == http.MethodGet {
				http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
			return
		}

		if sm.db != nil {
			tx, txErr := sm.db.Begin(r.Context())
			if txErr != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer tx.Rollback(r.Context())

			_, _ = tx.Exec(r.Context(), "SELECT set_config('app.current_role', 'admin', true)")

			var expiresAt time.Time
			err = tx.QueryRow(r.Context(), "SELECT expires_at FROM admin_sessions WHERE token = $1", cookie.Value).Scan(&expiresAt)

			if err != nil {
				if r.Method == http.MethodGet {
					http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
				} else {
					w.WriteHeader(http.StatusUnauthorized)
				}
				return
			}

			if time.Now().After(expiresAt) {
				_, _ = tx.Exec(r.Context(), "DELETE FROM admin_sessions WHERE token = $1", cookie.Value)
				_ = tx.Commit(r.Context())
				if r.Method == http.MethodGet {
					http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
				} else {
					w.WriteHeader(http.StatusUnauthorized)
				}
				return
			}
			_ = tx.Commit(r.Context())
		}

		adminCtx := datastore.WithRole(r.Context(), datastore.RoleAdmin)
		next(w, r.WithContext(adminCtx))
	}
}

func (sm *SecurityManager) BasicAuth(adminUsername, adminPassword string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if sm.IsLockedOut(r) {
			w.WriteHeader(http.StatusTooManyRequests)
			_, err := w.Write([]byte("Too many failed login attempts. Please try again later."))
			if err != nil {
				sm.logger.Error("security: write failed", slog.Any("error", err))
			}
			return
		}

		username, password, ok := r.BasicAuth()
		if !ok || username != adminUsername || password != adminPassword {
			sm.RecordFailedAttempt(r)
			w.Header().Set("WWW-Authenticate", `Basic realm="Admin"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		sm.RecordSuccessfulAttempt(r)
		adminCtx := datastore.WithRole(r.Context(), datastore.RoleAdmin)
		next(w, r.WithContext(adminCtx))
	}
}

func ResolveLocale(r *http.Request) string {
	queryLang := r.URL.Query().Get("lang")
	if queryLang == "en" || queryLang == "uk" || queryLang == "ru" {
		return queryLang
	}

	cookie, err := r.Cookie("lang")
	if err == nil {
		lang := cookie.Value
		if lang == "en" || lang == "uk" || lang == "ru" {
			return lang
		}
	}

	accept := r.Header.Get("Accept-Language")
	if strings.Contains(accept, "uk") {
		return "uk"
	}
	if strings.Contains(accept, "ru") {
		return "ru"
	}
	return "en"
}

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(buf []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}

	return rw.ResponseWriter.Write(buf)
}

func StructuredLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)
			duration := time.Since(start)

			logger.Info("http request completed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.status),
				slog.Duration("duration", duration),
				slog.String("ip", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)
		})
	}
}

func ValidateTurnstileRequest(r *http.Request, secret string) (bool, error) {
	token := r.FormValue("cf-turnstile-response")
	if token == "" {
		return false, fmt.Errorf("missing turnstile token")
	}

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

	return VerifyTurnstile(secret, token, ip)
}
