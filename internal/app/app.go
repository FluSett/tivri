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
	"tivri/internal/features/settings"
	"tivri/internal/i18n"
)

const (
	defaultReadTimeout        = 10 * time.Second
	defaultWriteTimeout       = 10 * time.Second
	defaultIdleTimeout        = 120 * time.Second
	defaultReadHeaderTimeout  = 3 * time.Second
	gracefulShutdown          = 10 * time.Second
	notifyTimeout             = 5 * time.Second
)

type PageData struct {
	CurrentPath       string
	Lang              string
	T                 i18n.Translation
	PortfolioItems    []portfolio.PortfolioItem
	Leads             []project_intake.Lead
	ContactMessages   []messaging.ContactMessage
	LeadsJSON         string
	MessagesJSON      string
	IsAdmin           bool
	IsAdminLogin      bool
	AdminTab          string
	Error             string
	HighQueueActive   bool
	MaintenanceActive bool
	TurnstileSiteKey  string
	AppURL            string
	ContactEmail      string
}

type App struct {
	cfg              *config.Config
	db               *pgxpool.Pool
	translator       *i18n.Translator
	templates        map[string]*template.Template
	portfolioHandler *portfolio.Handler
	leadHandler      *project_intake.Handler
	contactHandler   *messaging.Handler
	settingsRepo     settings.Repository
	logger           *slog.Logger
	webFS            fs.FS
	securityMgr      *security.SecurityManager
	eventBus         eventbus.Bus
	telegramWorker   *notifications.TelegramWorker
}

func New(ctx context.Context) (*App, error) {
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

	if cfg.Env == "production" {
		if cfg.AdminPassword == "" || cfg.AdminPassword == "secret" {
			logger.Warn("admin password is unset or using default — set ADMIN_PASSWORD in production")
		}
		if cfg.TelegramBotToken == "" || cfg.TelegramChatID == "" {
			logger.Warn("telegram notifications disabled — set TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID")
		}
		if cfg.TurnstileSiteKey == "" || cfg.TurnstileSecretKey == "" {
			logger.Warn("turnstile is disabled — set TURNSTILE_SITE_KEY and TURNSTILE_SECRET_KEY")
		}
	}

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

	webUIFS, err := fs.Sub(tivri.WebFS, "web")
	if err != nil {
		db.Close()
		return nil, err
	}

	templates, err := parseTemplates(webUIFS)
	if err != nil {
		db.Close()
		return nil, err
	}

	portfolioHandler := portfolio.NewHandler(portfolioRepo, eventBus, templates["home"])
	leadHandler := project_intake.NewHandler(leadRepo, eventBus, templates["home"], translator, cfg.TurnstileSecretKey)
	contactHandler := messaging.NewHandler(contactRepo, eventBus, templates["home"], translator, cfg.TurnstileSecretKey)
	settingsRepo := settings.NewRepository(db)
	outboxWorker := eventbus.NewOutboxWorker(db, eventBus, logger)
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
	eventBus.Subscribe("system.booted", telegramWorker.HandleEvent)
	eventBus.Subscribe("system.shutdown", telegramWorker.HandleEvent)

	go outboxWorker.Start(ctx)

	securityMgr := security.NewSecurityManager(ctx, logger, db)

	return &App{
		cfg:              cfg,
		db:               db,
		translator:       translator,
		templates:        templates,
		portfolioHandler: portfolioHandler,
		leadHandler:      leadHandler,
		contactHandler:   contactHandler,
		settingsRepo:     settingsRepo,
		logger:           logger,
		webFS:            webUIFS,
		securityMgr:      securityMgr,
		eventBus:         eventBus,
		telegramWorker:   telegramWorker,
	}, nil
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
		Addr:              ":" + a.cfg.Port,
		Handler:           router,
		ReadTimeout:       defaultReadTimeout,
		WriteTimeout:      defaultWriteTimeout,
		IdleTimeout:       defaultIdleTimeout,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
	}

	fmt.Printf("Server starting on %s\n", server.Addr)

	go func() {
		<-ctx.Done()
		a.logger.Info("shutting down server gracefully...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulShutdown)
		defer cancel()

		notifyCtx, notifyCancel := context.WithTimeout(context.Background(), notifyTimeout)
		if err := a.telegramWorker.NotifySystemDown(notifyCtx); err != nil {
			a.logger.Error("failed to send telegram shutdown notification", slog.String("error", err.Error()))
		}
		notifyCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			a.logger.Error("failed to gracefully shut down http server", slog.String("error", err.Error()))
		}
	}()

	a.eventBus.Publish(ctx, eventbus.Event{Type: "system.booted"})

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}


