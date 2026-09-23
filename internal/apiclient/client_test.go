package apiclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bookerbot-tgbot/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIClient_GetMerchants(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/admin/merchants/list", r.URL.Path)
		assert.Equal(t, "test-api-key", r.Header.Get("X-API-Key"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.MerchantListResponse{
			Total: 2,
			Merchants: []models.Merchant{
				{ID: "m1", Name: "Test Merchant 1", WelcomeMessage: "Welcome!"},
				{ID: "m2", Name: "Test Merchant 2"},
			},
		})
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-api-key", 10*time.Second)
	resp, err := client.GetMerchants(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Equal(t, "m1", resp.Merchants[0].ID)
	assert.Equal(t, "Welcome!", resp.Merchants[0].WelcomeMessage)
}

func TestAPIClient_GetServices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/catalog/services", r.URL.Path)
		assert.Equal(t, "m1", r.URL.Query().Get("merchant_id"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.ServiceListResponse{
			Total: 1,
			Services: []models.Service{
				{ID: "s1", Name: "Haircut", DurationMinutes: 30, Price: 25.0},
			},
		})
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-api-key", 10*time.Second)
	resp, err := client.GetServices(context.Background(), "m1")
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Equal(t, "Haircut", resp.Services[0].Name)
	assert.Equal(t, 25.0, resp.Services[0].Price)
}

func TestAPIClient_GetStaff(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/catalog/staff", r.URL.Path)
		assert.Equal(t, "m1", r.URL.Query().Get("merchant_id"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.StaffListResponse{
			Total: 1,
			Staff: []models.Staff{
				{ID: "st1", Name: "Alice", Services: []string{"s1"}},
			},
		})
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-api-key", 10*time.Second)
	resp, err := client.GetStaff(context.Background(), "m1")
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Equal(t, "Alice", resp.Staff[0].Name)
	assert.Equal(t, []string{"s1"}, resp.Staff[0].Services)
}

func TestAPIClient_GetAvailableDates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/availability/dates", r.URL.Path)
		assert.Equal(t, "s1", r.URL.Query().Get("service_id"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]models.AvailableDate{
			{Date: "2026-07-10", SlotsAvailable: 5},
			{Date: "2026-07-11", SlotsAvailable: 3},
		})
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-api-key", 10*time.Second)
	dates, err := client.GetAvailableDates(context.Background(), "s1")
	require.NoError(t, err)
	assert.Len(t, dates, 2)
	assert.Equal(t, "2026-07-10", dates[0].Date)
	assert.Equal(t, 5, dates[0].SlotsAvailable)
}

func TestAPIClient_GetAvailableSlots(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/availability/slots", r.URL.Path)
		assert.Equal(t, "s1", r.URL.Query().Get("service_id"))
		assert.Equal(t, "2026-07-10", r.URL.Query().Get("date"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]models.AvailableSlot{
			{StartTime: "2026-07-10T09:00:00Z", EndTime: "2026-07-10T09:30:00Z"},
			{StartTime: "2026-07-10T10:00:00Z", EndTime: "2026-07-10T10:30:00Z"},
		})
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-api-key", 10*time.Second)
	slots, err := client.GetAvailableSlots(context.Background(), "s1", "2026-07-10", "")
	require.NoError(t, err)
	assert.Len(t, slots, 2)
	assert.Equal(t, "2026-07-10T09:00:00Z", slots[0].StartTime)
}

func TestAPIClient_CreateBooking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/bookings", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(models.Booking{
			ID:        "bk1",
			UserID:    "user1",
			ServiceID: "s1",
			StaffID:   "st1",
			StartTime: "2026-07-10T09:00:00Z",
			Status:    "confirmed",
		})
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-api-key", 10*time.Second)
	req := &models.BookingRequest{
		UserID:    "user1",
		ServiceID: "s1",
		StaffID:   "st1",
		StartTime: "2026-07-10T09:00:00Z",
	}
	booking, err := client.CreateBooking(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "bk1", booking.ID)
	assert.Equal(t, "confirmed", booking.Status)
}

func TestAPIClient_GetBookings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/bookings", r.URL.Path)
		assert.Equal(t, "user1", r.URL.Query().Get("user_id"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.BookingListResponse{
			Total: 1,
			Bookings: []models.Booking{
				{ID: "bk1", UserID: "user1", Status: "confirmed"},
			},
		})
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-api-key", 10*time.Second)
	resp, err := client.GetBookings(context.Background(), "user1")
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Equal(t, "bk1", resp.Bookings[0].ID)
}

func TestAPIClient_CancelBooking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/bookings/bk1/cancel", r.URL.Path)
		assert.Equal(t, "PUT", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-api-key", 10*time.Second)
	err := client.CancelBooking(context.Background(), "bk1")
	assert.NoError(t, err)
}

func TestAPIClient_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.ErrorResponse{
			ErrorCode: "NOT_FOUND",
			Message:   "Resource not found",
		})
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-api-key", 10*time.Second)
	_, err := client.GetMerchants(context.Background())
	require.Error(t, err)

	var apiErr *APIError
	assert.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	assert.Equal(t, "NOT_FOUND", apiErr.ErrorCode)
}

func TestAPIClient_RetryOnServerError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.MerchantListResponse{Total: 0})
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-api-key", 10*time.Second)
	_, err := client.GetMerchants(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 3, attempts, "should retry twice before succeeding")
}

func TestAPIClient_NoRetryOnClientError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-api-key", 10*time.Second)
	_, err := client.GetMerchants(context.Background())
	require.Error(t, err)
	assert.Equal(t, 1, attempts, "should not retry on 4xx errors")
}

func TestAPIClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	client := NewAPIClient(server.URL, "test-api-key", 10*time.Second)
	_, err := client.GetMerchants(ctx)
	assert.Error(t, err)
}

func TestAPIClient_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "test-api-key", 50*time.Millisecond)
	_, err := client.GetMerchants(context.Background())
	require.Error(t, err)

	var netErr *NetworkError
	assert.ErrorAs(t, err, &netErr)
}

func TestAPIClient_XAPIKeyHeader(t *testing.T) {
	tests := []struct {
		name     string
		callFunc func(*APIClient) error
	}{
		{"GetMerchants", func(c *APIClient) error {
			_, err := c.GetMerchants(context.Background())
			return err
		}},
		{"GetServices", func(c *APIClient) error {
			_, err := c.GetServices(context.Background(), "m1")
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "test-api-key", r.Header.Get("X-API-Key"), tt.name)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{}"))
			}))
			defer server.Close()

			client := NewAPIClient(server.URL, "test-api-key", 10*time.Second)
			err := tt.callFunc(client)
			assert.NoError(t, err)
		})
	}
}
