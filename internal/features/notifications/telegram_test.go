package notifications

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tivri/internal/eventbus"
	"tivri/internal/features/messaging"
	"tivri/internal/features/project_intake"
)

func TestTelegramWorker_HandleEvent(t *testing.T) {
	var receivedPayload map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/botmock-token/sendMessage" {
			t.Errorf("expected url path /botmock-token/sendMessage, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_ = json.NewDecoder(r.Body).Decode(&receivedPayload)

		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	worker := NewTelegramWorker("mock-token", "mock-chat-id")
	worker.apiURL = server.URL

	t.Run("Project Intake Applied", func(t *testing.T) {
		event := eventbus.Event{
			Type: "project_intake.applied",
			Payload: project_intake.ProjectAppliedEvent{
				ID:           42,
				CompanyName:  "Acme Corp",
				ProjectScope: "Build a web app",
				Budget:       500000,
				ContactEmail: "admin@acme.com",
				ContactInfo:  "+123456789",
				Timestamp:    time.Now(),
			},
		}

		err := worker.HandleEvent(context.Background(), event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if receivedPayload["chat_id"] != "mock-chat-id" {
			t.Errorf("expected chat_id mock-chat-id, got %v", receivedPayload["chat_id"])
		}
		if receivedPayload["parse_mode"] != "Markdown" {
			t.Errorf("expected parse_mode Markdown, got %v", receivedPayload["parse_mode"])
		}
		text, ok := receivedPayload["text"].(string)
		if !ok || text == "" {
			t.Errorf("expected text payload, got %v", receivedPayload["text"])
		}
	})

	t.Run("Contact Created", func(t *testing.T) {
		event := eventbus.Event{
			Type: "contact.created",
			Payload: &messaging.ContactMessage{
				ID:        99,
				Email:     "hello@world.com",
				Topic:     "Inquiry",
				Message:   "Hello from the outside",
				Status:    "new",
				CreatedAt: time.Now(),
			},
		}

		err := worker.HandleEvent(context.Background(), event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if receivedPayload["chat_id"] != "mock-chat-id" {
			t.Errorf("expected chat_id mock-chat-id, got %v", receivedPayload["chat_id"])
		}
		if receivedPayload["parse_mode"] != "Markdown" {
			t.Errorf("expected parse_mode Markdown, got %v", receivedPayload["parse_mode"])
		}
		text, ok := receivedPayload["text"].(string)
		if !ok || text == "" {
			t.Errorf("expected text payload, got %v", receivedPayload["text"])
		}
	})

	t.Run("High Queue Changed", func(t *testing.T) {
		event := eventbus.Event{
			Type:      "settings.high_queue_changed",
			Payload:   true,
			Timestamp: time.Now(),
		}

		err := worker.HandleEvent(context.Background(), event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if receivedPayload["chat_id"] != "mock-chat-id" {
			t.Errorf("expected chat_id mock-chat-id, got %v", receivedPayload["chat_id"])
		}
		if receivedPayload["parse_mode"] != "Markdown" {
			t.Errorf("expected parse_mode Markdown, got %v", receivedPayload["parse_mode"])
		}
		text, ok := receivedPayload["text"].(string)
		if !ok || text == "" {
			t.Errorf("expected text payload, got %v", receivedPayload["text"])
		}
	})
}

