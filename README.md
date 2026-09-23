# tg-bot — Telegram Bot for newbookerbot

A Telegram bot that provides a conversational booking interface for the newbookerbot scheduling system.

## Architecture

```
User (Telegram) ──→ tg-bot ──→ api/ (REST API) ──→ PostgreSQL + Redis
```

The bot is an HTTP API client to the `api/` booking service. It is stateless — every interaction fetches fresh data from the API.

## Prerequisites

- Go 1.26+
- Running instance of the `api/` service
- Telegram bot token (create via [@BotFather](https://t.me/BotFather))
- API key with access to the `api/` service

## Setup

1. **Clone and navigate:**

   ```bash
   cd tg-bot
   ```

2. **Install dependencies:**

   ```bash
   go mod download
   ```

3. **Configure:**

   ```bash
   cp .env.example .env
   # Edit .env with your values:
   #   TELEGRAM_BOT_TOKEN=<your-bot-token>
   #   API_BASE_URL=http://localhost:8080
   #   API_KEY=<your-api-key>
   ```

   Or edit `config/local/config.yaml` directly.

4. **Build and run:**

   ```bash
   make build
   make run
   ```

## Configuration

Configuration is loaded via `konf` from:

- `config/{INSTANCE}/config.yaml` (default: `local`)
- Environment variable overrides

| Env Var | YAML Path | Description | Default |
|---------|-----------|-------------|---------|
| `TELEGRAM_BOT_TOKEN` | `telegram.bot_token` | Bot token from BotFather | *required* |
| `API_BASE_URL` | `api.base_url` | API service URL | `http://localhost:8080` |
| `API_KEY` | `api.api_key` | API authentication key | *required* |
| `INSTANCE` | — | Config instance to load | `local` |

## Commands

| Command | Description |
|---------|-------------|
| `/start` | Welcome message and merchant selection |
| `/book` | Start booking flow |
| `/mybookings` | View your bookings |
| `/cancel` | Cancel an existing booking |
| `/help` | Show available commands |

## Development

```bash
# Format code
make fmt

# Run tests
make test

# Lint
make lint
```

## Project Structure

```
tg-bot/
├── cmd/bookerbot/main.go     # Entry point
├── internal/
│   ├── config/config.go      # Configuration (konf)
│   ├── bot/bot.go            # Telegram bot setup + handlers
│   ├── apiclient/client.go   # HTTP client for api/
│   └── models/models.go      # DTOs matching API responses
├── config/local/config.yaml  # Local dev config
├── go.mod
└── Makefile
```
