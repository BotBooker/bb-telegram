## 1. Project Scaffolding

- [x] 1.1 Initialize Go module in `tg-bot/` with `go mod init bookerbot-tgbot`
- [x] 1.2 Create directory structure: `cmd/bookerbot/`, `internal/config/`, `internal/bot/`, `internal/apiclient/`, `internal/models/`, `config/local/`
- [x] 1.3 Create `tg-bot/config/local/config.yaml` with all required settings (telegram token placeholder, API base URL, API key, webhook config, server port, log level)
- [x] 1.4 Create `tg-bot/.env.example` and `tg-bot/.gitignore`
- [x] 1.5 Add `gopkg.in/telebot.v3` dependency

## 2. Configuration

- [x] 2.1 Implement `internal/config/config.go` using `konf` with the same patterns as `api/internal/config/config.go`
- [x] 2.2 Define `Config` struct with `TelegramConfig`, `APIConfig`, `ServerConfig`, `LoggingConfig`
- [x] 2.3 Support `INSTANCE` env var to select YAML config file (same as `api/`)
- [x] 2.4 Support environment variable overrides for all settings
- [x] 2.5 Validate required fields on load (bot token, API base URL, API key)

## 3. API Client

- [x] 3.1 Implement `internal/apiclient/client.go` with `APIClient` struct, constructor `NewAPIClient(baseURL, apiKey string)`
- [x] 3.2 Define typed request/response structs in `internal/models/` matching `api/spec/openapi.yaml` schemas (Service, Staff, Merchant, Booking, BookingRequest, ErrorResponse, date/slot responses)
- [x] 3.3 Implement `X-API-Key` header injection on all requests
- [x] 3.4 Implement `GetMerchants(ctx) ([]Merchant, error)` — calls `GET /api/v1/admin/merchants/list`
- [x] 3.5 Implement `GetServices(ctx, merchantID) ([]Service, error)` — calls `GET /api/v1/catalog/services`
- [x] 3.6 Implement `GetStaff(ctx, merchantID) ([]Staff, error)` — calls `GET /api/v1/catalog/staff`
- [x] 3.7 Implement `GetAvailableDates(ctx, serviceID) ([]string, error)` — calls `GET /api/v1/availability/dates`
- [x] 3.8 Implement `GetAvailableSlots(ctx, serviceID, date, staffID) ([]Slot, error)` — calls `GET /api/v1/availability/slots`
- [x] 3.9 Implement `CreateBooking(ctx, req BookingRequest) (*Booking, error)` — calls `POST /api/v1/bookings`
- [x] 3.10 Implement `GetBookings(ctx, userID, status) ([]Booking, error)` — calls `GET /api/v1/bookings`
- [x] 3.11 Implement `CancelBooking(ctx, bookingID) error` — calls `PUT /api/v1/bookings/{id}/cancel`
- [x] 3.12 Implement HTTP status code → typed error mapping (`APIError`, `NetworkError`)
- [x] 3.13 Implement retry with exponential backoff for idempotent GET requests (3 retries, 1s/2s/4s)
- [x] 3.14 Implement configurable request timeout (default 10s) with context propagation
- [x] 3.15 Write table-driven unit tests for APIClient methods using `httptest` server

## 4. Bot Core (`internal/bot/`)

- [x] 4.1 Implement `internal/bot/bot.go` — `NewBot(cfg, apiClient)` constructor, `telebot` initialization
- [x] 4.2 Implement webhook mode (register webhook path, set Telegram webhook URL) vs long-polling mode toggle
- [x] 4.3 Implement logging middleware — log every incoming update with `chat_id`, `username`, command
- [x] 4.4 Implement recovery middleware — catch panics, log stack trace, prevent bot crash
- [x] 4.5 Register Telegram bot commands: `/start`, `/book`, `/mybookings`, `/cancel`, `/help`
- [x] 4.6 Implement graceful shutdown (SIGINT/SIGTERM → stop polling/webhook, close connections)

## 5. Command Handlers

- [x] 5.1 Implement `/start` handler — fetch merchants from API, present as inline keyboard buttons with welcome messages
- [x] 5.2 Implement `/help` handler — display list of all available commands with descriptions
- [x] 5.3 Implement `/book` handler entry point — if no merchant selected, show merchant list; otherwise jump to services
- [x] 5.4 Implement merchant selection callback — store merchant ID, fetch and display services
- [x] 5.5 Implement service selection callback — store service ID, fetch and display available dates
- [x] 5.6 Implement date selection callback — store date, fetch and display available time slots
- [x] 5.7 Implement slot selection callback — store time slot, fetch staff list for merchant, display staff selection
- [x] 5.8 Implement staff selection callback — store staff ID, display booking summary and confirmation button
- [x] 5.9 Implement booking confirmation callback — call `CreateBooking`, display success/conflict/error message
- [x] 5.10 Implement cancel/back button at each conversation step
- [x] 5.11 Implement `/mybookings` handler — fetch bookings for user's `chat_id`, display formatted list or "no bookings" message
- [x] 5.12 Implement `/cancel` handler — fetch active bookings, display as inline buttons for selection, call `CancelBooking` on selection
- [x] 5.13 Ensure all error messages to users are friendly (never expose raw error details or stack traces)

## 6. Entry Point & Integration

- [x] 6.1 Implement `cmd/bookerbot/main.go` — load config, initialize API client, initialize bot, start server
- [x] 6.2 Wire up graceful shutdown: signal handling → bot stop → process exit
- [x] 6.3 Create `tg-bot/Makefile` with `build`, `run`, `test`, `mock`, `fmt`, `lint` targets (mirroring `api/Makefile`)
- [x] 6.4 Create `tg-bot/README.md` with setup instructions (bot token, config, running locally)

## 7. Testing

- [x] 7.1 Write table-driven tests for `internal/apiclient/` using `httptest.NewServer` to mock the API
- [x] 7.2 Write tests for `internal/bot/` handlers using `telebot`'s offline-testing patterns (or mock `telebot.Bot`)
- [x] 7.3 Write tests for `internal/config/` to verify config loading and validation
- [x] 7.4 Ensure all tests pass: `go test ./... -count=2 -race` with no race conditions
