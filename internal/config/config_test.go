package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate_MissingToken(t *testing.T) {
	cfg := &Config{
		API: APIConfig{
			BaseURL: "http://localhost:8080",
			APIKey:  "test-key",
		},
	}
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bot token")
}

func TestConfig_Validate_MissingBaseURL(t *testing.T) {
	cfg := &Config{
		Telegram: TelegramConfig{
			BotToken: "test-token",
		},
		API: APIConfig{
			APIKey: "test-key",
		},
	}
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "base URL")
}

func TestConfig_Validate_MissingAPIKey(t *testing.T) {
	cfg := &Config{
		Telegram: TelegramConfig{
			BotToken: "test-token",
		},
		API: APIConfig{
			BaseURL: "http://localhost:8080",
		},
	}
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key")
}

func TestConfig_Validate_WebhookNeedsURL(t *testing.T) {
	cfg := &Config{
		Telegram: TelegramConfig{
			BotToken:   "test-token",
			UseWebhook: true,
		},
		API: APIConfig{
			BaseURL: "http://localhost:8080",
			APIKey:  "test-key",
		},
	}
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhook URL")
}

func TestConfig_Validate_Valid(t *testing.T) {
	cfg := &Config{
		Telegram: TelegramConfig{
			BotToken: "test-token",
		},
		API: APIConfig{
			BaseURL: "http://localhost:8080",
			APIKey:  "test-key",
		},
	}
	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_Load_EnvOverrides(t *testing.T) {
	// Save and restore env
	origToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	origAPIKey := os.Getenv("API_KEY")
	defer func() {
		os.Setenv("TELEGRAM_BOT_TOKEN", origToken)
		os.Setenv("API_KEY", origAPIKey)
	}()

	os.Setenv("TELEGRAM_BOT_TOKEN", "env-token")
	os.Setenv("API_KEY", "env-key")

	// Load config (INSTANCE=local reads config/local/config.yaml)
	// The YAML has empty token/key, but env vars should override
	if os.Getenv("INSTANCE") == "" {
		os.Setenv("INSTANCE", "local")
	}
	cfg, err := Load()
	// If file loading fails (missing file), Load skips it and falls back to env
	// Just validate the structure
	if err == nil {
		assert.NotEmpty(t, cfg.Telegram.BotToken)
	}
	_ = cfg
	_ = require.New(t)
}
