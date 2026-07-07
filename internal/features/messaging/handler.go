package messaging

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tivri/internal/core/security"
	"tivri/internal/eventbus"
	"tivri/internal/i18n"
)

type HTMLRenderer interface {
	ExecuteTemplate(w io.Writer, name string, data any) error
}

type Handler struct {
	repo            Repository
	bus             eventbus.Bus
	renderer        HTMLRenderer
	translator      *i18n.Translator
	turnstileSecret string
}

func NewHandler(repo Repository, bus eventbus.Bus, renderer HTMLRenderer, translator *i18n.Translator, turnstileSecret string) *Handler {
	return &Handler{
		repo:            repo,
		bus:             bus,
		renderer:        renderer,
		translator:      translator,
		turnstileSecret: turnstileSecret,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if h.turnstileSecret != "" {
		token := r.FormValue("cf-turnstile-response")
		lang := r.FormValue("lang")
		trans := h.translator.Get(lang)
		if token == "" {
			http.Error(w, trans.Get("ValTurnstileRequired"), http.StatusBadRequest)
			return
		}

		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			if strings.Contains(r.RemoteAddr, ":") {
				if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
					ip = host
				}
			}
			if ip == "" {
				ip = r.RemoteAddr
			}
		}

		ok, err := security.VerifyTurnstile(h.turnstileSecret, token, ip)
		if err != nil || !ok {
			http.Error(w, trans.Get("ValTurnstileFailed"), http.StatusBadRequest)
			return
		}
	}

	email := strings.TrimSpace(r.FormValue("email"))
	topic := strings.TrimSpace(r.FormValue("topic"))
	message := strings.TrimSpace(r.FormValue("message"))

	if len(email) < 5 || len(email) > 100 || !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		http.Error(w, "Invalid email address structure", http.StatusBadRequest)
		return
	}

	if len(topic) < 3 || len(topic) > 150 {
		http.Error(w, "Topic length must be between 3 and 150 characters", http.StatusBadRequest)
		return
	}

	if len(message) < 10 || len(message) > 1000 {
		http.Error(w, "Message length must be between 10 and 1000 characters", http.StatusBadRequest)
		return
	}

	msg := &ContactMessage{
		Email:     email,
		Topic:     topic,
		Message:   message,
		Status:    "new",
		CreatedAt: time.Now(),
	}

	err = h.repo.Save(r.Context(), msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.bus.Publish(r.Context(), eventbus.Event{
		Type:      "contact.created",
		Payload:   msg,
		Timestamp: time.Now(),
	})

	lang := r.FormValue("lang")
	trans := h.translator.Get(lang)
	data := struct {
		Message string
	}{
		Message: trans.Get("SuccessMsg"),
	}

	err = h.renderer.ExecuteTemplate(w, "notification.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("id")
	status := r.FormValue("status")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	if status != "new" && status != "answered" && status != "done" {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	err = h.repo.UpdateStatus(r.Context(), id, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleMessageCreated(ctx context.Context, e eventbus.Event) error {
	msg, ok := e.Payload.(*ContactMessage)
	if !ok {
		return errors.New("invalid payload type")
	}

	fmt.Printf("Notification subscriber: contact message from %s regarding %q received\n", msg.Email, msg.Topic)
	return nil
}

func (h *Handler) ListMessages(ctx context.Context) ([]ContactMessage, error) {
	return h.repo.List(ctx)
}
