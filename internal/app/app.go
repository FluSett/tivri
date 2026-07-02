package app

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/jackc/pgx/v5/stdlib"

	"tivri"
	"tivri/internal/config"
	"tivri/internal/domain/contact"
	contactpg "tivri/internal/domain/contact/postgres"
	"tivri/internal/domain/lead"
	leadpg "tivri/internal/domain/lead/postgres"
	"tivri/internal/domain/portfolio"
	portfoliopg "tivri/internal/domain/portfolio/postgres"
	"tivri/internal/i18n"
	webhandler "tivri/services/web/handler"
)

type PageData struct {
	Lang            string
	T               i18n.Translation
	PortfolioItems  []portfolio.PortfolioItem
	Leads           []lead.Lead
	ContactMessages []contact.ContactMessage
	IsAdmin         bool
	IsAdminLogin    bool
	AdminTab        string
	Error           string
}

type App struct {
	cfg              *config.Config
	db               *sql.DB
	translator       *i18n.Translator
	templates        map[string]*template.Template
	portfolioHandler *webhandler.PortfolioHandler
	leadHandler      *webhandler.LeadHandler
	contactHandler   *webhandler.ContactHandler
	portfolioSvc     *portfolio.Service
	leadSvc          *lead.Service
	contactSvc       *contact.Service
	logger           *slog.Logger
	webFS            fs.FS
}

