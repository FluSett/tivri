package app

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"tivri"
	"tivri/internal/config"
	"tivri/internal/core/database"
	"tivri/internal/core/security"
	"tivri/internal/eventbus"
	"tivri/internal/features/messaging"
	"tivri/internal/features/notifications"
	"tivri/internal/features/portfolio"
	"tivri/internal/features/project_intake"
	"tivri/internal/i18n"
)

type PageData struct {
	Lang            string
	T               i18n.Translation
	PortfolioItems  []portfolio.PortfolioItem
	Leads           []project_intake.Lead
	ContactMessages []messaging.ContactMessage
	LeadsJSON       string
	MessagesJSON    string
	IsAdmin         bool
	IsAdminLogin    bool
	AdminTab        string
	Error           string
	HighQueueActive bool
	MaintenanceActive bool
	TurnstileSiteKey string
}

type App struct {
	cfg              *config.Config
	db               *pgxpool.Pool
	translator       *i18n.Translator
	templates        map[string]*template.Template
	portfolioHandler *portfolio.Handler
	leadHandler      *project_intake.Handler
	contactHandler   *messaging.Handler
	logger           *slog.Logger
	webFS            fs.FS
	securityMgr      *security.SecurityManager
	eventBus         eventbus.Bus
	telegramWorker   *notifications.TelegramWorker
}

