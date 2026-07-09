package config

import (
	"fmt"
	"os"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"github.com/rs/zerolog/log"
)

// Config holds all configuration for the Telegram bot.
type Config struct {
	Server   ServerConfig   `koanf:"server"`
	Telegram TelegramConfig `koanf:"telegram"`
	API      APIConfig      `koanf:"api"`
	Logging  LoggingConfig  `koanf:"logging"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port         string        `koanf:"port"`
	ReadTimeout  time.Duration `koanf:"read_timeout"`
	WriteTimeout time.Duration `koanf:"write_timeout"`
	IdleTimeout  time.Duration `koanf:"idle_timeout"`
}

// TelegramConfig holds Telegram bot configuration.
type TelegramConfig struct {
	BotToken     string        `koanf:"bot_token"`
	UseWebhook   bool          `koanf:"use_webhook"`
	WebhookURL   string        `koanf:"webhook_url"`
	WebhookPath  string        `koanf:"webhook_path"`
	PollInterval time.Duration `koanf:"poll_interval"`
}

// APIConfig holds API client configuration.
type APIConfig struct {
	BaseURL string        `koanf:"base_url"`
	APIKey  string        `koanf:"api_key"`
	Timeout time.Duration `koanf:"timeout"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level string `koanf:"level"`
	JSON  bool   `koanf:"json"`
}

// Load loads configuration from YAML file and environment variables.
func Load() (*Config, error) {
	k := koanf.New(".")

	instance := os.Getenv("INSTANCE")
	if instance == "" {
		instance = "local"
	}

	path := "config/" + instance + "/config.yaml"
	log.Info().Msgf("using instance: %s", instance)
	log.Info().Msgf("loading configuration from: %s", path)

	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		log.Warn().Err(err).Msg("failed to load config file, using defaults")
	}

	// Load from environment variables
	if err := k.Load(env.Provider("", "_", nil), nil); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	var config Config
	if err := k.Unmarshal("", &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	config.applyDefaults()

	// Validate required fields
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

// applyDefaults sets sensible defaults for any unconfigured fields.
func (c *Config) applyDefaults() {
	if c.Server.Port == "" {
		c.Server.Port = "8081"
	}
	if c.Server.ReadTimeout <= 0 {
		c.Server.ReadTimeout = 30 * time.Second
	}
	if c.Server.WriteTimeout <= 0 {
		c.Server.WriteTimeout = 30 * time.Second
	}
	if c.Server.IdleTimeout <= 0 {
		c.Server.IdleTimeout = 60 * time.Second
	}
	if c.Telegram.WebhookPath == "" {
		c.Telegram.WebhookPath = "/webhook"
	}
	if c.Telegram.PollInterval <= 0 {
		c.Telegram.PollInterval = 1 * time.Second
	}
	if c.API.Timeout <= 0 {
		c.API.Timeout = 10 * time.Second
	}
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
}

// Validate checks that required configuration fields are present.
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
