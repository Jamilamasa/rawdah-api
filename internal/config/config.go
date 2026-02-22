package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Port           string
	Env            string
	AllowedOrigins []string

	DatabaseURL string
	AutoMigrate bool

	JWTAccessSecret  string
	JWTRefreshSecret string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	ChildTokenTTL    time.Duration

	R2AccountID           string
	R2AccessKeyID         string
	R2SecretAccessKey     string
	R2Bucket              string
	PresignExpiresSeconds int

	BrevoAPIKey      string
	BrevoSenderEmail string
	BrevoSenderName  string

	AdultPlatformURL string
	KidsPlatformURL  string

	OpenRouterAPIKey        string
	OpenRouterModel         string
	OpenRouterFallbackModel string

	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string

	CronSecret string
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	// Read .env file if it exists, but don't fail if it doesn't
	_ = viper.ReadInConfig()

	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENV", "production")
	viper.SetDefault("ACCESS_TOKEN_TTL", "15m")
	viper.SetDefault("REFRESH_TOKEN_TTL", "168h")
	viper.SetDefault("CHILD_TOKEN_TTL", "4h")
	viper.SetDefault("OPENROUTER_MODEL", "google/gemma-2-9b-it")
	viper.SetDefault("OPENROUTER_FALLBACK_MODEL", "mistralai/mistral-7b-instruct")
	viper.SetDefault("PRESIGN_EXPIRES_SECONDS", 600)
	viper.SetDefault("AUTO_MIGRATE", true)
	viper.SetDefault("ADULT_PLATFORM_URL", "https://app.rawdah.app")
	viper.SetDefault("KIDS_PLATFORM_URL", "https://kids.rawdah.app")

	accessTTL, err := time.ParseDuration(viper.GetString("ACCESS_TOKEN_TTL"))
	if err != nil {
		return nil, err
	}
	refreshTTL, err := time.ParseDuration(viper.GetString("REFRESH_TOKEN_TTL"))
	if err != nil {
		return nil, err
	}
	childTTL, err := time.ParseDuration(viper.GetString("CHILD_TOKEN_TTL"))
	if err != nil {
		return nil, err
	}

	originsStr := viper.GetString("ALLOWED_ORIGINS")
	var origins []string
	if originsStr != "" {
		for _, o := range strings.Split(originsStr, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				origins = append(origins, o)
			}
		}
	}

	cfg := &Config{
		Port:           viper.GetString("PORT"),
		Env:            strings.ToLower(viper.GetString("ENV")),
		AllowedOrigins: origins,

		DatabaseURL: viper.GetString("DATABASE_URL"),
		AutoMigrate: viper.GetBool("AUTO_MIGRATE"),

		JWTAccessSecret:  viper.GetString("JWT_ACCESS_SECRET"),
		JWTRefreshSecret: viper.GetString("JWT_REFRESH_SECRET"),
		AccessTokenTTL:   accessTTL,
		RefreshTokenTTL:  refreshTTL,
		ChildTokenTTL:    childTTL,

		R2AccountID:           viper.GetString("R2_ACCOUNT_ID"),
		R2AccessKeyID:         viper.GetString("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey:     viper.GetString("R2_SECRET_ACCESS_KEY"),
		R2Bucket:              viper.GetString("R2_BUCKET"),
		PresignExpiresSeconds: viper.GetInt("PRESIGN_EXPIRES_SECONDS"),

		BrevoAPIKey:      viper.GetString("BREVO_API_KEY"),
		BrevoSenderEmail: viper.GetString("BREVO_SENDER_EMAIL"),
		BrevoSenderName:  viper.GetString("BREVO_SENDER_NAME"),
		AdultPlatformURL: viper.GetString("ADULT_PLATFORM_URL"),
		KidsPlatformURL:  viper.GetString("KIDS_PLATFORM_URL"),

		OpenRouterAPIKey:        viper.GetString("OPENROUTER_API_KEY"),
		OpenRouterModel:         viper.GetString("OPENROUTER_MODEL"),
		OpenRouterFallbackModel: viper.GetString("OPENROUTER_FALLBACK_MODEL"),

		VAPIDPublicKey:  viper.GetString("VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey: viper.GetString("VAPID_PRIVATE_KEY"),
		VAPIDSubject:    viper.GetString("VAPID_SUBJECT"),

		CronSecret: viper.GetString("CRON_SECRET"),
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validate(cfg *Config) error {
	var errs []string

	if cfg.Port == "" {
		errs = append(errs, "PORT is required")
	}
	if cfg.DatabaseURL == "" {
		errs = append(errs, "DATABASE_URL is required")
	}
	if len(cfg.JWTAccessSecret) < 64 {
		errs = append(errs, "JWT_ACCESS_SECRET must be at least 64 characters")
	}
	if len(cfg.JWTRefreshSecret) < 64 {
		errs = append(errs, "JWT_REFRESH_SECRET must be at least 64 characters")
	}
	if cfg.AccessTokenTTL <= 0 {
		errs = append(errs, "ACCESS_TOKEN_TTL must be > 0")
	}
	if cfg.RefreshTokenTTL <= 0 {
		errs = append(errs, "REFRESH_TOKEN_TTL must be > 0")
	}
	if cfg.ChildTokenTTL <= 0 {
		errs = append(errs, "CHILD_TOKEN_TTL must be > 0")
	}
	if cfg.Env == "production" && len(cfg.AllowedOrigins) == 0 {
		errs = append(errs, "ALLOWED_ORIGINS is required in production")
	}
	if cfg.PresignExpiresSeconds <= 0 {
		errs = append(errs, "PRESIGN_EXPIRES_SECONDS must be > 0")
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed: %s", strings.Join(errs, "; "))
	}

	return nil
}
