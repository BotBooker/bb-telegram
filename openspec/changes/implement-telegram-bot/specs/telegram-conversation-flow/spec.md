## ADDED Requirements

### Requirement: Start command with merchant selection
The system SHALL respond to `/start` by fetching the list of merchants from the API and presenting them as inline keyboard buttons. Each button's callback data SHALL encode the merchant ID. The system SHALL show each merchant's name and welcome message if one is set.

#### Scenario: User sends /start with multiple merchants available
- **WHEN** user sends `/start` and the API returns 3 merchants
- **THEN** the bot replies with a welcome message and an inline keyboard showing all 3 merchant names

#### Scenario: No merchants in the system
- **WHEN** user sends `/start` and the API returns an empty merchant list
- **THEN** the bot replies: "No merchants are currently available. Please check back later."

### Requirement: Book command initiates booking flow
The system SHALL respond to `/book` by presenting a merchant selection if not already selected, or proceeding directly to service selection if a merchant context exists. The booking flow follows this sequence: select merchant → select service → select date → select time slot → confirm booking.

#### Scenario: User sends /book without prior merchant selection
- **WHEN** user sends `/book` and no merchant is selected
- **THEN** the bot presents the merchant list as inline keyboard buttons

#### Scenario: User sends /book with merchant already selected
- **WHEN** user sends `/book` and a merchant is selected from a prior interaction
- **THEN** the bot proceeds directly to showing services for that merchant

### Requirement: Service browsing
The system SHALL fetch services for the selected merchant via `GET /catalog/services?merchant_id={id}` and present them as inline keyboard buttons. Each button SHALL show the service name, duration, and price (if set). Selecting a service SHALL advance to date selection.

#### Scenario: Merchant has multiple services
- **WHEN** user selects a merchant and the API returns services
- **THEN** the bot displays service names with duration and price as inline buttons

#### Scenario: Merchant has no services
- **WHEN** user selects a merchant and the API returns zero services
- **THEN** the bot replies: "This merchant has no services available yet. Please try another merchant."

#### Scenario: API call fails during service fetch
- **WHEN** the API call to `/catalog/services` fails
- **THEN** the bot replies: "Unable to load services at the moment. Please try again later." and logs the error

### Requirement: Date selection
The system SHALL fetch available dates for the selected service via `GET /availability/dates?service_id={id}` and present them as inline keyboard buttons. Dates with no available slots SHALL be omitted.

#### Scenario: Dates are available in the next 14 days
- **WHEN** user selects a service and the API returns 5 dates with available slots
- **THEN** the bot displays those 5 dates as formatted inline buttons (e.g., "Mon, Jul 10")

#### Scenario: No available dates
- **WHEN** user selects a service and the API returns empty dates
- **THEN** the bot replies: "No available dates found for this service. Please try another service or check back later."

### Requirement: Time slot selection
The system SHALL fetch available time slots for the selected date and service via `GET /availability/slots?service_id={id}&date={date}` and present them as inline keyboard buttons. Each slot SHALL display its start time.

#### Scenario: Time slots available on selected date
- **WHEN** user selects a date and the API returns available slots
- **THEN** the bot displays time slots as inline buttons in chronological order

#### Scenario: No available time slots
- **WHEN** user selects a date and the API returns empty slots
- **THEN** the bot replies: "No available time slots on this date. Please select a different date."

### Requirement: Staff selection
The system SHALL fetch staff members via `GET /catalog/staff?merchant_id={id}` and present them as inline keyboard buttons. Alternatively, if the API's availability slots already factor in staff, the system SHALL filter slots by passing `staff_id` to `GET /availability/slots`.

#### Scenario: Multiple staff members available
- **WHEN** user reaches staff selection step and the API returns staff members
- **THEN** the bot displays staff names as inline buttons

### Requirement: Booking confirmation
The system SHALL present a summary of the selected booking (merchant, service, staff, date, time, duration, price) and ask the user to confirm. Upon confirmation, the system SHALL call `POST /bookings` with the complete `BookingRequest` payload. On success, the system SHALL display a confirmation message with the booking ID. On 409 Conflict, the system SHALL inform the user the slot was taken and invite them to choose another slot.

#### Scenario: User confirms booking and it succeeds
- **WHEN** user taps "Confirm Booking" and the API returns 201 Created
- **THEN** the bot replies with a confirmation message including the booking ID, service name, date, and time

#### Scenario: Booking slot taken by another user
- **WHEN** user confirms booking and the API returns 409 Conflict
- **THEN** the bot replies: "Sorry, this slot is no longer available. Please choose another time." and offers available slots again

#### Scenario: API returns an unexpected error
- **WHEN** user confirms booking and the API returns 500
- **THEN** the bot replies: "Something went wrong while creating your booking. Please try again." and logs the error

### Requirement: My Bookings command
The system SHALL respond to `/mybookings` by fetching bookings for the user via `GET /bookings?user_id={chat_id}` and displaying them in a readable list. Each booking SHALL show service name, date, time, and status. If the user has no bookings, the system SHALL inform the user.

#### Scenario: User has active bookings
- **WHEN** user sends `/mybookings` and the API returns bookings
- **THEN** the bot displays each booking with service, date, time, and status

#### Scenario: User has no bookings
- **WHEN** user sends `/mybookings` and the API returns an empty list
- **THEN** the bot replies: "You have no bookings yet. Use /book to schedule one."

### Requirement: Cancel booking command
The system SHALL respond to `/cancel` by showing the user's active (non-cancelled) bookings as inline buttons. Selecting a booking SHALL call `PUT /bookings/{id}/cancel`. On success, the system SHALL display a cancellation confirmation.

#### Scenario: User cancels a booking successfully
- **WHEN** user selects a booking to cancel and the API returns 200
- **THEN** the bot replies: "Booking cancelled successfully."

#### Scenario: User tries to cancel a non-existent booking
- **WHEN** user selects a booking and the API returns 404
- **THEN** the bot replies: "This booking was not found. It may have already been cancelled."

### Requirement: Conversation cancellation
The system SHALL allow the user to cancel the current conversation flow at any step. This SHALL be supported via a "Cancel" inline button in each step, or by sending `/start` or `/book` to restart.

#### Scenario: User taps Cancel during service selection
- **WHEN** user taps "Cancel" during any step of the booking flow
- **THEN** the bot replies: "Booking cancelled." and returns to the main menu

### Requirement: Error handling for user-facing messages
The system SHALL respond to all API errors with user-friendly messages, never exposing raw error details or stack traces. The system SHALL log full error details server-side.

#### Scenario: Network error contacting the API
- **WHEN** an HTTP request to the API times out
- **THEN** the bot replies: "I'm having trouble reaching the booking service. Please try again in a moment." and logs the full error
