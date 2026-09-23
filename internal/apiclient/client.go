package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"bookerbot-tgbot/internal/models"

	"github.com/rs/zerolog/log"
)

// APIError represents an error returned by the API service
type APIError struct {
	StatusCode int
	ErrorCode  string
	Message    string
	Details    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d (%s): %s", e.StatusCode, e.ErrorCode, e.Message)
}

// NetworkError represents a network-level error (timeout, connection refused, etc.)
type NetworkError struct {
	Err error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: %v", e.Err)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

// APIClient is an HTTP client for the newbookerbot API service
type APIClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL, apiKey string, timeout time.Duration) *APIClient {
	return &APIClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// GetMerchants fetches all merchants from the API
func (c *APIClient) GetMerchants(ctx context.Context) (*models.MerchantListResponse, error) {
	var resp models.MerchantListResponse
	err := c.doGet(ctx, "/api/v1/admin/merchants/list", &resp)
	if err != nil {
		return nil, fmt.Errorf("get merchants: %w", err)
	}
	return &resp, nil
}

// GetServices fetches services for a merchant from the API
func (c *APIClient) GetServices(ctx context.Context, merchantID string) (*models.ServiceListResponse, error) {
	var resp models.ServiceListResponse
	path := fmt.Sprintf("/api/v1/catalog/services?merchant_id=%s", merchantID)
	err := c.doGet(ctx, path, &resp)
	if err != nil {
		return nil, fmt.Errorf("get services: %w", err)
	}
	return &resp, nil
}

// GetStaff fetches staff for a merchant from the API
func (c *APIClient) GetStaff(ctx context.Context, merchantID string) (*models.StaffListResponse, error) {
	var resp models.StaffListResponse
	path := fmt.Sprintf("/api/v1/catalog/staff?merchant_id=%s", merchantID)
	err := c.doGet(ctx, path, &resp)
	if err != nil {
		return nil, fmt.Errorf("get staff: %w", err)
	}
	return &resp, nil
}

// GetAvailableDates fetches available dates for a service from the API
func (c *APIClient) GetAvailableDates(ctx context.Context, serviceID string) ([]models.AvailableDate, error) {
	var dates []models.AvailableDate
	path := fmt.Sprintf("/api/v1/availability/dates?service_id=%s", serviceID)
	err := c.doGet(ctx, path, &dates)
	if err != nil {
		return nil, fmt.Errorf("get available dates: %w", err)
	}
	return dates, nil
}

// GetAvailableSlots fetches available time slots for a service, date, and optional staff from the API
func (c *APIClient) GetAvailableSlots(ctx context.Context, serviceID, date, staffID string) ([]models.AvailableSlot, error) {
	var slots []models.AvailableSlot
	path := fmt.Sprintf("/api/v1/availability/slots?service_id=%s&date=%s", serviceID, date)
	if staffID != "" {
		path += "&staff_id=" + staffID
	}
	err := c.doGet(ctx, path, &slots)
	if err != nil {
		return nil, fmt.Errorf("get available slots: %w", err)
	}
	return slots, nil
}

// CreateBooking creates a new booking via the API
func (c *APIClient) CreateBooking(ctx context.Context, req *models.BookingRequest) (*models.Booking, error) {
	var booking models.Booking
	err := c.doPost(ctx, "/api/v1/bookings", req, &booking)
	if err != nil {
		return nil, fmt.Errorf("create booking: %w", err)
	}
	return &booking, nil
}

// GetBookings fetches bookings for a user from the API
func (c *APIClient) GetBookings(ctx context.Context, userID string) (*models.BookingListResponse, error) {
	var resp models.BookingListResponse
	path := fmt.Sprintf("/api/v1/bookings?user_id=%s", userID)
	err := c.doGet(ctx, path, &resp)
	if err != nil {
		return nil, fmt.Errorf("get bookings: %w", err)
	}
	return &resp, nil
}

// CancelBooking cancels a booking via the API
func (c *APIClient) CancelBooking(ctx context.Context, bookingID string) error {
	err := c.doPut(ctx, fmt.Sprintf("/api/v1/bookings/%s/cancel", bookingID), nil, nil)
	if err != nil {
		return fmt.Errorf("cancel booking: %w", err)
	}
	return nil
}

// doGet performs a GET request with retry for transient errors
func (c *APIClient) doGet(ctx context.Context, path string, result interface{}) error {
	return c.doRequestWithRetry(ctx, http.MethodGet, path, nil, result)
}

// doPost performs a POST request (no retry for non-idempotent writes)
func (c *APIClient) doPost(ctx context.Context, path string, body, result interface{}) error {
	return c.doRequest(ctx, http.MethodPost, path, body, result)
}

// doPut performs a PUT request (no retry for non-idempotent writes)
func (c *APIClient) doPut(ctx context.Context, path string, body, result interface{}) error {
	return c.doRequest(ctx, http.MethodPut, path, body, result)
}

// doRequestWithRetry retries GET requests on transient errors with exponential backoff
func (c *APIClient) doRequestWithRetry(ctx context.Context, method, path string, body, result interface{}) error {
	maxRetries := 3
	backoffs := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := c.doRequest(ctx, method, path, body, result)
		if err == nil {
			return nil
		}

		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode < 500 {
			// Don't retry client errors (4xx)
			return err
		}

		lastErr = err
		if attempt < maxRetries {
			log.Warn().Err(err).Str("path", path).Int("attempt", attempt+1).Msg("request failed, retrying")
			select {
			case <-time.After(backoffs[attempt]):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return lastErr
}

// doRequest performs a single HTTP request
func (c *APIClient) doRequest(ctx context.Context, method, path string, body, result interface{}) error {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return &NetworkError{Err: fmt.Errorf("create request: %w", err)}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &NetworkError{Err: fmt.Errorf("execute request: %w", err)}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &NetworkError{Err: fmt.Errorf("read response body: %w", err)}
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		apiErr.StatusCode = resp.StatusCode
		// Try to parse the error response
		var errResp models.ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil {
			apiErr.ErrorCode = errResp.ErrorCode
			apiErr.Message = errResp.Message
			apiErr.Details = errResp.Details
		} else {
			apiErr.Message = string(respBody)
		}
		return &apiErr
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return nil
}
