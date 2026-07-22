package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"tivri/internal/config"
	"tivri/internal/core"
	"tivri/internal/core/security"
	"tivri/internal/eventbus"
	"tivri/internal/services"
	"tivri/internal/web/middleware"
	"tivri/internal/web/render"
	"tivri/internal/web/response"
)

type AdminHandler struct {
	renderer         *render.Renderer
	cfg              *config.Config
	securityMgr      *security.SecurityManager
	portfolioService *services.PortfolioService
	portfolioRepo    core.PortfolioRepository
	intakeRepo       core.LeadRepository
	contactRepo      core.ContactRepository
	settingsRepo     core.SettingsRepository
	eventBus         eventbus.Bus
}

func NewAdminHandler(
	renderer *render.Renderer,
	cfg *config.Config,
	securityMgr *security.SecurityManager,
	portfolioService *services.PortfolioService,
	portfolioRepo core.PortfolioRepository,
	intakeRepo core.LeadRepository,
	contactRepo core.ContactRepository,
	settingsRepo core.SettingsRepository,
	eventBus eventbus.Bus,
) *AdminHandler {
	return &AdminHandler{
		renderer:         renderer,
		cfg:              cfg,
		securityMgr:      securityMgr,
		portfolioService: portfolioService,
		portfolioRepo:    portfolioRepo,
		intakeRepo:       intakeRepo,
		contactRepo:      contactRepo,
		settingsRepo:     settingsRepo,
		eventBus:         eventBus,
	}
}

func (h *AdminHandler) HandleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	baseData := middleware.GetBaseData(r.Context())
	baseData.IsAdmin = true
	baseData.PageTitle = "Admin Dashboard"

	items, err := h.portfolioRepo.List(r.Context())
	if err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
		return
	}

	leadPage, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if leadPage <= 0 {
		leadPage = 1
	}
	leadSort := r.URL.Query().Get("sort_by")
	if leadSort == "" {
		leadSort = "date_desc"
	}
	leadParams := core.LeadListParams{
		Page:           leadPage,
		PageSize:       10,
		SortBy:         leadSort,
		ClientStatus:   r.URL.Query().Get("client_status"),
		InternalStatus: r.URL.Query().Get("internal_status"),
		SearchQuery:    r.URL.Query().Get("search_query"),
	}

	leadsPaginated, err := h.intakeRepo.List(r.Context(), leadParams)
	if err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
		return
	}

	msgPage, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if msgPage <= 0 {
		msgPage = 1
	}
	msgSort := r.URL.Query().Get("sort_by")
	if msgSort == "" {
		msgSort = "date_desc"
	}
	msgParams := core.MessageListParams{
		Page:        msgPage,
		PageSize:    10,
		SortBy:      msgSort,
		Status:      r.URL.Query().Get("status"),
		SearchQuery: r.URL.Query().Get("search_query"),
	}

	msgsPaginated, err := h.contactRepo.List(r.Context(), msgParams)
	if err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
		return
	}

	tab := r.PathValue("tab")
	if tab != "portfolio" && tab != "leads" && tab != "messages" {
		tab = "portfolio"
	}

	var leadsJSON, msgsJSON string
	if raw, err := json.Marshal(leadsPaginated); err == nil {
		leadsJSON = string(raw)
	}
	if raw, err := json.Marshal(msgsPaginated); err == nil {
		msgsJSON = string(raw)
	}

	highQueueActive, _ := h.settingsRepo.GetHighQueue(r.Context())
	maintenanceActive, _ := h.settingsRepo.GetMaintenance(r.Context())

	data := struct {
		render.BaseData
		PortfolioItems    []core.PortfolioItem
		Leads             core.PaginatedLeads
		ContactMessages   core.PaginatedMessages
		LeadsJSON         string
		MessagesJSON      string
		AdminTab          string
		HighQueueActive   bool
		MaintenanceActive bool
	}{
		BaseData:          baseData,
		PortfolioItems:    items,
		Leads:             leadsPaginated,
		ContactMessages:   msgsPaginated,
		LeadsJSON:         leadsJSON,
		MessagesJSON:      msgsJSON,
		AdminTab:          tab,
		HighQueueActive:   highQueueActive,
		MaintenanceActive: maintenanceActive,
	}

	if err := h.renderer.RenderPage(w, "admin", data); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
	}
}
