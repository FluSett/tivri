package project_intake

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tivri/internal/eventbus"
	"tivri/internal/i18n"
)

type HTMLRenderer interface {
	ExecuteTemplate(w io.Writer, name string, data any) error
}

type Handler struct {
	repo       Repository
	bus        eventbus.Bus
	renderer   HTMLRenderer
	translator *i18n.Translator
}

func NewHandler(repo Repository, bus eventbus.Bus, renderer HTMLRenderer, translator *i18n.Translator) *Handler {
	return &Handler{
		repo:       repo,
		bus:        bus,
		renderer:   renderer,
		translator: translator,
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

	companyName := strings.TrimSpace(r.FormValue("company_name"))
	projectScope := strings.TrimSpace(r.FormValue("project_scope"))
	budgetStr := r.FormValue("budget")
	contactEmail := strings.TrimSpace(r.FormValue("contact_email"))
	contactPhone := strings.TrimSpace(r.FormValue("contact_phone"))

	if len(companyName) < 2 || len(companyName) > 150 {
		http.Error(w, "Name/company must be between 2 and 150 characters", http.StatusBadRequest)
		return
	}

	if len(projectScope) < 20 || len(projectScope) > 2000 {
		http.Error(w, "Project scope must be between 20 and 2000 characters", http.StatusBadRequest)
		return
	}

	if len(contactEmail) < 5 || len(contactEmail) > 254 || !strings.Contains(contactEmail, "@") {
		http.Error(w, "Invalid contact email address", http.StatusBadRequest)
		return
	}

	var budget int64
	if budgetStr == "other" {
		customBudgetStr := strings.TrimSpace(r.FormValue("custom_budget"))
		budget, err = strconv.ParseInt(customBudgetStr, 10, 64)
		if err != nil || budget < 100 {
			http.Error(w, "Invalid custom budget value (must be at least 100 USD)", http.StatusBadRequest)
			return
		}
		budget = budget * 100
	} else {
		budget, err = strconv.ParseInt(budgetStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid budget selection", http.StatusBadRequest)
			return
		}
	}

	ld := &Lead{
		CompanyName:    companyName,
		ProjectScope:   projectScope,
		Budget:         budget,
		ContactEmail:   contactEmail,
		ContactPhone:   contactPhone,
		ClientStatus:   "pending",
		InternalStatus: "pending",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err = h.repo.Save(r.Context(), ld)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.bus.Publish(r.Context(), eventbus.Event{
		Type: "project_intake.applied",
		Payload: ProjectAppliedEvent{
			ID:           ld.ID,
			CompanyName:  ld.CompanyName,
			ProjectScope: ld.ProjectScope,
			Budget:       ld.Budget,
			ContactEmail: ld.ContactEmail,
			ContactPhone: ld.ContactPhone,
			Timestamp:    time.Now(),
		},
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
	statusType := r.FormValue("type")
	status := r.FormValue("status")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid lead ID", http.StatusBadRequest)
		return
	}

	if status != "pending" && status != "active" && status != "done" && status != "canceled" {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	leads, err := h.repo.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var targetLead *Lead
	for i := range leads {
		if leads[i].ID == id {
			targetLead = &leads[i]
			break
		}
	}

	if targetLead == nil {
		http.Error(w, "Lead not found", http.StatusNotFound)
		return
	}

	clientStatus := targetLead.ClientStatus
	internalStatus := targetLead.InternalStatus

	switch statusType {
	case "client":
		clientStatus = status
	case "internal":
		internalStatus = status
	default:
		http.Error(w, "Invalid status type", http.StatusBadRequest)
		return
	}

	err = h.repo.UpdateStatus(r.Context(), id, clientStatus, internalStatus)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleProjectApplied(ctx context.Context, e eventbus.Event) error {
	evt, ok := e.Payload.(ProjectAppliedEvent)
	if !ok {
		return errors.New("invalid payload type")
	}

	fmt.Printf("Notification subscriber: email dispatched to %s regarding lead ID %d\n", evt.ContactEmail, evt.ID)
	return nil
}

func (h *Handler) ListLeads(ctx context.Context) ([]Lead, error) {
	return h.repo.List(ctx)
}
