package config

import (
	"os"
	"testing"
)

func TestConfigLoad_Defaults(t *testing.T) {
	os.Unsetenv("APP_ENV")
	os.Unsetenv("DB_DSN")
	os.Unsetenv("PORT")
	os.Unsetenv("ADMIN_USERNAME")
	os.Unsetenv("ADMIN_PASSWORD")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Env != "development" {
		t.Errorf("expected Env 'development', got %s", cfg.Env)
	}

	if cfg.Port != "8080" {
		t.Errorf("expected Port '8080', got %s", cfg.Port)
	}

	if cfg.AdminUsername != "admin" {
		t.Errorf("expected default admin username 'admin', got %s", cfg.AdminUsername)
	}

	if cfg.DBDSN != "postgres://postgres:postgres@localhost:5432/tivri?sslmode=disable" {
		t.Errorf("expected default DB DSN postgres local url, got %s", cfg.DBDSN)
	}
}

func TestConfigLoad_Overrides(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	os.Setenv("DB_DSN", "postgres://test_user:pass@remote:5432/db")
	os.Setenv("PORT", "9000")
	os.Setenv("ADMIN_USERNAME", "custom_admin")
	os.Setenv("ADMIN_PASSWORD", "custom_password")

	defer func() {
		os.Unsetenv("APP_ENV")
		os.Unsetenv("DB_DSN")
		os.Unsetenv("PORT")
		os.Unsetenv("ADMIN_USERNAME")
		os.Unsetenv("ADMIN_PASSWORD")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Env != "production" {
		t.Errorf("expected overridden Env 'production', got %s", cfg.Env)
	}

	if cfg.Port != "9000" {
		t.Errorf("expected overridden Port '9000', got %s", cfg.Port)
	}

	if cfg.AdminUsername != "custom_admin" {
		t.Errorf("expected overridden admin username 'custom_admin', got %s", cfg.AdminUsername)
	}

	if cfg.DBDSN != "postgres://test_user:pass@remote:5432/db" {
		t.Errorf("expected overridden DB DSN, got %s", cfg.DBDSN)
	}
}
