package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"tivri/internal/core"
	"tivri/internal/eventbus"
)

const (
	telegramAPITimeout = 5 * time.Second
	retryBackoffBase   = 500 * time.Millisecond
)

type TelegramWorker struct {
	token  string
	chatID string
	client *http.Client
	apiURL string
}

func NewTelegramWorker(token, chatID string) *TelegramWorker {
	return &TelegramWorker{
		token:  token,
		chatID: chatID,
		client: &http.Client{
			Timeout: telegramAPITimeout,
		},
		apiURL: "https://api.telegram.org",
	}
}

func (w *TelegramWorker) HandleEvent(ctx context.Context, e eventbus.Event) error {
	if w.token == "" || w.chatID == "" {
		return nil
	}

	var message string

	switch e.Type {
	case "project_intake.applied":
		var evt core.ProjectAppliedEventPayload
		switch p := e.Payload.(type) {
		case core.ProjectAppliedEventPayload:
			evt = p
		case []byte:
			if err := json.Unmarshal(p, &evt); err != nil {
				return fmt.Errorf("notifications/telegram: unmarshal ProjectAppliedEventPayload failed: %w", err)
			}
		default:
			return errors.New("notifications/telegram: invalid project intake payload type")
		}

		var priorityStr string
		if evt.DeadlineNeeded {
			priorityStr = fmt.Sprintf("Yes (Details: %s)", escapeMarkdown(evt.DeadlineSpec))
		} else {
			priorityStr = "No (Standard Queue)"
		}

		message = fmt.Sprintf(
			"🚀 *New Project Intake Received*\n\n"+
				"*ID:* %d\n"+
				"*Company:* %s\n"+
				"*Email:* %s\n"+
				"*Budget:* %s\n"+
				"*Priority Requested:* %s\n"+
				"*Additional Contact/Notes:* %s\n\n"+
				"*Scope:*\n%s",
			evt.ID,
			escapeMarkdown(evt.CompanyName),
			escapeMarkdown(evt.ContactEmail),
			core.FormatBudgetTier(evt.Budget, evt.IsCustomBudget),
			priorityStr,
			escapeMarkdown(evt.ContactInfo),
			escapeMarkdown(evt.ProjectScope),
		)

	case "contact.created":
		var msg core.ContactMessage
		switch p := e.Payload.(type) {
		case *core.ContactMessage:
			if p != nil {
				msg = *p
			}
		case core.ContactMessage:
			msg = p
		case []byte:
			if err := json.Unmarshal(p, &msg); err != nil {
				return fmt.Errorf("notifications/telegram: unmarshal ContactMessage failed: %w", err)
			}
		default:
			return errors.New("notifications/telegram: invalid contact message payload type")
		}

		message = fmt.Sprintf(
			"✉️ *New Direct Message Received*\n\n"+
				"*ID:* %d\n"+
				"*Email:* %s\n"+
				"*Topic:* %s\n\n"+
				"*Message:*\n%s",
			msg.ID,
			escapeMarkdown(msg.Email),
			escapeMarkdown(msg.Topic),
			escapeMarkdown(msg.Message),
		)

	case "settings.high_queue_changed":
		enabled, ok := e.Payload.(bool)
		if !ok {
			return errors.New("notifications/telegram: invalid high-queue status payload type")
		}

		statusStr := "DISABLED"
		statusEmoji := "🟢"
		if enabled {
			statusStr = "ENABLED"
			statusEmoji = "🔴"
		}

		message = fmt.Sprintf(
			"%s *System Alert: High Queue Status Changed*\n\n"+
				"High Queue Mode has been set to *%s* by an administrator.",
			statusEmoji,
			statusStr,
		)

	case "settings.maintenance_changed":
		enabled, ok := e.Payload.(bool)
		if !ok {
			return errors.New("notifications/telegram: invalid maintenance status payload type")
		}

		statusStr := "DISABLED"
		statusEmoji := "🟢"
		if enabled {
			statusStr = "ENABLED"
			statusEmoji = "🛠️"
		}

		message = fmt.Sprintf(
			"%s *System Alert: Maintenance Mode Changed*\n\n"+
				"Maintenance Mode has been set to *%s* by an administrator.",
			statusEmoji,
			statusStr,
		)

	case "system.booted":
		return w.NotifySystemUp(ctx)

	case "system.shutdown":
		return w.NotifySystemDown(ctx)

	default:
		return nil
	}
	return w.sendTelegramMessage(ctx, message)
}

func (w *TelegramWorker) NotifySystemUp(ctx context.Context) error {
	if w.token == "" || w.chatID == "" {
		return nil
	}

	message := fmt.Sprintf("✅ *Server Status Update*\n\nServer has booted successfully at `%s`.", time.Now().Format(time.RFC1123))
	return w.sendTelegramMessage(ctx, message)
}

func (w *TelegramWorker) NotifySystemDown(ctx context.Context) error {
	if w.token == "" || w.chatID == "" {
		return nil
	}

	message := fmt.Sprintf("⚠️ *Server Status Update*\n\nServer is shutting down gracefully at `%s`.", time.Now().Format(time.RFC1123))
	return w.sendTelegramMessage(ctx, message)
}

func (w *TelegramWorker) sendTelegramMessage(ctx context.Context, text string) error {
	url := fmt.Sprintf("%s/bot%s/sendMessage", w.apiURL, w.token)
	payload := map[string]interface{}{
		"chat_id":    w.chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("notifications/telegram: marshal payload failed: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		reqCtx, cancel := context.WithTimeout(ctx, telegramAPITimeout)
		req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, bytes.NewBuffer(body))
		if err != nil {
			cancel()
			return fmt.Errorf("notifications/telegram: create request failed: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		resp, err := w.client.Do(req)
		cancel()

		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
			var errResp struct {
				Description string `json:"description"`
			}
			if json.NewDecoder(resp.Body).Decode(&errResp) == nil && errResp.Description != "" {
				lastErr = fmt.Errorf("telegram api returned status %d: %s", resp.StatusCode, errResp.Description)
			} else {
				lastErr = fmt.Errorf("telegram api returned status: %d", resp.StatusCode)
			}
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt) * retryBackoffBase):
		}
	}
	return fmt.Errorf("notifications/telegram: send message failed after 3 attempts: %w", lastErr)
}

func escapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"*", "\\*",
		"_", "\\_",
		"`", "\\`",
		"[", "\\[",
	)
	return replacer.Replace(text)
}
