## ADDED Requirements

### Requirement: Bot initialization and configuration
The system SHALL load configuration from YAML files and environment variables using `konf`, supporting `TELEGRAM_BOT_TOKEN`, API base URL, API key, webhook mode toggle, and server port. A missing bot token MUST cause the process to exit with a fatal error.

#### Scenario: Configuration loaded from local YAML file
- **WHEN** the bot starts with `INSTANCE=local`
- **THEN** configuration is loaded from `config/local/config.yaml` and overridden by matching environment variables

#### Scenario: Missing bot token causes fatal exit
- **WHEN** the bot starts without `TELEGRAM_BOT_TOKEN` set in config or env
- **THEN** the process logs a fatal error and exits with code 1

### Requirement: Telegram API integration
The system SHALL connect to the Telegram Bot API using `gopkg.in/telebot.v3`, supporting both webhook and long-polling modes selectable via configuration. In webhook mode, the bot SHALL register a handler at the configured webhook path and set the Telegram webhook URL on startup. In polling mode, the bot SHALL continuously poll for updates.

#### Scenario: Bot starts in webhook mode
- **WHEN** `telegram.use_webhook` is `true` and `telegram.webhook_url` is configured
- **THEN** the bot registers the webhook with Telegram and listens for incoming updates at the configured path

#### Scenario: Bot starts in long-polling mode
- **WHEN** `telegram.use_webhook` is `false`
- **THEN** the bot begins polling Telegram for updates every `telegram.poll_interval` seconds

#### Scenario: Bot responds to a ping
- **WHEN** the bot receives a valid Telegram update
- **THEN** the bot processes the update through registered handlers

### Requirement: Graceful shutdown
The system SHALL listen for SIGINT and SIGTERM signals and shut down gracefully: stop accepting new requests, finish in-flight updates, close the Telegram API connection, and exit cleanly.

#### Scenario: Bot shuts down on SIGTERM
- **WHEN** the bot process receives SIGTERM
- **THEN** the bot stops polling/unregisters webhook, closes connections, and exits with code 0 within 30 seconds

### Requirement: Structured logging
The system SHALL use `zerolog` for all logging, with configurable log level. Every log entry related to a user interaction SHALL include the `chat_id` and `username` context fields.

#### Scenario: User interaction logged with context
- **WHEN** a user sends a `/book` command
- **THEN** a log entry is emitted at INFO level containing the `chat_id`, `username`, and command name

### Requirement: Command registration
The system SHALL register the following bot commands with Telegram: `/start`, `/book`, `/mybookings`, `/cancel`, `/help`. Each command SHALL have a `telebot` handler registered.

#### Scenario: /start command triggers welcome handler
- **WHEN** a user sends `/start`
- **THEN** the welcome handler is invoked and the user sees a greeting with available commands

#### Scenario: /help command lists all commands
- **WHEN** a user sends `/help`
- **THEN** the bot replies with a message listing all available commands and their descriptions

### Requirement: Middleware pipeline
The system SHALL apply `telebot` middleware in this order: logging, error recovery, rate limiting. The recovery middleware SHALL catch panics in handlers and log them with a stack trace without crashing the bot.

#### Scenario: Panic in handler is recovered
- **WHEN** a handler panics during execution
- **THEN** the recovery middleware catches the panic, logs the error with stack trace, and the bot continues processing other updates
