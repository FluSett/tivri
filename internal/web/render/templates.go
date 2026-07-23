package render

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"
	"tivri/internal/core"
	"tivri/internal/i18n"
)

type BaseData struct {
	PageTitle               string
	CurrentPath             string
	Lang                    string
	T                       i18n.Translation
	IsAdmin                 bool
	IsAdminLogin            bool
	TurnstileSiteKey        string
	AppURL                  string
	ContactEmail            string
	Nonce                   string
	CloudflareInsightsToken string
	Error                   string
	CSRFToken               template.HTML
	CSRFTokenVal            string
	MaintenanceActive       bool
	HighQueueActive         bool
}

type Renderer struct {
	templates map[string]*template.Template
}

func NewRenderer(webUIFS fs.FS) (*Renderer, error) {
	assetVer := fmt.Sprintf("%d", time.Now().Unix())
	funcMap := template.FuncMap{
		"assetVersion": func() string {
			if os.Getenv("APP_ENV") == "development" || os.Getenv("APP_ENV") == "" {
				return fmt.Sprintf("%d", time.Now().UnixNano())
			}
			return assetVer
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict expects an even number of arguments")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"list": func(values ...interface{}) []interface{} {
			return values
		},
		"strings_has_suffix": strings.HasSuffix,
		"replaceEmail": func(text, email string) string {
			return strings.ReplaceAll(text, "{{email}}", email)
		},
		"AssetURL":     AssetURL,
		"add":          func(a, b int) int { return a + b },
		"sub":          func(a, b int) int { return a - b },
		"mul":          func(a, b int) int { return a * b },
		"formatBudget": func(val interface{}, _ ...interface{}) string {
			switch v := val.(type) {
			case int64:
				return core.FormatBudget(v)
			case int:
				return core.FormatBudget(int64(v))
			default:
				return "$0"
			}
		},
		"splitTags": func(s string) []string {
			var tags []string
			for _, tag := range strings.Split(s, ",") {
				trimmed := strings.TrimSpace(tag)
				if trimmed != "" {
					tags = append(tags, trimmed)
				}
			}
			return tags
		},
	}

	homeTmpl, err := template.New("base.layout.html").Funcs(funcMap).ParseFS(
		webUIFS,
		"layouts/base.layout.html",
		"templates/partials/layout/*.html",
		"templates/pages/public/home.html",
		"templates/partials/components/portfolio_card.html",
		"templates/partials/components/notification.html",
		"templates/partials/home/about.html",
		"templates/partials/home/benefits.html",
		"templates/partials/home/skills.html",
		"templates/partials/home/portfolio.html",
		"templates/partials/home/contact.html",
		"templates/partials/home/intake.html",
		"templates/partials/home/direct_msg.html",
		"templates/partials/components/section_nav.html",
		"templates/partials/components/lang_switcher.html",
		"templates/partials/components/skill_card.html",
		"templates/partials/components/benefit_card.html",
	)
	if err != nil {
		return nil, err
	}

	adminTmpl, err := template.New("base.layout.html").Funcs(funcMap).ParseFS(
		webUIFS,
		"layouts/base.layout.html",
		"templates/partials/layout/*.html",
		"templates/pages/admin/dashboard.html",
		"templates/partials/components/portfolio_card.html",
		"templates/partials/components/notification.html",
		"templates/partials/admin/portfolio.html",
		"templates/partials/admin/leads.html",
		"templates/partials/admin/messages.html",
		"templates/partials/components/lang_switcher.html",
		"templates/partials/components/dropdown.html",
		"templates/partials/components/pagination.html",
	)
	if err != nil {
		return nil, err
	}

	notFoundTmpl, err := template.New("base.layout.html").Funcs(funcMap).ParseFS(
		webUIFS,
		"layouts/base.layout.html",
		"templates/partials/layout/*.html",
		"templates/pages/public/404.html",
		"templates/partials/components/lang_switcher.html",
	)
	if err != nil {
		return nil, err
	}

	loginTmpl, err := template.New("base.layout.html").Funcs(funcMap).ParseFS(
		webUIFS,
		"layouts/base.layout.html",
		"templates/partials/layout/*.html",
		"templates/pages/admin/login.html",
		"templates/partials/components/lang_switcher.html",
	)
	if err != nil {
		return nil, err
	}

	maintenanceTmpl, err := template.New("base.layout.html").Funcs(funcMap).ParseFS(
		webUIFS,
		"layouts/base.layout.html",
		"templates/partials/layout/*.html",
		"templates/pages/public/maintenance.html",
		"templates/partials/components/lang_switcher.html",
	)
	if err != nil {
		return nil, err
	}

	privacyTmpl, err := template.New("base.layout.html").Funcs(funcMap).ParseFS(
		webUIFS,
		"layouts/base.layout.html",
		"templates/partials/layout/*.html",
		"templates/pages/public/privacy.html",
		"templates/partials/components/lang_switcher.html",
		"templates/partials/components/benefit_card.html",
	)
	if err != nil {
		return nil, err
	}

	termsTmpl, err := template.New("base.layout.html").Funcs(funcMap).ParseFS(
		webUIFS,
		"layouts/base.layout.html",
		"templates/partials/layout/*.html",
		"templates/pages/public/terms.html",
		"templates/partials/components/lang_switcher.html",
		"templates/partials/components/benefit_card.html",
	)
	if err != nil {
		return nil, err
	}

	robotsTmpl, err := template.New("robots.txt").ParseFS(
		webUIFS,
		"templates/pages/public/robots.txt",
	)
	if err != nil {
		return nil, err
	}

	sitemapTmpl, err := template.New("sitemap.xml").ParseFS(
		webUIFS,
		"templates/pages/public/sitemap.xml",
	)
	if err != nil {
		return nil, err
	}

	templates := map[string]*template.Template{
		"home":        homeTmpl,
		"admin":       adminTmpl,
		"notFound":    notFoundTmpl,
		"login":       loginTmpl,
		"maintenance": maintenanceTmpl,
		"privacy":     privacyTmpl,
		"terms":       termsTmpl,
		"robots":      robotsTmpl,
		"sitemap":     sitemapTmpl,
	}

	return &Renderer{templates: templates}, nil
}

func (r *Renderer) RenderPage(w io.Writer, templateName string, data any) error {
	tmpl, ok := r.templates[templateName]
	if !ok {
		return fmt.Errorf("template %s not found", templateName)
	}

	return tmpl.ExecuteTemplate(w, "base.layout.html", data)
}

func (r *Renderer) RenderRaw(w io.Writer, templateName string, data any) error {
	tmpl, ok := r.templates[templateName]
	if !ok {
		return fmt.Errorf("template %s not found", templateName)
	}
	return tmpl.Execute(w, data)
}

func (r *Renderer) RenderPartial(w io.Writer, bundleName, templateName string, data any) error {
	tmpl, ok := r.templates[bundleName]
	if !ok {
		return fmt.Errorf("bundle %s not found", bundleName)
	}
	return tmpl.ExecuteTemplate(w, templateName, data)
}
