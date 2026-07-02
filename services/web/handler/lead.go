package handler

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"tivri/internal/domain/lead"
	"tivri/internal/i18n"
)

type LeadHandler struct {
	service    *lead.Service
	templates  *template.Template
	translator *i18n.Translator
}

func NewLeadHandler(service *lead.Service, templates *template.Template, translator *i18n.Translator) *LeadHandler {
	return &LeadHandler{
		service:    service,
		templates:  templates,
		translator: translator,
	}
}

func (h *LeadHandler) Create(w http.ResponseWriter, r *http.Request) {
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
	} else {
		budget, err = strconv.ParseInt(budgetStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid budget selection", http.StatusBadRequest)
			return
		}
	}

	_, err = h.service.CreateLead(r.Context(), companyName, projectScope, budget, contactEmail, contactPhone)
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

func (h *LeadHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
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

	leads, err := h.service.ListLeads(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var targetLead *lead.Lead
	for _, l := range leads {
		if l.ID == id {
			targetLead = &l
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

	err = h.service.UpdateLeadStatus(r.Context(), id, clientStatus, internalStatus)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
