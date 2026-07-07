package security

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type turnstileResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

func VerifyTurnstile(secretKey, token, remoteIP string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("security: turnstile token is empty")
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	data := url.Values{}
	data.Set("secret", secretKey)
	data.Set("response", token)
	if remoteIP != "" {
		data.Set("remoteip", remoteIP)
	}

	resp, err := client.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", data)
	if err != nil {
		return false, fmt.Errorf("security: siteverify request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("security: siteverify returned non-200 status: %d", resp.StatusCode)
	}

	var result turnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("security: decode siteverify response failed: %w", err)
	}

	return result.Success, nil
}
