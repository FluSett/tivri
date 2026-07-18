package render

import (
	"encoding/json"
	"io/fs"
	"log"
	"strings"
)

var assetManifest map[string]string

func InitAssets(webFS fs.FS, manifestPath string) {
	assetManifest = make(map[string]string)

	file, err := fs.ReadFile(webFS, manifestPath)
	if err != nil {
		log.Printf("Warning: Could not read asset manifest %s: %v", manifestPath, err)
		return
	}

	var manifest struct {
		Outputs map[string]struct {
			EntryPoint string `json:"entryPoint"`
		} `json:"outputs"`
	}

	if err := json.Unmarshal(file, &manifest); err != nil {
		log.Printf("Warning: Could not parse asset manifest %s: %v", manifestPath, err)
		return
	}

	for outPath, info := range manifest.Outputs {
		if info.EntryPoint != "" {
			publicUrl := "/" + strings.TrimPrefix(outPath, "web/")
			assetManifest[info.EntryPoint] = publicUrl
		}
	}
}

func AssetURL(entryPoint string) string {
	if url, ok := assetManifest[entryPoint]; ok {
		return url
	}
	// Fallback if not found in manifest
	return "/" + strings.TrimPrefix(entryPoint, "web/")
}
