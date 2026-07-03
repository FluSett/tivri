package app

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"tivri"
)

type TemplateRenderer struct {
	templates map[string]*template.Template
}

func NewTemplateRenderer() (*TemplateRenderer, error) {
	webUIFS, err := fs.Sub(tivri.WebFS, "services/web/ui")
	if err != nil {
		return nil, err
	}

	funcMap := template.FuncMap{
		"formatCents": func(cents int64) string {
			dollars := cents / 100
			remainder := cents % 100
			return fmt.Sprintf("%d.%02d", dollars, remainder)
		},
	}

	homeTmpl, err := template.New("base.layout.html").Funcs(funcMap).ParseFS(
		webUIFS,
		"html/base.layout.html",
		"html/pages/public/home.html",
		"html/partials/portfolio.html",
		"html/partials/notification.html",
		"html/partials/home/about.html",
		"html/partials/home/benefits.html",
		"html/partials/home/skills.html",
		"html/partials/home/portfolio.html",
		"html/partials/home/contact.html",
		"html/partials/home/intake.html",
		"html/partials/home/direct_msg.html",
	)
	if err != nil {
		return nil, err
	}

	adminTmpl, err := template.New("base.layout.html").Funcs(funcMap).ParseFS(
		webUIFS,
		"html/base.layout.html",
		"html/pages/admin/dashboard.html",
		"html/partials/portfolio.html",
		"html/partials/notification.html",
		"html/partials/admin/portfolio.html",
		"html/partials/admin/leads.html",
		"html/partials/admin/messages.html",
	)
	if err != nil {
		return nil, err
	}

	notFoundTmpl, err := template.New("base.layout.html").Funcs(funcMap).ParseFS(
		webUIFS,
		"html/base.layout.html",
		"html/pages/public/404.html",
	)
	if err != nil {
		return nil, err
	}

	loginTmpl, err := template.New("base.layout.html").Funcs(funcMap).ParseFS(
		webUIFS,
		"html/base.layout.html",
		"html/pages/admin/login.html",
	)
	if err != nil {
		return nil, err
	}

	return &TemplateRenderer{
		templates: map[string]*template.Template{
			"home":     homeTmpl,
			"admin":    adminTmpl,
			"notFound": notFoundTmpl,
			"login":    loginTmpl,
		},
	}, nil
}

func (tr *TemplateRenderer) Render(w io.Writer, name string, data interface{}) error {
	tmpl, ok := tr.templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}
	return tmpl.ExecuteTemplate(w, "base.layout.html", data)
}

func (tr *TemplateRenderer) RenderPartial(w io.Writer, templateName, partialName string, data interface{}) error {
	tmpl, ok := tr.templates[templateName]
	if !ok {
		return fmt.Errorf("template %s not found", templateName)
	}
	return tmpl.ExecuteTemplate(w, partialName, data)
}
