package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	Env           string
	DBDSN         string
	AdminUsername string
	AdminPassword string
	Port          string
	LocalesDir    string
}

func Load() (*Config, error) {
	_ = loadEnvFile(".env")

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		if env == "development" {
			dbDSN = "tivri.db"
		} else {
			dbDSN = "postgres://postgres:postgres@localhost:5432/tivri?sslmode=disable"
		}
	}

	adminUsername := os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		adminUsername = "admin"
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = os.Getenv("ADMIN_SECRET")
	}
	if adminPassword == "" {
		adminPassword = "secret"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	localesDir := os.Getenv("LOCALES_DIR")
	if localesDir == "" {
		localesDir = "locales"
	}

	return &Config{
		Env:           env,
		DBDSN:         dbDSN,
		AdminUsername: adminUsername,
		AdminPassword: adminPassword,
		Port:          port,
		LocalesDir:    localesDir,
	}, nil
}

func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

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
			_ = os.Setenv(key, val)
		}
	}

	return scanner.Err()
}
