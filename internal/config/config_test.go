package config

import (
	"strings"
	"testing"
	"time"
)

func validConfig() *Config {
	return &Config{
		Port:                  "8080",
		Env:                   "production",
		AllowedOrigins:        []string{"https://app.rawdah.app"},
		DatabaseURL:           "postgres://example",
		AutoMigrate:           true,
		JWTAccessSecret:       strings.Repeat("a", 64),
		JWTRefreshSecret:      strings.Repeat("b", 64),
		AccessTokenTTL:        15 * time.Minute,
		RefreshTokenTTL:       7 * 24 * time.Hour,
		ChildTokenTTL:         4 * time.Hour,
		PresignExpiresSeconds: 600,
	}
}

func TestValidateConfigAcceptsSecureConfig(t *testing.T) {
	if err := validate(validConfig()); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
}

func TestValidateConfigRejectsShortJWTSecrets(t *testing.T) {
	cfg := validConfig()
	cfg.JWTAccessSecret = "short"
	cfg.JWTRefreshSecret = "short"

	if err := validate(cfg); err == nil {
		t.Fatalf("expected validation error for short secrets")
	}
}

func TestValidateConfigRequiresAllowedOriginsInProduction(t *testing.T) {
	cfg := validConfig()
	cfg.AllowedOrigins = nil

	if err := validate(cfg); err == nil {
		t.Fatalf("expected validation error for missing origins in production")
	}
}
