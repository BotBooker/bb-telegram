# Implementation Plan: Telegram Bot

## Overview
Build the Telegram bot component for `newbookerbot` as a standalone Go module that consumes the existing `api/` REST service. The bot provides conversational booking flows (browse → select → book) and command-based interactions.

## Prerequisites
- `api/` service must be running and accessible
- Valid Telegram bot token (from @BotFather)
- Go 1.26+ installed

## Architecture

```
tg-bot/
├── cmd/bookerbot/main.go     # Entry point
├── internal/
│   ├── config/config.go      # Configuration (konf)
│   ├── bot/bot.go            # Telegram bot setup + handlers
│   ├── apiclient/client.go   # HTTP client for api/ service
│   └── models/models.go      # DTOs matching API responses
├── config/local/config.yaml  # Local dev config
├── go.mod
└── Makefile
```

## Step-by-Step Implementation

### Phase 1: Scaffolding
1. Initialize Go module
2. Create directory structure
3. Add dependencies (`telebot.v3`, `konf`, `zerolog`, `testify`)
4. Create config file and `.env.example`

### Phase 2: API Client
1. Implement `APIClient` struct with typed request/response DTOs
2. Add `X-API-Key` header injection
3. Implement all 8 API endpoint methods
4. Add retry logic for GET requests
5. Add timeout and context propagation
6. Write `httptest`-based unit tests

### Phase 3: Bot Core
1. Initialize `telebot.Bot` with config
2. Implement webhook/long-polling toggle
3. Add middleware (logging, recovery, rate limiting)
4. Register bot commands with Telegram
5. Add graceful shutdown

### Phase 4: Command Handlers
1. `/start` — merchant selection flow
2. `/book` — full booking flow (merchant → service → date → slot → staff → confirm)
3. `/mybookings` — list user bookings
4. `/cancel` — cancel booking flow
5. `/help` — command listing

### Phase 5: Integration & Polish
1. Wire entry point (`main.go`)
2. Create `Makefile` and `README.md`
3. End-to-end testing

## Success Criteria
- Bot connects to Telegram and responds to all 5 commands
- Users can complete a full booking flow (select merchant → service → date → slot → staff → confirm → see booking)
- Users can view and cancel their bookings
- All API interactions include proper error handling with user-friendly messages
- Tests pass with `go test ./... -race`
- Code follows the same conventions established in `api/`

## Stop/Escalation Rules
- **Escalate** if the Telegram Bot API returns unexpected errors that can't be resolved
- **Stop** if any task requires changes to the `api/` codebase — the bot must work with the existing API as-is
- **Stop** if a dependency conflict makes `telebot.v3` incompatible with Go 1.26

## Resolved Questions and Assumptions
- **Library**: Using `gopkg.in/telebot.v3` (maintained successor to v2)
- **State**: Stateless — no server-side session store; Telegram inline keyboard callbacks carry entity IDs
- **User identification**: Telegram's `chat_id` is used as the `user_id` in API requests
- **Merchant selection**: On every `/book`, user selects a merchant from the API's merchant list
- **No admin operations**: Bot is end-user facing only; admin stays API-only
