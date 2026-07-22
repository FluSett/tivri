package render

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"log/slog"
	"strings"
)

var assetManifest map[string]string

func InitAssets(webFS fs.FS, manifestPath string) {
	assetManifest = make(map[string]string)

	file, err := fs.ReadFile(webFS, manifestPath)
	if err != nil {
		slog.Warn("Could not read asset manifest", "path", manifestPath, "error", err)
		return
	}

	var manifest struct {
		Outputs map[string]struct {
			EntryPoint string `json:"entryPoint"`
		} `json:"outputs"`
	}

	if err := json.Unmarshal(file, &manifest); err != nil {
		slog.Warn("Could not parse asset manifest", "path", manifestPath, "error", err)
		return
	}

	for outPath, info := range manifest.Outputs {
		if info.EntryPoint != "" {
			publicUrl := "/" + strings.TrimPrefix(outPath, "web/")
			if assetData, readErr := fs.ReadFile(webFS, outPath); readErr == nil {
				hash := sha256.Sum256(assetData)
				versionHex := hex.EncodeToString(hash[:4])
				publicUrl += "?v=" + versionHex
			}
			assetManifest[info.EntryPoint] = publicUrl
		}
	}
}

func AssetURL(entryPoint string) string {
	if url, ok := assetManifest[entryPoint]; ok {
		return url
	}
	return "/" + strings.TrimPrefix(entryPoint, "web/")
}
