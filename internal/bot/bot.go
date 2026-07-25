package bot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"bookerbot-tgbot/internal/apiclient"
	"bookerbot-tgbot/internal/config"
	"bookerbot-tgbot/internal/models"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/telebot.v3"
)

// Bot wraps the telebot bot instance and its dependencies
type Bot struct {
	bot    *telebot.Bot
	cfg    *config.Config
	api    *apiclient.APIClient
}

// NewBot creates and initializes a new Telegram bot
func NewBot(cfg *config.Config, apiClient *apiclient.APIClient) (*Bot, error) {
	// Configure logger
	level, err := zerolog.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	if !cfg.Logging.JSON {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}

	pref := telebot.Settings{
		Token:  cfg.Telegram.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
		OnError: func(err error, c telebot.Context) {
			log.Error().Err(err).Str("chat_id", strconv.FormatInt(c.Sender().ID, 10)).Msg("bot error")
		},
	}

	tb, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	bot := &Bot{
		bot: tb,
		cfg: cfg,
		api: apiClient,
	}

	// Register middleware
	tb.Use(bot.recoveryMiddleware())

	// Register commands
	tb.Handle("/start", bot.handleStart)
	tb.Handle("/book", bot.handleBook)
	tb.Handle("/mybookings", bot.handleMyBookings)
	tb.Handle("/cancel", bot.handleCancel)
	tb.Handle("/help", bot.handleHelp)

	// Register callback handlers
	tb.Handle(&btnConfirmBooking, bot.handleConfirmBooking)
	tb.Handle(&btnCancelFlow, bot.handleCancelFlow)
	tb.Handle(&btnBack, bot.handleBack)
	tb.Handle(telebot.OnCallback, bot.handleCallback)

	// Register Telegram bot commands
	cmds := []telebot.Command{
		{Text: "start", Description: "Start the bot and select a merchant"},
		{Text: "book", Description: "Book an appointment"},
		{Text: "mybookings", Description: "View your bookings"},
		{Text: "cancel", Description: "Cancel a booking"},
		{Text: "help", Description: "Show available commands"},
	}
	if err := tb.SetCommands(cmds); err != nil {
		log.Warn().Err(err).Msg("failed to set bot commands")
	}

	return bot, nil
}

// Start starts the bot (polling mode by default)
func (b *Bot) Start() {
	log.Info().Msg("starting telegram bot in polling mode")
	b.bot.Start()
}

// Run starts the bot and waits for shutdown signal
func (b *Bot) Run() error {
	go b.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down bot...")
	b.bot.Stop()
	log.Info().Msg("bot stopped gracefully")
	return nil
}

// recoveryMiddleware catches panics in handlers
func (b *Bot) recoveryMiddleware() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			defer func() {
				if r := recover(); r != nil {
					log.Error().
						Interface("panic", r).
						Str("chat_id", strconv.FormatInt(c.Sender().ID, 10)).
						Msg("panic recovered in handler")
					_ = c.Send("An unexpected error occurred. Please try again.")
				}
			}()
			return next(c)
		}
	}
}

// --- Button definitions ---

var btnConfirmBooking = telebot.Btn{Unique: "confirm_booking", Text: "✅ Confirm Booking"}
var btnCancelFlow = telebot.Btn{Unique: "cancel_flow", Text: "❌ Cancel"}
var btnBack = telebot.Btn{Unique: "go_back", Text: "⬅ Back"}

// Callback prefixes for routing
const (
	cbSelectMerchant  = "m|"
	cbSelectService   = "s|"
	cbSelectDate      = "d|"
	cbSelectSlot      = "sl|"
	cbSelectStaff     = "st|"
	cbConfirmBook     = "book|"
	cbCancelBooking   = "cancelbk|"
	cbMyBookingsPage  = "mb|"
)

// --- Command Handlers ---

