package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// This prevents leakage of internal Go stack traces, database states, or paths.
func Error(w http.ResponseWriter, r *http.Request, err error, statusCode int, userMsg string) {
	if err != nil {
		slog.Error("request failed",
			slog.String("path", r.URL.Path),
			slog.String("method", r.Method),
			slog.String("error", err.Error()),
			slog.Int("status", statusCode),
		)
	}

	if userMsg == "" {
		if statusCode >= 500 {
			userMsg = "An internal system error occurred."
		} else {
			userMsg = http.StatusText(statusCode)
		}
	}

	http.Error(w, userMsg, statusCode)
}

func JSONError(w http.ResponseWriter, r *http.Request, err error, statusCode int, userMsg string) {
	if err != nil {
		slog.Error("api request failed",
			slog.String("path", r.URL.Path),
			slog.String("method", r.Method),
			slog.String("error", err.Error()),
			slog.Int("status", statusCode),
		)
	}

	if userMsg == "" {
		if statusCode >= 500 {
			userMsg = "An internal system error occurred."
		} else {
			userMsg = http.StatusText(statusCode)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	payload := map[string]string{"error": userMsg}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("failed to encode error response", slog.String("error", err.Error()))
	}
}
