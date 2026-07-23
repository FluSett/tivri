package handlers

import (
	"net/http"
	"strconv"

	"tivri/internal/core"
	"tivri/internal/web/middleware"
	"tivri/internal/web/render"
	"tivri/internal/web/response"
)

func (h *AdminHandler) HandleAdminLeadsPartial(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	params := core.LeadListParams{
		Page:           page,
		PageSize:       10,
		SortBy:         r.URL.Query().Get("sort_by"),
		ClientStatus:   r.URL.Query().Get("client_status"),
		InternalStatus: r.URL.Query().Get("internal_status"),
		ServiceType:    r.URL.Query().Get("service_type"),
		SearchQuery:    r.URL.Query().Get("search_query"),
	}

	leadsPaginated, err := h.intakeRepo.List(r.Context(), params)
	if err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
		return
	}

	data := struct {
		render.BaseData
		Leads core.PaginatedLeads
	}{
		BaseData: middleware.GetBaseData(r.Context()),
		Leads:    leadsPaginated,
	}
	data.BaseData.IsAdmin = true

	q := r.URL.Query().Encode()
	if q != "" {
		w.Header().Set("HX-Replace-Url", "/admin/leads?"+q)
	} else {
		w.Header().Set("HX-Replace-Url", "/admin/leads")
	}

	if err := h.renderer.RenderPartial(w, "admin", "admin.leads.html", data); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
	}
}

func (h *AdminHandler) HandleLeadUpdateStatus(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		response.Error(w, r, err, http.StatusBadRequest, "")
		return
	}

	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err != nil {
		response.Error(w, r, nil, http.StatusBadRequest, "Invalid lead ID")
		return
	}

	statusType := r.FormValue("type")
	status := r.FormValue("status")

	targetLead, err := h.intakeRepo.Get(r.Context(), id)
	if err != nil {
		response.Error(w, r, nil, http.StatusNotFound, "Lead not found")
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
		response.Error(w, r, nil, http.StatusBadRequest, "Invalid status type")
		return
	}

	if err := h.intakeRepo.UpdateStatus(r.Context(), id, clientStatus, internalStatus); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
		return
	}

	w.WriteHeader(http.StatusOK)
}
