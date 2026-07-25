package models

// Merchant represents a merchant entity from the API
type Merchant struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	WelcomeMessage string `json:"welcome_message"`
}

// MerchantListResponse represents the API response for listing merchants
type MerchantListResponse struct {
	Total     int        `json:"total"`
	Merchants []Merchant `json:"merchants"`
}

// Service represents a service entity from the API
type Service struct {
	ID              string  `json:"id"`
	MerchantID      string  `json:"merchant_id"`
	Name            string  `json:"name"`
	DurationMinutes int     `json:"duration_minutes"`
	Price           float64 `json:"price,omitempty"`
}

// ServiceListResponse represents the API response for listing services
type ServiceListResponse struct {
	Total    int       `json:"total"`
	Services []Service `json:"services"`
}

// Staff represents a staff entity from the API
type Staff struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Services   []string `json:"services"`
	MerchantID string   `json:"merchant_id"`
}

// StaffListResponse represents the API response for listing staff
type StaffListResponse struct {
	Total int     `json:"total"`
	Staff []Staff `json:"staff"`
}

// Booking represents a booking entity from the API
type Booking struct {
	ID              string  `json:"id"`
	UserID          string  `json:"user_id"`
	ServiceID       string  `json:"service_id"`
	StaffID         string  `json:"staff_id"`
	StartTime       string  `json:"start_time"`
	DurationMinutes int     `json:"duration_minutes"`
	Price           float64 `json:"price,omitempty"`
	Paid            bool    `json:"paid"`
	Status          string  `json:"status"`
}

// BookingRequest is the request body for creating a booking
type BookingRequest struct {
	UserID          string  `json:"user_id"`
	ServiceID       string  `json:"service_id"`
	StaffID         string  `json:"staff_id"`
	StartTime       string  `json:"start_time"`
	DurationMinutes int     `json:"duration_minutes"`
	Price           float64 `json:"price,omitempty"`
}

// BookingListResponse represents the API response for listing bookings
type BookingListResponse struct {
	Total    int       `json:"total"`
	Bookings []Booking `json:"bookings"`
}

// AvailableDate represents an available date from the API
type AvailableDate struct {
	Date           string `json:"date"`
	SlotsAvailable int    `json:"slots_available"`
}

// AvailableSlot represents an available time slot from the API
type AvailableSlot struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
}
