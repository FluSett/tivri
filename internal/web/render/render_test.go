package render

import (
	"io/fs"
	"testing"
	"tivri"
)

func TestNewRenderer(t *testing.T) {
	webUIFS, err := fs.Sub(tivri.WebFS, "web")
	if err != nil {
		t.Fatalf("failed to get sub fs: %v", err)
	}

	renderer, err := NewRenderer(webUIFS)
	if err != nil {
		t.Fatalf("failed to create renderer: %v", err)
	}

	if renderer == nil {
		t.Fatal("expected non-nil renderer")
	}
}
