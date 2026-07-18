package app

import (
	"context"
	"fmt"
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
	"tivri/internal/datastore"
	"tivri/internal/eventbus"
	"tivri/internal/i18n"
	"tivri/internal/services"
	"tivri/internal/web/handlers"
	"tivri/internal/web/render"
	"tivri/internal/workers/notifications"
)

const (
	defaultReadTimeout       = 60 * time.Second
	defaultWriteTimeout      = 60 * time.Second
	defaultIdleTimeout       = 120 * time.Second
	defaultReadHeaderTimeout = 10 * time.Second
	gracefulShutdown         = 10 * time.Second
	notifyTimeout            = 5 * time.Second
)

type App struct {
	cfg            *config.Config
	db             *pgxpool.Pool
	logger         *slog.Logger
	webFS          fs.FS
	eventBus       eventbus.Bus
	telegramWorker *notifications.TelegramWorker
	router         http.Handler
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

	db, err := database.Connect(ctx, cfg.DBDSN, cfg.DBMaxConns, cfg.DBMinConns)
	if err != nil {
		return nil, err
	}

	err = database.Migrate(cfg.DBDSN, postgresMigrationFS)
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
	store := datastore.NewStore(db)

	portfolioRepo := datastore.NewPortfolioRepo(store)
	intakeRepo := datastore.NewIntakeRepo(store)
	contactRepo := datastore.NewContactRepo(store)
	settingsRepo := datastore.NewSettingsRepo(store)
	outboxRepo := datastore.NewOutboxRepo(store)

	intakeService := services.NewIntakeService(store, intakeRepo, outboxRepo)
	contactService := services.NewContactService(store, contactRepo, outboxRepo)
	portfolioService := services.NewPortfolioService(portfolioRepo, eventBus)

	webUIFS, err := fs.Sub(tivri.WebFS, "web")
	if err != nil {
		db.Close()
		return nil, err
	}

	render.InitAssets(webUIFS, "assets/dist/manifest.json")
	renderer, err := render.NewRenderer(webUIFS)
	if err != nil {
		db.Close()
		return nil, err
	}

	outboxWorker := eventbus.NewOutboxWorker(db, eventBus, logger)
	emailWorker := notifications.NewEmailWorker()
	telegramWorker := notifications.NewTelegramWorker(cfg.TelegramBotToken, cfg.TelegramChatID)

	eventBus.Subscribe("project_intake.applied", emailWorker.HandleEvent)
	eventBus.Subscribe("project_intake.applied", telegramWorker.HandleEvent)
	eventBus.Subscribe("contact.created", emailWorker.HandleEvent)
	eventBus.Subscribe("contact.created", telegramWorker.HandleEvent)
	eventBus.Subscribe("settings.high_queue_changed", telegramWorker.HandleEvent)
	eventBus.Subscribe("settings.maintenance_changed", telegramWorker.HandleEvent)
	eventBus.Subscribe("system.booted", telegramWorker.HandleEvent)
	eventBus.Subscribe("system.shutdown", telegramWorker.HandleEvent)

	go outboxWorker.Start(ctx)

	securityMgr := security.NewSecurityManager(ctx, logger, db)

	publicHandler := handlers.NewPublicHandler(
		renderer, translator, portfolioRepo, intakeService, contactService,
		settingsRepo, cfg.TurnstileSecretKey, cfg.Env == "production",
	)

	adminHandler := handlers.NewAdminHandler(
		renderer, cfg, securityMgr, portfolioService, portfolioRepo,
		intakeRepo, contactRepo, settingsRepo, eventBus,
	)

	router, err := newRouter(cfg, logger, webUIFS, securityMgr, settingsRepo, translator, renderer, publicHandler, adminHandler)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &App{
		cfg:            cfg,
		db:             db,
		logger:         logger,
		webFS:          webUIFS,
		eventBus:       eventBus,
		telegramWorker: telegramWorker,
		router:         router,
	}, nil
}

func (a *App) Close() error {
	if a.db != nil {
		a.db.Close()
	}
	return nil
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:              ":" + a.cfg.Port,
		Handler:           a.router,
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
