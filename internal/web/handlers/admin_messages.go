package handlers

import (
	"net/http"
	"strconv"

	"tivri/internal/core"
	"tivri/internal/web/middleware"
	"tivri/internal/web/render"
	"tivri/internal/web/response"
)

func (h *AdminHandler) HandleAdminMessagesPartial(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	params := core.MessageListParams{
		Page:        page,
		PageSize:    10,
		SortBy:      r.URL.Query().Get("sort_by"),
		Status:      r.URL.Query().Get("status"),
		SearchQuery: r.URL.Query().Get("search_query"),
	}

	msgsPaginated, err := h.contactRepo.List(r.Context(), params)
	if err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
		return
	}

	data := struct {
		render.BaseData
		ContactMessages core.PaginatedMessages
	}{
		BaseData:        middleware.GetBaseData(r.Context()),
		ContactMessages: msgsPaginated,
	}
	data.BaseData.IsAdmin = true

	qMsg := r.URL.Query().Encode()
	if qMsg != "" {
		w.Header().Set("HX-Replace-Url", "/admin/messages?"+qMsg)
	} else {
		w.Header().Set("HX-Replace-Url", "/admin/messages")
	}

	if err := h.renderer.RenderPartial(w, "admin", "admin.messages.html", data); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
	}
}

func (h *AdminHandler) HandleContactUpdateStatus(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		response.Error(w, r, err, http.StatusBadRequest, "")
		return
	}

	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err != nil {
		response.Error(w, r, nil, http.StatusBadRequest, "Invalid message ID")
		return
	}

	if err := h.contactRepo.UpdateStatus(r.Context(), id, r.FormValue("status")); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
		return
	}

	w.WriteHeader(http.StatusOK)
}
