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
)

type SecurityManager struct {
	mu             sync.RWMutex
	failedCount    map[string]int
	lockoutTime    map[string]time.Time
	activeSessions map[string]time.Time
	logger         *slog.Logger
}

func NewSecurityManager(ctx context.Context, logger *slog.Logger) *SecurityManager {
	sm := &SecurityManager{
		failedCount:    make(map[string]int),
		lockoutTime:    make(map[string]time.Time),
		activeSessions: make(map[string]time.Time),
		logger:         logger,
	}

	go sm.cleanupLoop(ctx)
	return sm
}

func (sm *SecurityManager) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
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

			for token, expiry := range sm.activeSessions {
				if now.Before(expiry) {
					continue
				}

				delete(sm.activeSessions, token)
			}

			sm.mu.Unlock()
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
		sm.lockoutTime[ip] = time.Now().Add(60 * time.Second)
	}
}

func (sm *SecurityManager) RecordSuccessfulAttempt(r *http.Request) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	ip := sm.getIP(r)
	delete(sm.failedCount, ip)
	delete(sm.lockoutTime, ip)
}

func (sm *SecurityManager) GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("security: token generation failed: %w", err)
	}

	token := hex.EncodeToString(b)

	sm.mu.Lock()
	sm.activeSessions[token] = time.Now().Add(24 * time.Hour)
	sm.mu.Unlock()

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

		sm.mu.Lock()
		expiry, exists := sm.activeSessions[cookie.Value]
		if exists && time.Now().After(expiry) {
			delete(sm.activeSessions, cookie.Value)
			exists = false
		}
		sm.mu.Unlock()

		if !exists {
			if r.Method == http.MethodGet {
				http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}

			return
		}

		next(w, r)
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
		next(w, r)
	}
}

func ResolveLocale(r *http.Request) string {
	path := r.URL.Path
	if strings.HasPrefix(path, "/en/") || path == "/en" {
		return "en"
	}
	if strings.HasPrefix(path, "/uk/") || path == "/uk" {
		return "uk"
	}
	if strings.HasPrefix(path, "/ru/") || path == "/ru" {
		return "ru"
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
