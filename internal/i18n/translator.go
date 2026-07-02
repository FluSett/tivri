package i18n

import (
	"encoding/json"
	"io/fs"
	"sync"
	"tivri"
)

type Translation map[string]string

func (t Translation) Get(key string) string {
	if val, ok := t[key]; ok {
		return val
	}
	return key
}

type Translator struct {
	mu           sync.RWMutex
	translations map[string]Translation
}

func NewTranslator() (*Translator, error) {
	t := &Translator{
		translations: make(map[string]Translation),
	}

	langs := []string{"en", "uk", "ru"}

	subFS, err := fs.Sub(tivri.LocalesFS, "locales")
	if err != nil {
		return nil, err
	}

	for _, lang := range langs {
		data, err := fs.ReadFile(subFS, lang+".json")
		if err != nil {
			return nil, err
		}

		var trans Translation
		err = json.Unmarshal(data, &trans)
		if err != nil {
			return nil, err
		}

		t.translations[lang] = trans
	}

	return t, nil
}

func (t *Translator) Get(lang string) Translation {
	t.mu.RLock()
	defer t.mu.RUnlock()

	res, ok := t.translations[lang]
	if !ok {
		return t.translations["en"]
	}

	return res
}
