package i18n

import (
	"testing"
)

func TestNewTranslator(t *testing.T) {
	trans, err := NewTranslator()
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}

	langs := []string{"en", "uk", "ru"}
	for _, lang := range langs {
		t.Run("check language "+lang, func(t *testing.T) {
			locale := trans.Get(lang)
			if locale == nil {
				t.Fatalf("locale %s is nil", lang)
			}

			val := locale.Get("MetaTitle")
			if val == "" || val == "MetaTitle" {
				t.Errorf("missing MetaTitle translation for %s, got %s", lang, val)
			}
		})
	}
}

func TestTranslation_GetFallback(t *testing.T) {
	transMap := Translation{
		"key_exists": "Value",
	}

	if val := transMap.Get("key_exists"); val != "Value" {
		t.Errorf("expected 'Value', got %s", val)
	}

	if val := transMap.Get("non_existent_key"); val != "non_existent_key" {
		t.Errorf("expected key fallback name 'non_existent_key', got %s", val)
	}
}
