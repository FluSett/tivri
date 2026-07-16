package config

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Env                string
	DBDSN              string
	AdminUsername      string
	AdminPassword      string
	Port               string
	LocalesDir         string
	TurnstileSiteKey   string
	TurnstileSecretKey string
	TelegramBotToken   string
	TelegramChatID     string
	AppURL             string
	ContactEmail       string
}

func getEnv(key string) string {
	return strings.TrimSpace(strings.ReplaceAll(os.Getenv(key), "\r", ""))
}

func Load() (*Config, error) {
	if err := loadEnvFile(".env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("config: load .env failed: %w", err)
	}

	env := getEnv("APP_ENV")
	if env == "" {
		env = "development"
	}

	dbDSN := getEnv("DB_DSN")
	if dbDSN == "" {
		if env == "development" {
			dbDSN = "tivri.db"
		} else {
			dbDSN = "postgres://postgres:postgres@localhost:5432/tivri?sslmode=disable"
		}
	}

	adminUsername := getEnv("ADMIN_USERNAME")
	if adminUsername == "" {
		adminUsername = "admin"
	}

	adminPassword := getEnv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "secret"
	}

	port := getEnv("PORT")
	if port == "" {
		port = "8080"
	}

	localesDir := getEnv("LOCALES_DIR")
	if localesDir == "" {
		localesDir = "locales"
	}

	turnstileSiteKey := getEnv("TURNSTILE_SITE_KEY")
	turnstileSecretKey := getEnv("TURNSTILE_SECRET_KEY")
	telegramBotToken := getEnv("TELEGRAM_BOT_TOKEN")
	telegramChatID := getEnv("TELEGRAM_CHAT_ID")

	appURL := getEnv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:8080"
	}

	contactEmail := getEnv("CONTACT_EMAIL")
	if contactEmail == "" {
		contactEmail = "contact@tivri.cc"
	}

	return &Config{
		Env:                env,
		DBDSN:              dbDSN,
		AdminUsername:      adminUsername,
		AdminPassword:      adminPassword,
		Port:               port,
		LocalesDir:         localesDir,
		TurnstileSiteKey:   turnstileSiteKey,
		TurnstileSecretKey: turnstileSecretKey,
		TelegramBotToken:   telegramBotToken,
		TelegramChatID:     telegramChatID,
		AppURL:             appURL,
		ContactEmail:       contactEmail,
	}, nil
}

func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	slog.Debug("config: loading environment overrides", slog.String("file", filename))

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key != "" {
			if err := os.Setenv(key, val); err != nil {
				return fmt.Errorf("config: setenv %q failed: %w", key, err)
			}
		}
	}

	return scanner.Err()
}
