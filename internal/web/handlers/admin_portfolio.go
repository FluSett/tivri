package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/deepteams/webp"

	"tivri/internal/core"
	"tivri/internal/web/response"
)

func (h *AdminHandler) HandlePortfolioCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		if err := r.ParseForm(); err != nil {
			slog.Warn("portfolio: parse multipart form failed", "error", err)
			response.Error(w, r, err, http.StatusBadRequest, "Failed to parse uploaded form")
			return
		}
	}

	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	if title == "" || description == "" {
		slog.Warn("portfolio: title or description empty")
		response.Error(w, r, fmt.Errorf("title and description are required"), http.StatusBadRequest, "Title and description are required")
		return
	}

	var media []string
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if files := r.MultipartForm.File["media"]; len(files) > 0 {
			uploadPaths, err := saveUploadedFiles(files, "web/assets/uploads")
			if err != nil {
				slog.Warn("portfolio: file upload failed", "error", err)
				response.Error(w, r, err, http.StatusBadRequest, err.Error())
				return
			}
			media = uploadPaths
		}
	}

	budget, _ := strconv.ParseInt(r.FormValue("budget"), 10, 64)

	item := &core.PortfolioItem{
		Title:       title,
		Description: description,
		Budget:      budget * 100,
		TechStack:   r.FormValue("tech_stack"),
		Media:       media,
	}

	if err := h.portfolioService.SaveItem(r.Context(), item); err != nil {
		slog.Error("portfolio: save item failed", "error", err)
		response.Error(w, r, err, http.StatusInternalServerError, "Failed to save portfolio item")
		return
	}

	if err := h.renderer.RenderPartial(w, "home", "components.portfolio_card.html", item); err != nil {
		slog.Error("portfolio: render partial failed", "error", err)
		response.Error(w, r, err, http.StatusInternalServerError, "")
	}
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
		if fileHeader.Size > 25*1024*1024 {
			return nil, fmt.Errorf("file %s exceeds 25MB limit", fileHeader.Filename)
		}

		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("open uploaded file failed: %w", err)
		}

		contentType := fileHeader.Header.Get("Content-Type")
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		isImage := strings.HasPrefix(contentType, "image/") || ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp" || ext == ".gif" || ext == ".svg"
		isVideo := strings.HasPrefix(contentType, "video/") || ext == ".mp4" || ext == ".webm" || ext == ".mov" || ext == ".mkv" || ext == ".avi"

		if !isImage && !isVideo {
			file.Close()
			return nil, fmt.Errorf("unsupported file format for %s", fileHeader.Filename)
		}

		isWebP := contentType == "image/webp" || ext == ".webp"

		randBytes := make([]byte, 4)
		var hexSuffix string
		if _, randErr := rand.Read(randBytes); randErr == nil {
			hexSuffix = hex.EncodeToString(randBytes)
		} else {
			hexSuffix = fmt.Sprintf("%d", time.Now().UnixNano())
		}

		uniqueName := fmt.Sprintf("%d_media_%s%s", time.Now().UnixNano(), hexSuffix, ext)
		if isImage && !isWebP {
			uniqueName = fmt.Sprintf("%d_media_%s.webp", time.Now().UnixNano(), hexSuffix)
		}
		filePath := filepath.Join(uploadDir, uniqueName)

		out, err := os.Create(filePath)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("create destination file failed: %w", err)
		}

		if isImage && !isWebP {
			err = convertToWebP(file, out)
			if err != nil {
				slog.Warn("portfolio: webp conversion failed, falling back to original copy", "filename", fileHeader.Filename, "error", err)
				out.Close()
				os.Remove(filePath)
				if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
					file.Close()
					return nil, fmt.Errorf("failed to seek uploaded file: %w", seekErr)
				}
				origName := fmt.Sprintf("%d_media_%s%s", time.Now().UnixNano(), hexSuffix, ext)
				origPath := filepath.Join(uploadDir, origName)
				origOut, createErr := os.Create(origPath)
				if createErr != nil {
					file.Close()
					return nil, fmt.Errorf("create original file failed: %w", createErr)
				}
				if _, copyErr := io.Copy(origOut, file); copyErr != nil {
					origOut.Close()
					os.Remove(origPath)
					file.Close()
					return nil, fmt.Errorf("copy original file failed: %w", copyErr)
				}
				origOut.Close()
				file.Close()
				savedPaths = append(savedPaths, "/assets/uploads/"+origName)
				continue
			}
		} else {
			if _, err = io.Copy(out, file); err != nil {
				out.Close()
				os.Remove(filePath)
				file.Close()
				return nil, fmt.Errorf("copy file failed: %w", err)
			}
		}

		file.Close()
		out.Close()
		savedPaths = append(savedPaths, "/assets/uploads/"+uniqueName)
	}

	return savedPaths, nil
}
