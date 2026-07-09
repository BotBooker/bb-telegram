package bot

import (
	"testing"
)

// Integration tests for bot package require a real Telegram bot token.
// These tests validate the package compiles and basic structures exist.
// Actual bot behavior testing should be done against a test Telegram bot.

func TestConstants_Defined(t *testing.T) {
	// Verify callback prefix constants are non-empty
	if cbSelectMerchant == "" {
		t.Error("cbSelectMerchant should not be empty")
	}
	if cbSelectService == "" {
		t.Error("cbSelectService should not be empty")
	}
	if cbSelectDate == "" {
		t.Error("cbSelectDate should not be empty")
	}
	if cbSelectSlot == "" {
		t.Error("cbSelectSlot should not be empty")
	}
	if cbSelectStaff == "" {
		t.Error("cbSelectStaff should not be empty")
	}
	if cbConfirmBook == "" {
		t.Error("cbConfirmBook should not be empty")
	}
	if cbCancelBooking == "" {
		t.Error("cbCancelBooking should not be empty")
	}
}

func TestButtonDefinitions(t *testing.T) {
	if btnConfirmBooking.Text == "" {
		t.Error("btnConfirmBooking should have text")
	}
	if btnCancelFlow.Text == "" {
		t.Error("btnCancelFlow should have text")
	}
	if btnBack.Text == "" {
		t.Error("btnBack should have text")
	}
}
