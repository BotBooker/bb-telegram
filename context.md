# Context: tg-bot (Telegram Bot for newbookerbot)

## Overview
The Telegram bot is the first client-facing interface for the `newbookerbot` booking system. It interacts with end users through Telegram, providing conversational booking flows powered by the backend API.

## Architecture
The bot is an **independent Go module** that acts as an **HTTP API client** to the `api/` service. It is stateless — all state is fetched live from the API on every interaction.

```
User (Telegram) ──→ tg-bot ──→ api/ (REST API) ──→ PostgreSQL + Redis
```

## Relevant Files
- `cmd/bookerbot/main.go`: Entry point (to be created)
- `internal/config/config.go`: Configuration loading via `konf` (to be created)
- `internal/bot/bot.go`: Telegram bot setup, middleware, handlers (to be created)
- `internal/apiclient/client.go`: HTTP client wrapping the `api/` service (to be created)
- `internal/models/`: Shared DTOs matching API response schemas (to be created)
- `config/local/config.yaml`: Local development config (to be created)
- `openspec/changes/implement-telegram-bot/`: OpenSpec change with full proposal, design, specs, and tasks

## Key Dependencies
- `gopkg.in/telebot.v3`: Telegram Bot API library with expressive handler API
- `github.com/nil-go/konf`: Configuration management (same as `api/`)
- `github.com/rs/zerolog`: Structured logging (same as `api/`)
- `github.com/stretchr/testify`: Testing framework (same as `api/`)

## Important Patterns
- **Stateless design**: No local state — every action fetches fresh data from the API
- **Conversation flow via callbacks**: Inline keyboard callback data encodes entity IDs (merchant, service, date, slot, staff); no server-side session store needed
- **Dependency injection**: Bot handlers receive `apiclient.APIClient` as a dependency, analogous to `api/` handlers receiving repositories
- **Webhook + long-polling dual mode**: Config flag toggles between webhook (production) and long polling (local dev)
- **Middleware pipeline**: Logging → Recovery → Rate Limiting (mirrors `api/` middleware pattern)
- **Table-driven tests**: Using `testify` and `httptest` for API client tests

## API Endpoints Consumed
| Endpoint | Usage |
|----------|-------|
| `GET /admin/merchants/list` | Merchant selection |
| `GET /catalog/services?merchant_id=` | Service browsing |
| `GET /catalog/staff?merchant_id=` | Staff selection |
| `GET /availability/dates?service_id=` | Date picking |
| `GET /availability/slots?service_id=&date=&staff_id=` | Time slot picking |
| `POST /bookings` | Creating bookings |
| `GET /bookings?user_id=&status=` | Listing user bookings |
| `PUT /bookings/{id}/cancel` | Cancelling bookings |

## Constraints
- Must authenticate with the API using `X-API-Key` header
- Must not modify any files in `api/`
- Must not modify the OpenAPI specs in `api/spec/`
- Must follow the same Go project conventions as `api/`
- Telegram bot token is required and configured via environment variable or config file

## Bot Commands
| Command | Description |
|---------|-------------|
| `/start` | Welcome + merchant selection |
| `/book` | Start booking flow |
| `/mybookings` | List user's bookings |
| `/cancel` | Cancel an existing booking |
| `/help` | Show available commands |
