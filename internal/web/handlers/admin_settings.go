package handlers

import (
	"net/http"
	"time"

	"tivri/internal/eventbus"
	"tivri/internal/web/response"
)

func (h *AdminHandler) HandleAdminSettingsHighQueue(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		response.Error(w, r, err, http.StatusBadRequest, "")
		return
	}
	enabled := r.FormValue("high_queue") == "true" || r.FormValue("high_queue") == "on" || r.FormValue("high_queue") == "1"
	if err := h.settingsRepo.SetHighQueue(r.Context(), enabled); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
		return
	}

	h.eventBus.Publish(r.Context(), eventbus.Event{
		Type:      "settings.high_queue_changed",
		Payload:   enabled,
		Timestamp: time.Now(),
	})
	w.WriteHeader(http.StatusOK)
}

func (h *AdminHandler) HandleAdminSettingsMaintenance(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		response.Error(w, r, err, http.StatusBadRequest, "")
		return
	}
	enabled := r.FormValue("maintenance") == "true" || r.FormValue("maintenance") == "on" || r.FormValue("maintenance") == "1"
	if err := h.settingsRepo.SetMaintenance(r.Context(), enabled); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
		return
	}

	h.eventBus.Publish(r.Context(), eventbus.Event{
		Type:      "settings.maintenance_changed",
		Payload:   enabled,
		Timestamp: time.Now(),
	})
	w.WriteHeader(http.StatusOK)
}
