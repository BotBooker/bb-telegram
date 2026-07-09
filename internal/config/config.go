package config

import (
	"fmt"
	"os"
	"time"

	"github.com/nil-go/konf"
	"github.com/nil-go/konf/provider/env"
	"github.com/nil-go/konf/provider/file"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `konf:"server"`
	Telegram TelegramConfig `konf:"telegram"`
	API      APIConfig      `konf:"api"`
	Logging  LoggingConfig  `konf:"logging"`
}

type ServerConfig struct {
	Port         string        `konf:"port,default=8081"`
	ReadTimeout  time.Duration `konf:"read_timeout,default=30s"`
	WriteTimeout time.Duration `konf:"write_timeout,default=30s"`
	IdleTimeout  time.Duration `konf:"idle_timeout,default=60s"`
}

type TelegramConfig struct {
	BotToken     string        `konf:"bot_token"`
	UseWebhook   bool          `konf:"use_webhook,default=false"`
	WebhookURL   string        `konf:"webhook_url"`
	WebhookPath  string        `konf:"webhook_path,default=/webhook"`
	PollInterval time.Duration `konf:"poll_interval,default=1s"`
}

type APIConfig struct {
	BaseURL string        `konf:"base_url"`
	APIKey  string        `konf:"api_key"`
	Timeout time.Duration `konf:"timeout,default=10s"`
}

type LoggingConfig struct {
	Level string `konf:"level,default=info"`
	JSON  bool   `konf:"json,default=false"`
}

func Load() (*Config, error) {
	cfg := konf.New()

	instance := os.Getenv("INSTANCE")
	if instance == "" {
		instance = "local"
	}

	path := "config/" + instance + "/config.yaml"
	log.Info().Msgf("using instance: %s", instance)
	log.Info().Msgf("loading configuration from: %s", path)

	if err := cfg.Load(file.New(path, file.WithUnmarshal(yaml.Unmarshal))); err != nil {
		log.Warn().Err(err).Msgf("failed to load config file, using defaults")
	}

	// Load from environment variables
	if err := cfg.Load(env.New(env.WithPrefix(""))); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	konf.SetDefault(cfg)

	var config Config
	if err := cfg.Unmarshal("", &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Validate required fields
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if c.Telegram.BotToken == "" {
		return fmt.Errorf("telegram bot token is required (set TELEGRAM_BOT_TOKEN env var or telegram.bot_token in config)")
	}
	if c.API.BaseURL == "" {
		return fmt.Errorf("API base URL is required (set API_BASE_URL env var or api.base_url in config)")
	}
	if c.API.APIKey == "" {
		return fmt.Errorf("API key is required (set API_KEY env var or api.api_key in config)")
	}
	if c.Telegram.UseWebhook && c.Telegram.WebhookURL == "" {
		return fmt.Errorf("webhook URL is required when use_webhook is true")
	}
	return nil
}
