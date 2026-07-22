package config

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env                     string
	DBDSN                   string
	DBMaxConns              int32
	DBMinConns              int32
	AdminUsername           string
	AdminPassword           string
	Port                    string
	LocalesDir              string
	TurnstileSiteKey        string
	TurnstileSecretKey      string
	TelegramBotToken        string
	TelegramChatID          string
	AppURL                  string
	ContactEmail            string
	CloudflareInsightsToken string
	CSRFAuthKey             []byte
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
		dbDSN = "postgres://postgres:postgres@localhost:5432/tivri?sslmode=disable"
	}

	adminUsername := getEnv("ADMIN_USERNAME")
	if adminUsername == "" {
		adminUsername = "admin"
	}

	adminPassword := getEnv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "secret"
	}

	dbMaxConnsStr := getEnv("DB_MAX_CONNS")
	dbMaxConns := int32(25)
	if val, err := strconv.ParseInt(dbMaxConnsStr, 10, 32); err == nil {
		dbMaxConns = int32(val)
	}

	dbMinConnsStr := getEnv("DB_MIN_CONNS")
	dbMinConns := int32(5)
	if val, err := strconv.ParseInt(dbMinConnsStr, 10, 32); err == nil {
		dbMinConns = int32(val)
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

	cloudflareInsightsToken := getEnv("CLOUDFLARE_INSIGHTS_TOKEN")

	csrfAuthKeyStr := getEnv("CSRF_AUTH_KEY")
	var csrfAuthKey []byte
	if csrfAuthKeyStr != "" {
		csrfAuthKey = []byte(csrfAuthKeyStr)
		if len(csrfAuthKey) != 32 {
			slog.Warn("CSRF_AUTH_KEY must be exactly 32 bytes, CSRF protection may fail or panic")
		}
	} else {
		csrfAuthKey = []byte("01234567890123456789012345678901")
		if env == "production" {
			slog.Warn("CSRF_AUTH_KEY is not set in production! Using insecure fallback key.")
		}
	}

	cfg := &Config{
		Env:                     env,
		DBDSN:                   dbDSN,
		DBMaxConns:              dbMaxConns,
		DBMinConns:              dbMinConns,
		AdminUsername:           adminUsername,
		AdminPassword:           adminPassword,
		Port:                    port,
		LocalesDir:              localesDir,
		TurnstileSiteKey:        turnstileSiteKey,
		TurnstileSecretKey:      turnstileSecretKey,
		TelegramBotToken:        telegramBotToken,
		TelegramChatID:          telegramChatID,
		AppURL:                  appURL,
		ContactEmail:            contactEmail,
		CloudflareInsightsToken: cloudflareInsightsToken,
		CSRFAuthKey:             csrfAuthKey,
	}

	if err := cfg.Validate(); err != nil {
		slog.Warn("config validation warning", "error", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.DBDSN == "" {
		return errors.New("config: DB_DSN cannot be empty")
	}
	if len(c.CSRFAuthKey) != 32 {
		return errors.New("config: CSRF_AUTH_KEY must be exactly 32 bytes")
	}
	if c.Env == "production" {
		if c.AdminPassword == "" || c.AdminPassword == "secret" {
			return errors.New("config: ADMIN_PASSWORD must be changed from default in production")
		}
	}
	return nil
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