func New(ctx context.Context) (*App, error) {
	if err := ensureAssetDirectories(); err != nil {
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

	db, err := database.Connect(ctx, cfg.DBDSN)
	if err != nil {
		return nil, err
	}

	err = database.Migrate(ctx, db, postgresMigrationSQL)
	if err != nil {
		db.Close()
		return nil, err
	}

	translator, err := i18n.NewTranslator()
	if err != nil {
		db.Close()
		return nil, err
	}

	eventBus := eventbus.NewMemoryEventBus(ctx, logger)
	portfolioRepo := portfolio.NewPostgresRepository(db)
	leadRepo := project_intake.NewPostgresRepository(db)
	contactRepo := messaging.NewPostgresRepository(db)

	funcMap := template.FuncMap{
		"formatCents": func(cents int64) string {
			dollars := cents / 100
			remainder := cents % 100
			return fmt.Sprintf("%d.%02d", dollars, remainder)
		},
	}

	webUIFS, err := fs.Sub(tivri.WebFS, "web")
	if err != nil {
		db.Close()
		return nil, err
	}

	homeTmpl, err := template.New("base.layout.html").Funcs(funcMap).ParseFS(
		webUIFS,
		"layouts/base.layout.html",
		"templates/pages/public/home.html",
		"templates/partials/portfolio.html",
		"templates/partials/notification.html",
		"templates/partials/home/about.html",
		"templates/partials/home/benefits.html",
		"templates/partials/home/skills.html",
		"templates/partials/home/portfolio.html",
		"templates/partials/home/contact.html",
		"templates/partials/home/intake.html",
		"templates/partials/home/direct_msg.html",
	)
	if err != nil {
		db.Close()
		return nil, err
	}

	adminTmpl, err := template.New("base.layout.html").Funcs(funcMap).ParseFS(
		webUIFS,
		"layouts/base.layout.html",
		"templates/pages/admin/dashboard.html",
		"templates/partials/portfolio.html",
		"templates/partials/notification.html",
		"templates/partials/admin/portfolio.html",
		"templates/partials/admin/leads.html",
		"templates/partials/admin/messages.html",
	)
	if err != nil {
		db.Close()
		return nil, err
	}

	notFoundTmpl, err := template.New("base.layout.html").Funcs(funcMap).ParseFS(
		webUIFS,
		"layouts/base.layout.html",
		"templates/pages/public/404.html",
	)
	if err != nil {
		db.Close()
		return nil, err
	}

	loginTmpl, err := template.New("base.layout.html").Funcs(funcMap).ParseFS(
		webUIFS,
		"layouts/base.layout.html",
		"templates/pages/admin/login.html",
	)
	if err != nil {
		db.Close()
		return nil, err
	}

	maintenanceTmpl, err := template.New("base.layout.html").Funcs(funcMap).ParseFS(
		webUIFS,
		"layouts/base.layout.html",
		"templates/pages/public/maintenance.html",
	)
	if err != nil {
		db.Close()
		return nil, err
	}

	templates := map[string]*template.Template{
		"home":        homeTmpl,
		"admin":       adminTmpl,
		"notFound":    notFoundTmpl,
		"login":       loginTmpl,
		"maintenance": maintenanceTmpl,
	}

	portfolioHandler := portfolio.NewHandler(portfolioRepo, eventBus, homeTmpl)
	leadHandler := project_intake.NewHandler(leadRepo, eventBus, homeTmpl, translator, cfg.TurnstileSecretKey)
	contactHandler := messaging.NewHandler(contactRepo, eventBus, homeTmpl, translator, cfg.TurnstileSecretKey)
	emailWorker := notifications.NewEmailWorker()
	telegramWorker := notifications.NewTelegramWorker(cfg.TelegramBotToken, cfg.TelegramChatID)

	eventBus.Subscribe("portfolio.created", portfolioHandler.HandlePortfolioCreated)
	eventBus.Subscribe("project_intake.applied", leadHandler.HandleProjectApplied)
	eventBus.Subscribe("project_intake.applied", emailWorker.HandleEvent)
	eventBus.Subscribe("project_intake.applied", telegramWorker.HandleEvent)
	eventBus.Subscribe("contact.created", contactHandler.HandleMessageCreated)
	eventBus.Subscribe("contact.created", emailWorker.HandleEvent)
	eventBus.Subscribe("contact.created", telegramWorker.HandleEvent)
	eventBus.Subscribe("settings.high_queue_changed", telegramWorker.HandleEvent)
	eventBus.Subscribe("settings.maintenance_changed", telegramWorker.HandleEvent)

	securityMgr := security.NewSecurityManager(ctx, logger)

	return &App{
		cfg:              cfg,
		db:               db,
		translator:       translator,
		templates:        templates,
		portfolioHandler: portfolioHandler,
		leadHandler:      leadHandler,
		contactHandler:   contactHandler,
		logger:           logger,
		webFS:            webUIFS,
		securityMgr:      securityMgr,
		eventBus:         eventBus,
		telegramWorker:   telegramWorker,
	}, nil
}

func (a *App) getHighQueueSetting(ctx context.Context) (bool, error) {
	var val string
	err := a.db.QueryRow(ctx, "SELECT value FROM system_settings WHERE key = $1", "high_queue").Scan(&val)
	if err != nil {
		return false, fmt.Errorf("app: get high_queue setting failed: %w", err)
	}
	return val == "true", nil
}

func (a *App) setHighQueueSetting(ctx context.Context, enabled bool) error {
	val := "false"
	if enabled {
		val = "true"
	}
	_, err := a.db.Exec(ctx, "INSERT INTO system_settings (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value", "high_queue", val)
	if err != nil {
		return fmt.Errorf("app: set high_queue setting failed: %w", err)
	}
	return nil
}

func (a *App) getMaintenanceSetting(ctx context.Context) (bool, error) {
	var val string
	err := a.db.QueryRow(ctx, "SELECT value FROM system_settings WHERE key = $1", "maintenance_mode").Scan(&val)
	if err != nil {
		return false, fmt.Errorf("app: get maintenance setting failed: %w", err)
	}
	return val == "true", nil
}

func (a *App) setMaintenanceSetting(ctx context.Context, enabled bool) error {
	val := "false"
	if enabled {
		val = "true"
	}
	_, err := a.db.Exec(ctx, "INSERT INTO system_settings (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value", "maintenance_mode", val)
	if err != nil {
		return fmt.Errorf("app: set maintenance setting failed: %w", err)
	}
	return nil
}

func (a *App) Close() error {
	if a.db != nil {
		a.db.Close()
	}
	return nil
}

func (a *App) Start(ctx context.Context) error {
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

	go func() {
		<-ctx.Done()
		a.logger.Info("shutting down server gracefully...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		notifyCtx, notifyCancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := a.telegramWorker.NotifySystemDown(notifyCtx); err != nil {
			a.logger.Error("failed to send telegram shutdown notification", slog.String("error", err.Error()))
		}
		notifyCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			a.logger.Error("failed to gracefully shut down http server", slog.String("error", err.Error()))
		}
	}()

	bootCtx, bootCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := a.telegramWorker.NotifySystemUp(bootCtx); err != nil {
		a.logger.Error("failed to send telegram boot notification", slog.String("error", err.Error()))
	}
	bootCancel()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func ensureAssetDirectories() error {
	base := "web/assets"
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
