package handler

import (
	"net/http"
	"strconv"
	"tivri/internal/domain/portfolio"
)

type PortfolioHandler struct {
	service   *portfolio.Service
	templates HTMLRenderer
}

func NewPortfolioHandler(service *portfolio.Service, templates HTMLRenderer) *PortfolioHandler {
	return &PortfolioHandler{
		service:   service,
		templates: templates,
	}
}

func (h *PortfolioHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		err = r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	techStack := r.FormValue("tech_stack")

	var media []string
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		files := r.MultipartForm.File["media"]
		if len(files) > 0 {
			uploadPaths, err := SaveUploadedFiles(files, "services/web/ui/static/uploads")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			media = uploadPaths
		}
	}

	var budgetVal int64
	budgetStr := r.FormValue("budget")

	if budgetStr != "" {
		if val, err := strconv.ParseInt(budgetStr, 10, 64); err == nil {
			budgetVal = val
		}
	}

	budget := budgetVal * 100

	item, err := h.service.CreatePortfolioItem(r.Context(), title, description, budget, techStack, media)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.templates.ExecuteTemplate(w, "portfolio.html", item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PortfolioHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	items, err := h.service.ListPortfolioItems(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, item := range items {
		err = h.templates.ExecuteTemplate(w, "portfolio.html", item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
