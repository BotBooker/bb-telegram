## ADDED Requirements

### Requirement: API client initialization
The system SHALL provide an `APIClient` struct initialized with the API base URL (from config) and an API key. The client SHALL be injected into bot handlers as a dependency.

#### Scenario: APIClient created with valid configuration
- **WHEN** `NewAPIClient(baseURL, apiKey)` is called with valid parameters
- **THEN** an `APIClient` instance is returned with the base URL stored and the API key set for all request headers

#### Scenario: APIClient created with empty base URL
- **WHEN** `NewAPIClient("", apiKey)` is called
- **THEN** an error is returned indicating the base URL is required

### Requirement: X-API-Key header on all requests
The system SHALL include the `X-API-Key` header with the configured API key on every outgoing HTTP request to the API service.

#### Scenario: POST request includes X-API-Key header
- **WHEN** the client sends a POST request to `/api/v1/bookings`
- **THEN** the request includes the `X-API-Key` header with the configured value

#### Scenario: GET request includes X-API-Key header
- **WHEN** the client sends a GET request to `/api/v1/catalog/services`
- **THEN** the request includes the `X-API-Key` header with the configured value

### Requirement: Typed request and response structs
The system SHALL define Go structs matching the API's JSON request/response schemas from `api/spec/openapi.yaml`. All API client methods SHALL accept and return these typed structs, not raw `[]byte` or `map[string]interface{}`.

#### Scenario: GetServices returns typed ServiceResponse
- **WHEN** `client.GetServices(ctx, merchantID)` is called
- **THEN** the response is unmarshalled into a `ServiceListResponse` struct with typed fields

#### Scenario: CreateBooking accepts typed BookingRequest
- **WHEN** `client.CreateBooking(ctx, bookingReq)` is called
- **THEN** the `BookingRequest` struct is serialized to JSON and sent in the request body

### Requirement: Error handling and status code mapping
The system SHALL map HTTP status codes to typed Go errors. 4xx responses SHALL parse the API's `Error` schema into an `APIError` struct. 5xx and network errors SHALL return distinct error types that can be checked with `errors.Is`.

#### Scenario: 404 response returns APIError
- **WHEN** the API returns 404 Not Found
- **THEN** the method returns an `APIError` with `StatusCode: 404`, `ErrorCode`, and `Message` fields populated

#### Scenario: Network timeout returns NetworkError
- **WHEN** the HTTP request times out
- **THEN** the method returns a `NetworkError` that wraps the underlying error

### Requirement: Retry with exponential backoff
The system SHALL retry idempotent GET requests on transient failures (5xx, network errors) up to 3 times with exponential backoff (1s, 2s, 4s). Non-idempotent requests (POST, PUT, DELETE) SHALL NOT be retried to avoid duplicate side effects.

#### Scenario: GET request retries on 503
- **WHEN** a GET request receives a 503 Service Unavailable
- **THEN** the client retries up to 3 times with 1s, 2s, and 4s delays between attempts

#### Scenario: POST request is not retried
- **WHEN** a POST request receives a 503 Service Unavailable
- **THEN** the client returns the error immediately without retrying

#### Scenario: GET request succeeds after one retry
- **WHEN** a GET request fails with 503 on the first attempt but succeeds on the second
- **THEN** the client returns the successful response and does not attempt a third retry

### Requirement: Request timeout
The system SHALL apply a configurable timeout (default 10 seconds) to all HTTP requests. Requests exceeding the timeout SHALL be cancelled and return a timeout error.

#### Scenario: Request exceeds timeout
- **WHEN** an API call takes longer than the configured timeout
- **THEN** the request is cancelled and a timeout error is returned

### Requirement: Context propagation
The system SHALL accept and propagate `context.Context` to all HTTP requests, enabling cancellation from upstream handlers.

#### Scenario: Context cancelled before request completes
- **WHEN** the parent context is cancelled while an API request is in-flight
- **THEN** the HTTP request is aborted and a context cancellation error is returned
