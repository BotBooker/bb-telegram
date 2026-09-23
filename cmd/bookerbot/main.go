package main

import (
	"os"

	"bookerbot-tgbot/internal/apiclient"
	"bookerbot-tgbot/internal/bot"
	"bookerbot-tgbot/internal/config"

	"github.com/rs/zerolog/log"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	// Initialize API client
	apiClient := apiclient.NewAPIClient(cfg.API.BaseURL, cfg.API.APIKey, cfg.API.Timeout)

	// Initialize bot
	tgBot, err := bot.NewBot(cfg, apiClient)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create bot")
	}

	// Run bot (blocks until shutdown signal)
	if err := tgBot.Run(); err != nil {
		log.Error().Err(err).Msg("bot exited with error")
		os.Exit(1)
	}
}
