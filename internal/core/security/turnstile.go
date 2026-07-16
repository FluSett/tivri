package security

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"
)

const turnstileAPITimeout = 5 * time.Second

type turnstileResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

func VerifyTurnstile(secretKey, token, remoteIP string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("security: turnstile token is empty")
	}

	client := &http.Client{
		Timeout: turnstileAPITimeout,
	}

	data := url.Values{}
	data.Set("secret", secretKey)
	data.Set("response", token)

	if remoteIP != "" {
		if ip := net.ParseIP(remoteIP); ip != nil {
			if !ip.IsLoopback() && !ip.IsPrivate() && !ip.IsUnspecified() {
				data.Set("remoteip", remoteIP)
			}
		}
	}

	resp, err := client.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", data)
	if err != nil {
		slog.Error("Turnstile request error", slog.String("error", err.Error()))
		return false, fmt.Errorf("security: siteverify request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		slog.Error("Turnstile returned non-200 status", slog.Int("status", resp.StatusCode), slog.String("body", string(bodyBytes)))
		return false, fmt.Errorf("security: siteverify returned non-200 status: %d", resp.StatusCode)
	}

	var result turnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Error("Turnstile decode response failed", slog.String("error", err.Error()))
		return false, fmt.Errorf("security: decode siteverify response failed: %w", err)
	}

	if !result.Success {
		slog.Warn("Turnstile verification failed", slog.Any("error-codes", result.ErrorCodes))
	}

	return result.Success, nil
}
