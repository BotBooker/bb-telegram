## Why

The `tg-bot/` component of the `newbookerbot` project is currently an empty shell. The backend API (`api/`) is fully implemented with all CRUD operations for bookings, services, staff, and merchants, plus Redis-backed availability checks — but there is no client-facing interface for end users to interact with. A Telegram bot is the first messenger interface needed to make the booking system usable by real customers, enabling them to browse services, check availability, and book appointments directly from Telegram.

## What Changes

- Create a new Go module in `tg-bot/` implementing a Telegram bot using the `telebot` library
- Bot acts as an HTTP API client to the existing `api/` service endpoints defined in `api/spec/openapi.yaml`
- Implement conversational flows: select merchant → browse services → check dates → pick time slots → confirm booking
- Implement command-based interactions: `/start`, `/book`, `/mybookings`, `/cancel`, `/help`
- Include `X-API-Key` authentication header on all API requests
- Follow the same patterns established in `api/`: `go-chi` for internal routing (webhook), `zerolog` for logging, `testify` for table-driven tests, `konf` for configuration
- Use Telegram webhooks for production; fall back to long polling for local development

## Capabilities

### New Capabilities
- `telegram-bot-core`: Bot initialization, configuration, webhook/long-polling setup, Telegram API integration, and graceful shutdown
- `telegram-conversation-flow`: Multi-step conversational booking UI — merchant selection, service browsing, date/slot picking, booking confirmation, and cancellation
- `telegram-api-client`: HTTP client wrapper around the `api/` REST service with API key auth, error handling, retries, and typed response parsing

### Modified Capabilities
<!-- None — no existing specs or code in tg-bot/ to modify -->

## Impact

- `tg-bot/`: Entirely new Go module with its own `go.mod`, bringing in `gopkg.in/telebot.v3` and reusing project-wide conventions (`zerolog`, `konf`, `testify`)
- `api/`: No changes required. The bot consumes the existing OpenAPI endpoints — `GET /catalog/services`, `GET /catalog/staff`, `GET /availability/dates`, `GET /availability/slots`, `POST /bookings`, `GET /bookings`, `GET /bookings/{id}`, `PUT /bookings/{id}/cancel`
- No database changes, no new migrations
- No changes to existing OpenAPI specs
