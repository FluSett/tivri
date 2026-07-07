package portfolio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"tivri/internal/eventbus"
)

type HTMLRenderer interface {
	ExecuteTemplate(w io.Writer, name string, data any) error
}

type Handler struct {
	repo        Repository
	bus         eventbus.Bus
	renderer    HTMLRenderer
	mu          sync.RWMutex
	cache       []PortfolioItem
	initialized bool
}

func NewHandler(repo Repository, bus eventbus.Bus, renderer HTMLRenderer) *Handler {
	return &Handler{
		repo:     repo,
		bus:      bus,
		renderer: renderer,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
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
			uploadPaths, err := SaveUploadedFiles(files, "web/assets/uploads")
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
	item := &PortfolioItem{
		Title:       title,
		Description: description,
		Budget:      budget,
		TechStack:   techStack,
		Media:       media,
	}

	err = h.repo.Save(r.Context(), item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.bus.Publish(r.Context(), eventbus.Event{
		Type:      "portfolio.created",
		Payload:   item,
		Timestamp: time.Now(),
	})

	err = h.renderer.ExecuteTemplate(w, "portfolio.html", item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandlePortfolioCreated(ctx context.Context, e eventbus.Event) error {
	item, ok := e.Payload.(*PortfolioItem)
	if !ok {
		return errors.New("portfolio: invalid payload type")
	}

	h.mu.Lock()
	if h.initialized {
		h.cache = append([]PortfolioItem{*item}, h.cache...)
	}
	h.mu.Unlock()

	return nil
}

func (h *Handler) ListItems(ctx context.Context) ([]PortfolioItem, error) {
	h.mu.RLock()
	if h.initialized {
		items := make([]PortfolioItem, len(h.cache))
		copy(items, h.cache)
		h.mu.RUnlock()
		return items, nil
	}
	h.mu.RUnlock()

	h.mu.Lock()
	defer h.mu.Unlock()
	if h.initialized {
		items := make([]PortfolioItem, len(h.cache))
		copy(items, h.cache)
		return items, nil
	}

	items, err := h.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("portfolio: list items failed: %w", err)
	}

	h.cache = items
	h.initialized = true

	copiedItems := make([]PortfolioItem, len(items))
	copy(copiedItems, items)
	return copiedItems, nil
}

func SaveUploadedFiles(files []*multipart.FileHeader, uploadDir string) ([]string, error) {
	var savedPaths []string
	err := os.MkdirAll(uploadDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("portfolio: create upload dir failed: %w", err)
	}

	for _, fileHeader := range files {
		if fileHeader.Size > 5*1024*1024 {
			return nil, fmt.Errorf("portfolio: file exceeds 5MB size: %s", fileHeader.Filename)
		}

		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("portfolio: open uploaded file failed: %w", err)
		}
		defer file.Close()

		ext := filepath.Ext(fileHeader.Filename)
		uniqueName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), "media", ext)
		filePath := filepath.Join(uploadDir, uniqueName)

		out, err := os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("portfolio: create file failed: %w", err)
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			return nil, fmt.Errorf("portfolio: copy file failed: %w", err)
		}

		savedPaths = append(savedPaths, "/assets/uploads/"+uniqueName)
	}

	return savedPaths, nil
}
