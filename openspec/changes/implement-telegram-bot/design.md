## Context

The `tg-bot/` directory is currently empty. The backend API (`api/`) is fully functional with PostgreSQL for persistence, Redis for availability caching, and a Go `go-chi` stack. The bot needs to be an independent Go module that acts as a stateless HTTP client to the API, managing all user interaction through Telegram's Bot API.

Key constraints:
- The bot must authenticate with the API using `X-API-Key` header
- The OpenAPI specs in `api/spec/` are immutable — the bot adapts to the existing API
- Follow the same Go patterns as `api/`: `zerolog` for logging, `konf` for config, `testify` for table-driven tests
- Telegram library: `gopkg.in/telebot.v3` (the maintained successor to the original `telebot.v2`)
- Telegram integration via webhooks in production, long polling for local development

## Goals / Non-Goals

**Goals:**
- Full conversational booking flow in Telegram (browse → select → book)
- Command-based entry points: `/start`, `/book`, `/mybookings`, `/cancel`, `/help`
- Multi-merchant support: users select which merchant to book with
- Stateless architecture: every bot interaction reads fresh state from the API
- Graceful shutdown and webhook/long-polling toggle via config

**Non-Goals:**
- Admin operations via bot (admin management stays API-only)
- Payment integration
- Push notifications or reminders
- Multi-language support
- Inline keyboards for admin CRUD
- Persistent user sessions (rely on Telegram `chat_id` + API calls)
- Bot-side caching — always call the API for fresh data

## Decisions

### 1. Telegram library: `gopkg.in/telebot.v3`
- **Alternative**: `go-telegram-bot-api/telegram-bot-api/v5`. `telebot.v3` is more expressive, has cleaner middleware/chaining, and a more idiomatic Go API. It's the maintained evolution of the widely-used v2.
- **Rationale**: Better handler ergonomics, built-in state management for conversations, inline keyboard helpers. The research document confirmed both are viable; `telebot` is preferred for complex conversation flows.

### 2. Stateless design — no local state
- **Alternative**: Store conversation state in Redis or in-memory maps.
- **Rationale**: The API is the single source of truth. Storing state in the bot would create consistency issues. `telebot`'s `State` feature will be used only for ephemeral conversation position tracking (e.g., "waiting for date selection"), not for caching data.

### 3. Dedicated API client package
- **Alternative**: Call `http.Get` / `http.Post` directly in bot handlers.
- **Rationale**: A dedicated `apiclient` package with typed request/response structs, error handling, retries, and the `X-API-Key` header injected automatically keeps bot handlers clean and testable. This mirrors the established repository pattern from `api/`.

### 4. Configuration via `konf` (reusing `api/` patterns)
- **Alternative**: Viper or env-only.
- **Rationale**: Project consistency. `api/` already uses `konf` with YAML + env overrides. Same `config/{instance}/config.yaml` + environment variable pattern.

### 5. Project layout mirrors `api/`
```
tg-bot/
├── cmd/bookerbot/        # Entry point
├── internal/
│   ├── config/           # Configuration loading
│   ├── bot/              # Telegram bot setup, handlers, middleware
│   ├── apiclient/        # HTTP client for the api/ service
│   └── models/           # Shared DTOs matching API responses
├── config/local/         # Local config
└── go.mod                # Independent module
```
- **Rationale**: Developers familiar with `api/` will immediately understand `tg-bot/`.

### 6. Conversation flow: state machine via `telebot` endpoints + callback data
- When a user presses an inline button, the callback data encodes the action and current selection (e.g., `merchant:uuid`, `date:2026-07-10`, `service:uuid`).
- This avoids server-side state for selections.
- **Rationale**: Telegram inline keyboard callbacks can carry up to 64 bytes of payload — enough for entity IDs. No need for a state store.

### 7. Webhook + long-polling dual mode
- Config flag `telegram.use_webhook` determines mode.
- Webhook path: `POST /webhook/{bot_token}`.
- **Rationale**: Long polling for local dev (no public URL needed), webhooks for production (lower latency, less overhead).

## Risks / Trade-offs

- **API availability dependency**: If the API is down, the bot is non-functional. Mitigation: informative error messages to users ("Service temporarily unavailable, please try again later"), retries with exponential backoff in the API client.
- **Max Messenger unknown**: Nothing in this design is Max-specific, but the bot architecture (stateless HTTP client) should be reusable for `max-bot` later.
- **Callback data size limit**: Telegram limits callback data to 64 bytes. UUIDs are 36 chars — we stay within limits by using only essential identifiers. If we ever need more complex state, switch to a short-lived state key stored in the bot process memory or a dedicated state endpoint in the API.
- **Rate limiting**: Telegram enforces ~30 messages/second per bot. The bot will implement middleware for rate limiting, though for the initial scope this is unlikely to be hit.

## Open Questions

- What is the expected bot name/username in Telegram? (Configurable via `TELEGRAM_BOT_TOKEN`, set by the deployer)
- Should the bot show merchant welcome messages? → Yes, as part of the `/start` flow
- Is there a need for an admin-approval flow for bookings? → Not in initial scope; bookings are auto-confirmed
