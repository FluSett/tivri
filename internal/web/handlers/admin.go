package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/deepteams/webp"

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

func (h *AdminHandler) HandleAdminLogin(w http.ResponseWriter, r *http.Request) {
	baseData := middleware.GetBaseData(r.Context())
	baseData.IsAdmin = true
	baseData.IsAdminLogin = true
	baseData.PageTitle = "Admin Login"

	if r.Method == http.MethodGet {
		data := struct{ render.BaseData }{BaseData: baseData}
		if err := h.renderer.RenderPage(w, "login", data); err != nil {
			response.Error(w, r, err, http.StatusInternalServerError, "")
		}
		return
	}

	if r.Method == http.MethodPost {
		if h.securityMgr.IsLockedOut(r) {
			baseData.Error = "Too many failed attempts. Locked out for 60 seconds."
			data := struct{ render.BaseData }{BaseData: baseData}
			w.WriteHeader(http.StatusTooManyRequests)
			if err := h.renderer.RenderPage(w, "login", data); err != nil {
				response.Error(w, r, err, http.StatusInternalServerError, "")
			}
			return
		}

		if h.cfg.TurnstileSiteKey != "" {
			token := r.FormValue("cf-turnstile-response")
			var ip string
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				parts := strings.Split(forwarded, ",")
				ip = strings.TrimSpace(parts[0])
			} else {
				if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
					ip = host
				} else {
					ip = r.RemoteAddr
				}
			}
			ok, err := security.VerifyTurnstile(h.cfg.TurnstileSecretKey, token, ip)
			if err != nil || !ok {
				baseData.Error = baseData.T.Get("ValTurnstileFailed")
				data := struct{ render.BaseData }{BaseData: baseData}
				w.WriteHeader(http.StatusBadRequest)
				if err = h.renderer.RenderPage(w, "login", data); err != nil {
					response.Error(w, r, err, http.StatusInternalServerError, "")
				}
				return
			}
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		userHash := sha256.Sum256([]byte(username))
		cfgUserHash := sha256.Sum256([]byte(h.cfg.AdminUsername))
		passHash := sha256.Sum256([]byte(password))
		cfgPassHash := sha256.Sum256([]byte(h.cfg.AdminPassword))

		userMatch := subtle.ConstantTimeCompare(userHash[:], cfgUserHash[:]) == 1
		passMatch := subtle.ConstantTimeCompare(passHash[:], cfgPassHash[:]) == 1

		if !userMatch || !passMatch {
			h.securityMgr.RecordFailedAttempt(r)
			baseData.Error = "Invalid username or password"
			data := struct{ render.BaseData }{BaseData: baseData}
			w.WriteHeader(http.StatusUnauthorized)
			if err := h.renderer.RenderPage(w, "login", data); err != nil {
				response.Error(w, r, err, http.StatusInternalServerError, "")
			}
			return
		}

		h.securityMgr.RecordSuccessfulAttempt(r)
		token, err := h.securityMgr.GenerateToken(r.Context())
		if err != nil {
			response.Error(w, r, err, http.StatusInternalServerError, "")
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "admin_session",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   h.cfg.Env == "production",
			SameSite: http.SameSiteStrictMode,
			MaxAge:   86400,
		})

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

func (h *AdminHandler) HandleAdminLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.Env == "production",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
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

	leadsPaginated, err := h.intakeRepo.List(r.Context(), core.LeadListParams{Page: 1, PageSize: 10, SortBy: "date_desc"})
	if err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
		return
	}

	msgsPaginated, err := h.contactRepo.List(r.Context(), core.MessageListParams{Page: 1, PageSize: 10, SortBy: "date_desc"})
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

	if err := h.renderer.RenderPartial(w, "admin", "admin.leads.html", data); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
	}
}

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

	if err := h.renderer.RenderPartial(w, "admin", "admin.messages.html", data); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
	}
}

func (h *AdminHandler) HandlePortfolioCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		if err := r.ParseForm(); err != nil {
			response.Error(w, r, err, http.StatusBadRequest, "")
			return
		}
	}

	var media []string
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if files := r.MultipartForm.File["media"]; len(files) > 0 {
			uploadPaths, err := saveUploadedFiles(files, "web/assets/uploads")
			if err != nil {
				response.Error(w, r, err, http.StatusBadRequest, "")
				return
			}
			media = uploadPaths
		}
	}

	budget, _ := strconv.ParseInt(r.FormValue("budget"), 10, 64)

	item := &core.PortfolioItem{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		Budget:      budget * 100,
		TechStack:   r.FormValue("tech_stack"),
		Media:       media,
	}

	if err := h.portfolioService.SaveItem(r.Context(), item); err != nil {
		response.Error(w, r, err, http.StatusInternalServerError, "")
		return
	}

	if err := h.renderer.RenderPartial(w, "home", "components.portfolio_card.html", item); err != nil {
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

func convertToWebP(src io.Reader, dst io.Writer) error {
	img, _, err := image.Decode(src)
	if err != nil {
		return err
	}
	return webp.Encode(dst, img, &webp.EncoderOptions{Quality: 85})
}

func saveUploadedFiles(files []*multipart.FileHeader, uploadDir string) ([]string, error) {
	var savedPaths []string
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("portfolio: create upload dir failed: %w", err)
	}

	for _, fileHeader := range files {
		if fileHeader.Size > 5*1024*1024 {
			return nil, fmt.Errorf("portfolio: file exceeds 5MB")
		}

		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("portfolio: open uploaded file failed: %w", err)
		}

		contentType := fileHeader.Header.Get("Content-Type")
		isImage := strings.HasPrefix(contentType, "image/")
		isVideo := strings.HasPrefix(contentType, "video/")

		if !isImage && !isVideo {
			file.Close()
			return nil, fmt.Errorf("portfolio: invalid file type")
		}

		isWebP := contentType == "image/webp" || strings.ToLower(filepath.Ext(fileHeader.Filename)) == ".webp"

		randBytes := make([]byte, 4)
		var hexSuffix string
		if _, randErr := rand.Read(randBytes); randErr == nil {
			hexSuffix = hex.EncodeToString(randBytes)
		} else {
			hexSuffix = fmt.Sprintf("%d", time.Now().UnixNano())
		}

		uniqueName := fmt.Sprintf("%d_media_%s%s", time.Now().UnixNano(), hexSuffix, filepath.Ext(fileHeader.Filename))
		if isImage && !isWebP {
			uniqueName = fmt.Sprintf("%d_media_%s.webp", time.Now().UnixNano(), hexSuffix)
		}
		filePath := filepath.Join(uploadDir, uniqueName)

		out, err := os.Create(filePath)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("portfolio: create file failed: %w", err)
		}

		if isImage && !isWebP {
			err = convertToWebP(file, out)
			if err != nil {
				out.Close()
				os.Remove(filePath)
				file.Close()
				return nil, fmt.Errorf("portfolio: webp conversion failed: %w", err)
			}
		} else {
			if _, err = io.Copy(out, file); err != nil {
				out.Close()
				os.Remove(filePath)
				file.Close()
				return nil, fmt.Errorf("portfolio: copy file failed: %w", err)
			}
		}

		file.Close()
		out.Close()
		savedPaths = append(savedPaths, "/assets/uploads/"+uniqueName)
	}

	return savedPaths, nil
}
