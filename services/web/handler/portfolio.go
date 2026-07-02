package handler

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"tivri/internal/domain/portfolio"
)

type PortfolioHandler struct {
	service   *portfolio.Service
	templates *template.Template
}

func NewPortfolioHandler(service *portfolio.Service, templates *template.Template) *PortfolioHandler {
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

		for _, fileHeader := range files {
			if fileHeader.Size > 5*1024*1024 {
				http.Error(w, "File "+fileHeader.Filename+" exceeds maximum size of 5MB", http.StatusBadRequest)
				return
			}

			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer file.Close()

			uploadDir := "services/web/ui/static/uploads"

			err = os.MkdirAll(uploadDir, 0755)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			ext := filepath.Ext(fileHeader.Filename)
			uniqueName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), "media", ext)
			filePath := filepath.Join(uploadDir, uniqueName)

			out, err := os.Create(filePath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer out.Close()

			_, err = io.Copy(out, file)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			media = append(media, "/static/uploads/"+uniqueName)
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