func (b *Bot) handleStart(c telebot.Context) error {
	chatID := strconv.FormatInt(c.Sender().ID, 10)
	username := c.Sender().Username
	log.Info().Str("chat_id", chatID).Str("username", username).Str("command", "/start").Msg("command received")

	return sendUserMessage(c, "handleStart", func() (string, *telebot.ReplyMarkup, error) {
		resp, err := b.api.GetMerchants(context.Background())
		if err != nil {
			return "", nil, err
		}

		if len(resp.Merchants) == 0 {
			return "No merchants are currently available. Please check back later.", nil, nil
		}

		markup := &telebot.ReplyMarkup{}
		var rows []telebot.Row
		for _, m := range resp.Merchants {
			text := m.Name
			if m.WelcomeMessage != "" {
				text = fmt.Sprintf("%s — %s", m.Name, m.WelcomeMessage)
			}
			btn := markup.Data(text, cbSelectMerchant, m.ID)
			rows = append(rows, markup.Row(btn))
		}
		markup.Inline(rows...)

		return "Welcome! Please select a merchant:", markup, nil
	})
}

func (b *Bot) handleHelp(c telebot.Context) error {
	chatID := strconv.FormatInt(c.Sender().ID, 10)
	log.Info().Str("chat_id", chatID).Str("command", "/help").Msg("command received")

	msg := `*Available Commands:*
/start — Start the bot and select a merchant
/book — Book an appointment
/mybookings — View your bookings
/cancel — Cancel an existing booking
/help — Show this help message

*How to book:*
1. Use /book to start
2. Select a merchant → service → date → time slot → staff member
3. Confirm your booking

All bookings are managed through our scheduling system.`
	return c.Send(msg, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

func (b *Bot) handleBook(c telebot.Context) error {
	chatID := strconv.FormatInt(c.Sender().ID, 10)
	log.Info().Str("chat_id", chatID).Str("command", "/book").Msg("command received")

	return sendUserMessage(c, "handleBook", func() (string, *telebot.ReplyMarkup, error) {
		resp, err := b.api.GetMerchants(context.Background())
		if err != nil {
			return "", nil, err
		}

		if len(resp.Merchants) == 0 {
			return "No merchants are currently available. Please check back later.", nil, nil
		}

		markup := &telebot.ReplyMarkup{}
		var rows []telebot.Row
		for _, m := range resp.Merchants {
			btn := markup.Data(m.Name, cbSelectMerchant, m.ID)
			rows = append(rows, markup.Row(btn))
		}
		markup.Inline(rows...)

		return "Select a merchant:", markup, nil
	})
}

func (b *Bot) handleMyBookings(c telebot.Context) error {
	chatID := strconv.FormatInt(c.Sender().ID, 10)
	log.Info().Str("chat_id", chatID).Str("command", "/mybookings").Msg("command received")

	return sendUserMessage(c, "handleMyBookings", func() (string, *telebot.ReplyMarkup, error) {
		resp, err := b.api.GetBookings(context.Background(), chatID)
		if err != nil {
			return "", nil, err
		}

		if len(resp.Bookings) == 0 {
			return "You have no bookings yet. Use /book to schedule one.", nil, nil
		}

		msg := "*Your Bookings:*\n\n"
		for _, bk := range resp.Bookings {
			emoji := "📅"
			switch bk.Status {
			case "cancelled":
				emoji = "❌"
			case "pending":
				emoji = "⏳"
			}
			msg += fmt.Sprintf("%s *%s*\n", emoji, bk.ServiceID)
			msg += fmt.Sprintf("  Date: %s\n", bk.StartTime[:10])
			msg += fmt.Sprintf("  Time: %s\n", bk.StartTime[11:16])
			msg += fmt.Sprintf("  Status: %s\n", bk.Status)
			msg += fmt.Sprintf("  ID: `%s`\n\n", bk.ID)
		}

		return msg, nil, nil
	}, telebot.ModeMarkdown)
}

func (b *Bot) handleCancel(c telebot.Context) error {
	chatID := strconv.FormatInt(c.Sender().ID, 10)
	log.Info().Str("chat_id", chatID).Str("command", "/cancel").Msg("command received")

	return sendUserMessage(c, "handleCancel", func() (string, *telebot.ReplyMarkup, error) {
		resp, err := b.api.GetBookings(context.Background(), chatID)
		if err != nil {
			return "", nil, err
		}

		// Filter only active bookings (not cancelled)
		var active []models.Booking
		for _, bk := range resp.Bookings {
			if bk.Status != "cancelled" {
				active = append(active, bk)
			}
		}

		if len(active) == 0 {
			return "You have no active bookings to cancel.", nil, nil
		}

		markup := &telebot.ReplyMarkup{}
		var rows []telebot.Row
		for _, bk := range active {
			text := fmt.Sprintf("%s at %s", bk.StartTime[:10], bk.StartTime[11:16])
			btn := markup.Data(text, cbCancelBooking, bk.ID)
			rows = append(rows, markup.Row(btn))
		}
		markup.Inline(rows...)

		return "Select a booking to cancel:", markup, nil
	}, telebot.ModeMarkdown)
}

// --- Callback Handlers ---

func (b *Bot) handleCallback(c telebot.Context) error {
	data := c.Callback().Data
	switch {
	case strings.HasPrefix(data, cbSelectMerchant):
		return b.onMerchantSelected(c, data[len(cbSelectMerchant):])
	case strings.HasPrefix(data, cbSelectService):
		return b.onServiceSelected(c, data[len(cbSelectService):])
	case strings.HasPrefix(data, cbSelectDate):
		return b.onDateSelected(c, data[len(cbSelectDate):])
	case strings.HasPrefix(data, cbSelectSlot):
		return b.onSlotSelected(c, data[len(cbSelectSlot):])
	case strings.HasPrefix(data, cbSelectStaff):
		return b.onStaffSelected(c, data[len(cbSelectStaff):])
	case strings.HasPrefix(data, cbConfirmBook):
		return b.onBookingConfirmed(c, data[len(cbConfirmBook):])
	case strings.HasPrefix(data, cbCancelBooking):
		return b.onCancelBookingSelected(c, data[len(cbCancelBooking):])
	default:
			log.Warn().Str("chat_id", strconv.FormatInt(c.Sender().ID, 10)).Str("data", data).Msg("unknown callback")
		_ = c.Respond(&telebot.CallbackResponse{Text: "Unknown action"})
	}
	return nil
}

func (b *Bot) handleConfirmBooking(c telebot.Context) error {
	// This is handled by the callback router with prefix "book|"
	_ = c.Respond(&telebot.CallbackResponse{Text: "Processing..."})
	return nil
}

func (b *Bot) handleCancelFlow(c telebot.Context) error {
	_ = c.Respond(&telebot.CallbackResponse{Text: "Cancelled"})
	_ = c.Delete()
	return c.Send("Booking cancelled. Use /book to start a new booking or /help to see all commands.")
}

func (b *Bot) handleBack(c telebot.Context) error {
	_ = c.Respond(&telebot.CallbackResponse{Text: "Going back..."})
	_ = c.Delete()
	return c.Send("Let's start over. Use /book to begin.")
}

func (b *Bot) onMerchantSelected(c telebot.Context, merchantID string) error {
	_ = c.Respond(&telebot.CallbackResponse{Text: "Loading services..."})

	return sendUserMessage(c, "onMerchantSelected", func() (string, *telebot.ReplyMarkup, error) {
		resp, err := b.api.GetServices(context.Background(), merchantID)
		if err != nil {
			return "", nil, err
		}

		if len(resp.Services) == 0 {
			return "This merchant has no services available yet. Please try another merchant.", nil, nil
		}

		markup := &telebot.ReplyMarkup{}
		var rows []telebot.Row
		for _, svc := range resp.Services {
			text := svc.Name
			if svc.Price > 0 {
				text = fmt.Sprintf("%s (%d min · $%.0f)", svc.Name, svc.DurationMinutes, svc.Price)
			} else {
				text = fmt.Sprintf("%s (%d min)", svc.Name, svc.DurationMinutes)
			}
			// Encode: merchantID + "|" + serviceID
			payload := fmt.Sprintf("%s|%s", merchantID, svc.ID)
			btn := markup.Data(text, cbSelectService, payload)
			rows = append(rows, markup.Row(btn))
		}
		rows = append(rows, markup.Row(btnCancelFlow))
		markup.Inline(rows...)

		return "Select a service:", markup, nil
	})
}

func (b *Bot) onServiceSelected(c telebot.Context, payload string) error {
	_ = c.Respond(&telebot.CallbackResponse{Text: "Loading dates..."})

	parts := strings.SplitN(payload, "|", 2)
	if len(parts) != 2 {
		return c.Send("Invalid selection. Please try again.")
	}
	merchantID := parts[0]
	serviceID := parts[1]

	return sendUserMessage(c, "onServiceSelected", func() (string, *telebot.ReplyMarkup, error) {
		dates, err := b.api.GetAvailableDates(context.Background(), serviceID)
		if err != nil {
			return "", nil, err
		}

		if len(dates) == 0 {
			return "No available dates found for this service. Please try another service or check back later.", nil, nil
		}

		markup := &telebot.ReplyMarkup{}
		var rows []telebot.Row
		for _, d := range dates {
			label := fmt.Sprintf("%s (%d slots)", d.Date, d.SlotsAvailable)
			payload := fmt.Sprintf("%s|%s|%s", merchantID, serviceID, d.Date)
			btn := markup.Data(label, cbSelectDate, payload)
			rows = append(rows, markup.Row(btn))
		}
		rows = append(rows, markup.Row(btnCancelFlow))
		markup.Inline(rows...)

		return "Select a date:", markup, nil
	})
}

func (b *Bot) onDateSelected(c telebot.Context, payload string) error {
	_ = c.Respond(&telebot.CallbackResponse{Text: "Loading time slots..."})

	parts := strings.SplitN(payload, "|", 3)
	if len(parts) != 3 {
		return c.Send("Invalid selection. Please try again.")
	}
	merchantID := parts[0]
	serviceID := parts[1]
	date := parts[2]

	return sendUserMessage(c, "onDateSelected", func() (string, *telebot.ReplyMarkup, error) {
		slots, err := b.api.GetAvailableSlots(context.Background(), serviceID, date, "")
		if err != nil {
			return "", nil, err
		}

		if len(slots) == 0 {
			return "No available time slots on this date. Please select a different date.", nil, nil
		}

		markup := &telebot.ReplyMarkup{}
		var rows []telebot.Row
		for _, slot := range slots {
			// Extract time portion for cleaner display
			timeLabel := slot.StartTime[11:16]
			payload := fmt.Sprintf("%s|%s|%s|%s", merchantID, serviceID, date, slot.StartTime)
			btn := markup.Data(timeLabel, cbSelectSlot, payload)
			rows = append(rows, markup.Row(btn))
		}
		rows = append(rows, markup.Row(btnCancelFlow))
		markup.Inline(rows...)

		return fmt.Sprintf("Available slots for %s:", date), markup, nil
	})
}

func (b *Bot) onSlotSelected(c telebot.Context, payload string) error {
	_ = c.Respond(&telebot.CallbackResponse{Text: "Loading staff..."})

	parts := strings.SplitN(payload, "|", 4)
	if len(parts) != 4 {
		return c.Send("Invalid selection. Please try again.")
	}
	merchantID := parts[0]
	serviceID := parts[1]
	date := parts[2]
	startTime := parts[3]

	return sendUserMessage(c, "onSlotSelected", func() (string, *telebot.ReplyMarkup, error) {
		resp, err := b.api.GetStaff(context.Background(), merchantID)
		if err != nil {
			return "", nil, err
		}

		if len(resp.Staff) == 0 {
			return "No staff members are available. Please try again later.", nil, nil
		}

		markup := &telebot.ReplyMarkup{}
		var rows []telebot.Row
		for _, st := range resp.Staff {
			payload := fmt.Sprintf("%s|%s|%s|%s|%s", merchantID, serviceID, date, startTime, st.ID)
			btn := markup.Data(st.Name, cbSelectStaff, payload)
			rows = append(rows, markup.Row(btn))
		}
		rows = append(rows, markup.Row(btnCancelFlow))
		markup.Inline(rows...)

		return fmt.Sprintf("Slot: %s at %s\n\nSelect a staff member:", date, startTime[11:16]), markup, nil
	})
}

func (b *Bot) onStaffSelected(c telebot.Context, payload string) error {
	_ = c.Respond(&telebot.CallbackResponse{Text: "Preparing confirmation..."})

	parts := strings.SplitN(payload, "|", 5)
	if len(parts) != 5 {
		return c.Send("Invalid selection. Please try again.")
	}
	_ = parts[0] // merchantID
	serviceID := parts[1]
	date := parts[2]
	startTime := parts[3]
	staffID := parts[4]

	chatID := strconv.FormatInt(c.Sender().ID, 10)
	confirmPayload := fmt.Sprintf("%s|%s|%s|%s", serviceID, staffID, startTime, chatID)

	markup := &telebot.ReplyMarkup{}
	markup.Inline(
		markup.Row(telebot.Btn{Text: btnConfirmBooking.Text, Unique: btnConfirmBooking.Unique, Data: cbConfirmBook + confirmPayload}),
		markup.Row(btnCancelFlow),
	)

	msg := fmt.Sprintf("*Booking Summary:*\n\n📅 Date: %s\n🕐 Time: %s\n\nConfirm your booking?", date, startTime[11:16])
	return c.Send(msg, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown, ReplyMarkup: markup})
}

func (b *Bot) onBookingConfirmed(c telebot.Context, payload string) error {
	_ = c.Respond(&telebot.CallbackResponse{Text: "Creating booking..."})

	parts := strings.SplitN(payload, "|", 4)
	if len(parts) != 4 {
		return c.Send("Invalid booking data. Please try again.")
	}
	serviceID := parts[0]
	staffID := parts[1]
	startTime := parts[2]
	chatID := parts[3]

	req := &models.BookingRequest{
		UserID:    chatID,
		ServiceID: serviceID,
		StaffID:   staffID,
		StartTime: startTime,
	}

	booking, err := b.api.CreateBooking(context.Background(), req)
	if err != nil {
		var apiErr *apiclient.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 409 {
			_ = c.Delete()
			return c.Send("Sorry, this slot is no longer available. Please choose another time. Use /book to start over.")
		}
		log.Error().Err(err).Str("chat_id", chatID).Msg("booking creation failed")
		return c.Send("Something went wrong while creating your booking. Please try again.")
	}

	_ = c.Delete()

	msg := fmt.Sprintf("*Booking Confirmed!* ✅\n\n📅 Date: %s\n🕐 Time: %s\n🔢 Booking ID: `%s`\n\nUse /mybookings to view your bookings.",
		startTime[:10], startTime[11:16], booking.ID)
	return c.Send(msg, &telebot.SendOptions{ParseMode: telebot.ModeMarkdown})
}

func (b *Bot) onCancelBookingSelected(c telebot.Context, bookingID string) error {
	_ = c.Respond(&telebot.CallbackResponse{Text: "Cancelling booking..."})

	err := b.api.CancelBooking(context.Background(), bookingID)
	if err != nil {
		var apiErr *apiclient.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			return c.Send("This booking was not found. It may have already been cancelled.")
		}
		log.Error().Err(err).Str("booking_id", bookingID).Msg("cancel booking failed")
		return c.Send("Failed to cancel booking. Please try again.")
	}

	_ = c.Delete()
	return c.Send("✅ Booking cancelled successfully.")
}

// --- Helpers ---

func sendUserMessage(c telebot.Context, handler string, fn func() (string, *telebot.ReplyMarkup, error), opts ...interface{}) error {
	msg, markup, err := fn()
	if err != nil {
		log.Error().Err(err).Str("handler", handler).Str("chat_id", strconv.FormatInt(c.Sender().ID, 10)).Msg("handler error")
		return c.Send("I'm having trouble reaching the booking service. Please try again in a moment.")
	}

	sendOpts := &telebot.SendOptions{}
	if markup != nil {
		sendOpts.ReplyMarkup = markup
	}
	for _, opt := range opts {
		switch o := opt.(type) {
		case string:
			sendOpts.ParseMode = o
		}
	}

	return c.Send(msg, sendOpts)
}
