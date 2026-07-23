package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tivri/internal/core"
	"tivri/internal/core/security"
	"tivri/internal/i18n"
	"tivri/internal/services"
	"tivri/internal/web/middleware"
	"tivri/internal/web/render"
	"tivri/internal/web/response"
)

type databasePinger interface {
	Ping(ctx context.Context) error
}

type PublicHandler struct {
	renderer        *render.Renderer
	translator      *i18n.Translator
	portfolioRepo   core.PortfolioRepository
	intakeService   *services.IntakeService
	contactService  *services.ContactService
	settingsRepo    core.SettingsRepository
	dbPinger        databasePinger
	turnstileSecret string
	isProd          bool
}

func NewPublicHandler(
	renderer *render.Renderer,
	translator *i18n.Translator,
	portfolioRepo core.PortfolioRepository,
	intakeService *services.IntakeService,
	contactService *services.ContactService,
	settingsRepo core.SettingsRepository,
	dbPinger databasePinger,
	turnstileSecret string,
	isProd bool,
) *PublicHandler {
	return &PublicHandler{
		renderer:        renderer,
		translator:      translator,
		portfolioRepo:   portfolioRepo,
		intakeService:   intakeService,
		contactService:  contactService,
		settingsRepo:    settingsRepo,
		dbPinger:        dbPinger,
		turnstileSecret: turnstileSecret,
		isProd:          isProd,
	}
}

func (h *PublicHandler) HandleAPILang(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	if lang != "en" && lang != "uk" && lang != "ru" {
		lang = "en"
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "lang",
		Value:    lang,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.isProd,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   31536000,
	})

	currentURL := r.Header.Get("HX-Current-URL")
	if currentURL == "" {
		currentURL = "/"
	}

	w.Header().Set("HX-Location", currentURL)
	w.WriteHeader(http.StatusOK)
}

func (h *PublicHandler) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.dbPinger != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := h.dbPinger.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"error","message":"database unreachable"}`))
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok","database":"healthy"}`))
}

func (h *PublicHandler) HandleRobots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	baseData := middleware.GetBaseData(r.Context())
	data := struct{ AppURL string }{AppURL: baseData.AppURL}
	if err := h.renderer.RenderRaw(w, "robots", data); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
	}
}

func (h *PublicHandler) HandleSitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	baseData := middleware.GetBaseData(r.Context())
	now := time.Now().Format("2006-01-02")
	data := struct {
		Now    string
		AppURL string
		Langs  []string
		Pages  []string
	}{
		Now:    now,
		AppURL: baseData.AppURL,
		Langs:  []string{"", "?lang=en", "?lang=uk", "?lang=ru"},
		Pages:  []string{"privacy", "terms"},
	}

	if err := h.renderer.RenderRaw(w, "sitemap", data); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
	}
}

func (h *PublicHandler) HandlePrivacy(w http.ResponseWriter, r *http.Request) {
	baseData := middleware.GetBaseData(r.Context())
	baseData.PageTitle = baseData.T.Get("PrivacyTitle")
	data := struct{ render.BaseData }{BaseData: baseData}
	if err := h.renderer.RenderPage(w, "privacy", data); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
	}
}

func (h *PublicHandler) HandleTerms(w http.ResponseWriter, r *http.Request) {
	baseData := middleware.GetBaseData(r.Context())
	baseData.PageTitle = baseData.T.Get("TermsTitle")
	data := struct{ render.BaseData }{BaseData: baseData}
	if err := h.renderer.RenderPage(w, "terms", data); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
	}
}

