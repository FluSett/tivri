package handler

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

func SaveUploadedFiles(files []*multipart.FileHeader, uploadDir string) ([]string, error) {
	var savedPaths []string

	err := os.MkdirAll(uploadDir, 0755)
	if err != nil {
		return nil, err
	}

	for _, fileHeader := range files {
		if fileHeader.Size > 5*1024*1024 {
			return nil, fmt.Errorf("file %s exceeds maximum size of 5MB", fileHeader.Filename)
		}

		file, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}
		defer file.Close()

		ext := filepath.Ext(fileHeader.Filename)
		uniqueName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), "media", ext)
		filePath := filepath.Join(uploadDir, uniqueName)

		out, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			return nil, err
		}

		savedPaths = append(savedPaths, "/static/uploads/"+uniqueName)
	}

	return savedPaths, nil
}
