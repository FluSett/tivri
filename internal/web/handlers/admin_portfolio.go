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