func (h *PublicHandler) HandleHome(w http.ResponseWriter, r *http.Request) {
	baseData := middleware.GetBaseData(r.Context())

	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		baseData.PageTitle = "404 Not Found"
		data := struct{ render.BaseData }{BaseData: baseData}
		if err := h.renderer.RenderPage(w, "notFound", data); err != nil {
			response.Error(w, r, err, http.StatusInternalServerError, "")
		}
		return
	}

	items, err := h.portfolioRepo.List(r.Context())
	if err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
		return
	}

	highQueueActive, _ := h.settingsRepo.GetHighQueue(r.Context())

	data := struct {
		render.BaseData
		PortfolioItems  []core.PortfolioItem
		HighQueueActive bool
	}{
		BaseData:        baseData,
		PortfolioItems:  items,
		HighQueueActive: highQueueActive,
	}

	if err := h.renderer.RenderPage(w, "home", data); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
	}
}

func (h *PublicHandler) HandleIntakeCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		response.Error(w, r, err, http.StatusBadRequest, "")
		return
	}

	if h.turnstileSecret != "" {
		if ok, err := security.ValidateTurnstileRequest(r, h.turnstileSecret); err != nil || !ok {
			lang := r.FormValue("lang")
			trans := h.translator.Get(lang)
			response.Error(w, r, nil, http.StatusBadRequest, trans.Get("ValTurnstileFailed"))
			return
		}
	}

	companyName := core.SanitizeString(r.FormValue("company_name"))
	serviceType := core.SanitizeString(r.FormValue("service_type"))
	if serviceType == "" {
		serviceType = "full_project"
	}
	projectScope := core.SanitizeString(r.FormValue("project_scope"))
	existingURL := core.SanitizeString(r.FormValue("existing_url"))
	techStack := core.SanitizeString(r.FormValue("tech_stack"))
	contactEmail := core.SanitizeString(r.FormValue("contact_email"))
	contactInfo := core.SanitizeString(r.FormValue("contact_info"))
	deadlineNeededStr := r.FormValue("deadline_needed")
	deadlineNeeded := deadlineNeededStr == "true" || deadlineNeededStr == "on" || deadlineNeededStr == "1"

	budgetStr := strings.TrimSpace(r.FormValue("budget"))
	budget, err := strconv.ParseInt(budgetStr, 10, 64)
	if err != nil || budget < 5 {
		response.Error(w, r, nil, http.StatusBadRequest, "Invalid budget amount (minimum $5 USD)")
		return
	}

	ld := &core.Lead{
		CompanyName:    companyName,
		ServiceType:    serviceType,
		ProjectScope:   projectScope,
		ExistingURL:    existingURL,
		TechStack:      techStack,
		Budget:         budget,
		ContactEmail:   contactEmail,
		ContactInfo:    contactInfo,
		DeadlineNeeded: deadlineNeeded,
		DeadlineSpec:   core.SanitizeString(r.FormValue("deadline_spec")),
		ClientStatus:   "pending",
		InternalStatus: "pending",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := h.intakeService.Apply(r.Context(), ld); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
		return
	}

	trans := h.translator.Get(r.FormValue("lang"))
	if err := h.renderer.RenderPartial(w, "home", "components.notification.html", struct{ Message string }{Message: trans.Get("SuccessMsg")}); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
	}
}

func (h *PublicHandler) HandleContactCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		response.Error(w, r, err, http.StatusBadRequest, "")
		return
	}

	if h.turnstileSecret != "" {
		if ok, err := security.ValidateTurnstileRequest(r, h.turnstileSecret); err != nil || !ok {
			lang := r.FormValue("lang")
			trans := h.translator.Get(lang)
			response.Error(w, r, nil, http.StatusBadRequest, trans.Get("ValTurnstileFailed"))
			return
		}
	}

	msg := &core.ContactMessage{
		Email:     core.SanitizeString(r.FormValue("email")),
		Topic:     core.SanitizeString(r.FormValue("topic")),
		Message:   core.SanitizeString(r.FormValue("message")),
		Status:    "new",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.contactService.SendMessage(r.Context(), msg); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
		return
	}

	trans := h.translator.Get(r.FormValue("lang"))
	if err := h.renderer.RenderPartial(w, "home", "components.notification.html", struct{ Message string }{Message: trans.Get("SuccessMsg")}); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
	}
}
