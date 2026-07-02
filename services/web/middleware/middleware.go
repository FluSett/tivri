package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	lockoutMu      sync.Mutex
	failedCount    = make(map[string]int)
	lockoutTime    = make(map[string]time.Time)
	activeSessions = make(map[string]time.Time)
)

func init() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			lockoutMu.Lock()
			now := time.Now()
			for ip, until := range lockoutTime {
				if now.Before(until) {
					continue
				}
				delete(lockoutTime, ip)
				delete(failedCount, ip)
			}
			for token, expiry := range activeSessions {
				if now.Before(expiry) {
					continue
				}
				delete(activeSessions, token)
			}
			lockoutMu.Unlock()
		}
	}()
}

func getIP(r *http.Request) string {
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

func IsLockedOut(r *http.Request) bool {
	lockoutMu.Lock()
	defer lockoutMu.Unlock()
	ip := getIP(r)
	until, ok := lockoutTime[ip]
	if ok && time.Now().Before(until) {
		return true
	}
	return false
}

func RecordFailedAttempt(r *http.Request) {
	lockoutMu.Lock()
	defer lockoutMu.Unlock()
	ip := getIP(r)
	failedCount[ip]++
	if failedCount[ip] >= 3 {
		lockoutTime[ip] = time.Now().Add(60 * time.Second)
	}
}

func RecordSuccessfulAttempt(r *http.Request) {
	lockoutMu.Lock()
	defer lockoutMu.Unlock()
	ip := getIP(r)
	delete(failedCount, ip)
	delete(lockoutTime, ip)
}

func GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	lockoutMu.Lock()
	activeSessions[token] = time.Now().Add(24 * time.Hour)
	lockoutMu.Unlock()
	return token, nil
}

func CookieAuth(adminUsername, adminPassword string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if IsLockedOut(r) {
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
		lockoutMu.Lock()
		expiry, exists := activeSessions[cookie.Value]
		if exists && time.Now().After(expiry) {
			delete(activeSessions, cookie.Value)
			exists = false
		}
		lockoutMu.Unlock()
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

func BasicAuth(adminUsername, adminPassword string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if IsLockedOut(r) {
			w.WriteHeader(http.StatusTooManyRequests)
			if _, err := w.Write([]byte("Too many failed login attempts. Please try again later.")); err != nil {
				slog.Error("failed to write response", slog.Any("error", err))
			}
			return
		}
		username, password, ok := r.BasicAuth()
		if !ok || username != adminUsername || password != adminPassword {
			RecordFailedAttempt(r)
			w.Header().Set("WWW-Authenticate", `Basic realm="Admin"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		RecordSuccessfulAttempt(r)
		next(w, r)
	}
}

func ResolveLocale(r *http.Request) string {
	cookie, err := r.Cookie("lang")
	if err == nil {
		lang := cookie.Value
		if lang == "en" || lang == "uk" || lang == "ru" {
			return lang
		}
	}
	qLang := r.URL.Query().Get("lang")
	if qLang == "en" || qLang == "uk" || qLang == "ru" {
		return qLang
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