func New() (*App, error) {
	if err := ensureStaticDirectories(); err != nil {
		return nil, err
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	var logger *slog.Logger
	if cfg.Env == "production" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	slog.SetDefault(logger)

	logger.Info("configuration loaded",
		slog.String("env", cfg.Env),
		slog.String("port", cfg.Port),
	)

	if cfg.Env == "development" {
		logger.Info("development admin credentials",
			slog.String("username", cfg.AdminUsername),
			slog.String("password", cfg.AdminPassword),
		)
	}

	driverName := "sqlite"
	if cfg.Env == "production" {
		driverName = "pgx"
	}

	db, err := sql.Open(driverName, cfg.DBDSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = migrate(db, driverName); err != nil {
		db.Close()
		return nil, err
	}

	translator, err := i18n.NewTranslator()
	if err != nil {
		db.Close()
		return nil, err
	}

	portfolioRepo := portfoliopg.NewSQLRepository(db, driverName)
	leadRepo := leadpg.NewSQLRepository(db, driverName)
	contactRepo := contactpg.NewSQLRepository(db, driverName)

	portfolioSvc := portfolio.NewService(portfolioRepo)
	leadSvc := lead.NewService(leadRepo)
	contactSvc := contact.NewService(contactRepo)

	funcMap := template.FuncMap{
		"formatCents": func(cents int64) string {
			dollars := cents / 100
			remainder := cents % 100
			return fmt.Sprintf("%d.%02d", dollars, remainder)
		},
	}

	webUIFS, err := fs.Sub(tivri.WebFS, "services/web/ui")
	if err != nil {
		db.Close()
		return nil, err
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
		db.Close()
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
		db.Close()
		return nil, err
	}

	notFoundTmpl, err := template.New("base.layout.html").Funcs(funcMap).ParseFS(
		webUIFS,
		"html/base.layout.html",
		"html/pages/public/404.html",
	)
	if err != nil {
		db.Close()
		return nil, err
	}

	loginTmpl, err := template.New("base.layout.html").Funcs(funcMap).ParseFS(
		webUIFS,
		"html/base.layout.html",
		"html/pages/admin/login.html",
	)
	if err != nil {
		db.Close()
		return nil, err
	}

	templates := map[string]*template.Template{
		"home":     homeTmpl,
		"admin":    adminTmpl,
		"notFound": notFoundTmpl,
		"login":    loginTmpl,
	}

	portfolioHandler := webhandler.NewPortfolioHandler(portfolioSvc, homeTmpl)
	leadHandler := webhandler.NewLeadHandler(leadSvc, homeTmpl, translator)
	contactHandler := webhandler.NewContactHandler(contactSvc, homeTmpl, translator)

	return &App{
		cfg:              cfg,
		db:               db,
		translator:       translator,
		templates:        templates,
		portfolioHandler: portfolioHandler,
		leadHandler:      leadHandler,
		contactHandler:   contactHandler,
		portfolioSvc:     portfolioSvc,
		leadSvc:          leadSvc,
		contactSvc:       contactSvc,
		logger:           logger,
		webFS:            webUIFS,
	}, nil
}

func (a *App) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

func (a *App) Start() error {
	router, err := a.newRouter()
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:         ":" + a.cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Printf("Server starting on %s\n", server.Addr)
	return server.ListenAndServe()
}

func migrate(db *sql.DB, driverName string) error {
	migrations := map[string]string{
		"pgx":    postgresMigrationSQL,
		"sqlite": sqliteMigrationSQL,
	}

	sqlStr, ok := migrations[driverName]
	if !ok {
		sqlStr = migrations["sqlite"]
	}
	_, err := db.Exec(sqlStr)
	if err != nil {
		return err
	}

	if driverName != "pgx" {
		var count int
		err = db.QueryRow("SELECT count(*) FROM pragma_table_info('portfolio_items') WHERE name='media'").Scan(&count)
		if err != nil {
			return err
		}
		if count == 0 {
			_, err = db.Exec("ALTER TABLE portfolio_items ADD COLUMN media TEXT NOT NULL DEFAULT '[]'")
			if err != nil {
				return err
			}
		}
		err = db.QueryRow("SELECT count(*) FROM pragma_table_info('intake_leads') WHERE name='client_status'").Scan(&count)
		if err != nil {
			return err
		}
		if count == 0 {
			_, err = db.Exec("ALTER TABLE intake_leads ADD COLUMN client_status TEXT NOT NULL DEFAULT 'pending'")
			if err != nil {
				return err
			}
			_, err = db.Exec("ALTER TABLE intake_leads ADD COLUMN internal_status TEXT NOT NULL DEFAULT 'pending'")
			if err != nil {
				return err
			}
		}
		err = db.QueryRow("SELECT count(*) FROM pragma_table_info('intake_leads') WHERE name='updated_at'").Scan(&count)
		if err != nil {
			return err
		}
		if count == 0 {
			_, err = db.Exec("ALTER TABLE intake_leads ADD COLUMN updated_at DATETIME DEFAULT CURRENT_TIMESTAMP")
			if err != nil {
				return err
			}
		}
		err = db.QueryRow("SELECT count(*) FROM pragma_table_info('contact_messages') WHERE name='status'").Scan(&count)
		if err != nil {
			return err
		}
		if count == 0 {
			_, err = db.Exec("ALTER TABLE contact_messages ADD COLUMN status TEXT NOT NULL DEFAULT 'new'")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func ensureStaticDirectories() error {
	base := "services/web/ui/static"

	if _, err := os.Stat(base); os.IsNotExist(err) {
		return nil
	}

	dirs := []string{
		base + "/favicons",
		base + "/img/branding",
		base + "/img/backgrounds",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}

	moves := map[string]string{
		base + "/bg-lg.jpg":   base + "/img/backgrounds/bg-lg.jpg",
		base + "/bg-md.jpg":   base + "/img/backgrounds/bg-md.jpg",
		base + "/bg-sm.jpg":   base + "/img/backgrounds/bg-sm.jpg",
		base + "/logo.png":    base + "/img/branding/logo.png",
		base + "/logo.webp":   base + "/img/branding/logo.webp",
		base + "/favicon.png": base + "/favicons/favicon.png",
	}

	for src, dst := range moves {
		if _, err := os.Stat(src); err == nil {
			err = os.Rename(src, dst)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
