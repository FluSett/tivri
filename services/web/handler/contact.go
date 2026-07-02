package handler

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"tivri/internal/domain/contact"
	"tivri/internal/i18n"
)

type ContactHandler struct {
	service    *contact.Service
	templates  *template.Template
	translator *i18n.Translator
}

func NewContactHandler(service *contact.Service, templates *template.Template, translator *i18n.Translator) *ContactHandler {
	return &ContactHandler{
		service:    service,
		templates:  templates,
		translator: translator,
	}
}

func (h *ContactHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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

	_, err = h.service.CreateMessage(r.Context(), email, topic, message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lang := r.FormValue("lang")
	trans := h.translator.Get(lang)
	data := struct {
		Message string
	}{
		Message: trans.Get("SuccessMsg"),
	}

	err = h.templates.ExecuteTemplate(w, "notification.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *ContactHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
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

	err = h.service.UpdateMessageStatus(r.Context(), id, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
