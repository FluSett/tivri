package app

import (
	"net/http"
	"time"

	"tivri/internal/core/security"
)

func (a *App) handleRobots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	data := struct{ AppURL string }{AppURL: a.cfg.AppURL}
	err := a.templates["robots"].Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleSitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	now := time.Now().Format("2006-01-02")
	data := struct {
		Now    string
		AppURL string
		Langs  []string
		Pages  []string
	}{
		Now:    now,
		AppURL: a.cfg.AppURL,
		Langs:  []string{"", "?lang=en", "?lang=uk", "?lang=ru"},
		Pages:  []string{"privacy", "terms"},
	}
	
	err := a.templates["sitemap"].Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handlePrivacy(w http.ResponseWriter, r *http.Request) {
	lang := security.ResolveLocale(r)
	data := PageData{
		CurrentPath:      "/privacy",
		Lang:             lang,
		T:                a.translator.Get(lang),
		TurnstileSiteKey: a.cfg.TurnstileSiteKey,
		AppURL:           a.cfg.AppURL,
		ContactEmail:     a.cfg.ContactEmail,
	}
	err := a.templates["privacy"].ExecuteTemplate(w, "base.layout.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleTerms(w http.ResponseWriter, r *http.Request) {
	lang := security.ResolveLocale(r)
	data := PageData{
		CurrentPath:      "/terms",
		Lang:             lang,
		T:                a.translator.Get(lang),
		TurnstileSiteKey: a.cfg.TurnstileSiteKey,
		AppURL:           a.cfg.AppURL,
		ContactEmail:     a.cfg.ContactEmail,
	}
	err := a.templates["terms"].ExecuteTemplate(w, "base.layout.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	if p != "/" {
		w.WriteHeader(http.StatusNotFound)
		lang := security.ResolveLocale(r)
		data := PageData{
			CurrentPath:      r.URL.Path,
			Lang:             lang,
			T:                a.translator.Get(lang),
			TurnstileSiteKey: a.cfg.TurnstileSiteKey,
			AppURL:           a.cfg.AppURL,
			ContactEmail:     a.cfg.ContactEmail,
		}

		err := a.templates["notFound"].ExecuteTemplate(w, "base.layout.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	lang := security.ResolveLocale(r)

	items, err := a.portfolioHandler.ListItems(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	highQueueActive, _ := a.settingsRepo.GetHighQueue(r.Context())
	data := PageData{
		CurrentPath:      "/",
		Lang:             lang,
		T:                a.translator.Get(lang),
		PortfolioItems:   items,
		HighQueueActive:  highQueueActive,
		TurnstileSiteKey: a.cfg.TurnstileSiteKey,
		AppURL:           a.cfg.AppURL,
		ContactEmail:     a.cfg.ContactEmail,
	}

	err = a.templates["home"].ExecuteTemplate(w, "base.layout.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
