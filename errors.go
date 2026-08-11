package bachs

import (
	"errors"
	"fmt"
)

// Stable, machine-readable error codes returned by the Bachs API in the
// error_code field. Branch on these rather than on the human-readable Detail
// message, which may change between releases.
//
// See https://docs.bachs.io/api-reference/error-reference for the full list.
const (
	ErrCodeValidation          = "VALIDATION_ERROR"
	ErrCodeBadRequest          = "BAD_REQUEST"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeForbidden           = "FORBIDDEN"
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeConflict            = "CONFLICT"
	ErrCodeIdempotencyConflict = "IDEMPOTENCY_CONFLICT"
	ErrCodeTooManyRequests     = "TOO_MANY_REQUESTS"
	ErrCodeInternalServerError = "INTERNAL_SERVER_ERROR"
	ErrCodeBadGateway          = "BAD_GATEWAY"
	ErrCodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
)

// FieldError describes a single field that failed validation. It is present on
// APIError.Fields only for VALIDATION_ERROR responses.
type FieldError struct {
	// Field is the name of the field that failed validation, e.g. "amount".
	Field string `json:"field"`
	// Message is a human-readable description of the failure.
	Message string `json:"message"`
	// Type is a short error-type identifier, e.g. "missing" or "value_error".
	Type string `json:"type"`
}

// APIError is returned when the Bachs API responds with a non-2xx status. It
// carries the structured error body alongside transport-level context (HTTP
// status and the x-request-id, useful when contacting support).
//
// Inspect it with errors.As:
//
//	cust, err := client.Customers.Get(ctx, "cust_missing")
//	var apiErr *bachs.APIError
//	if errors.As(err, &apiErr) && apiErr.ErrorCode == bachs.ErrCodeNotFound {
//	    // ...
//	}
//
// or with the Is* helpers in this package.
type APIError struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int
	// ErrorCode is the stable, machine-readable code (the error_code field).
	// Compare it against the ErrCode* constants.
	ErrorCode string
	// Detail is the human-readable error message.
	Detail string
	// DocURL links to this code's entry in the error reference, when provided.
	DocURL string
	// Fields holds per-field validation failures. Populated only for
	// VALIDATION_ERROR responses; nil otherwise.
	Fields []FieldError
	// RequestID is the x-request-id response header. Include it when
	// contacting Bachs support.
	RequestID string
	// RetryAfter is the number of seconds the API asked the caller to wait
	// before retrying, parsed from the Retry-After header on a 429. Zero when
	// not present.
	RetryAfter int
	// Raw is the unparsed response body, retained for debugging when the body
	// did not match the standard error shape.
	Raw []byte
}

// Error implements the error interface.
func (e *APIError) Error() string {
	code := e.ErrorCode
	if code == "" {
		code = "unknown"
	}
	msg := fmt.Sprintf("bachs: %s (status %d, code %s)", e.Detail, e.StatusCode, code)
	if e.RequestID != "" {
		msg += fmt.Sprintf(" [request %s]", e.RequestID)
	}
	return msg
}

// errorEnvelope is the wire shape of the standard Bachs error object.
type errorEnvelope struct {
	Detail    string       `json:"detail"`
	ErrorCode string       `json:"error_code"`
	DocURL    string       `json:"doc_url"`
	Errors    []FieldError `json:"errors"`
}

// asAPIError extracts an *APIError from err, or returns nil if err is not one.
func asAPIError(err error) *APIError {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return nil
}

// hasCode reports whether err is an *APIError with the given error_code.
func hasCode(err error, code string) bool {
	apiErr := asAPIError(err)
	return apiErr != nil && apiErr.ErrorCode == code
}

// IsValidation reports whether err is a VALIDATION_ERROR (HTTP 422). When true,
// the offending fields are available on APIError.Fields.
func IsValidation(err error) bool { return hasCode(err, ErrCodeValidation) }

// IsNotFound reports whether err is a NOT_FOUND error (HTTP 404).
func IsNotFound(err error) bool { return hasCode(err, ErrCodeNotFound) }

// IsAuth reports whether err is an authentication or authorization failure
// (UNAUTHORIZED / 401 or FORBIDDEN / 403).
func IsAuth(err error) bool {
	return hasCode(err, ErrCodeUnauthorized) || hasCode(err, ErrCodeForbidden)
}

// IsConflict reports whether err is a CONFLICT or IDEMPOTENCY_CONFLICT (HTTP
// 409), such as a duplicate reference or an idempotency-key fingerprint
// mismatch.
func IsConflict(err error) bool {
	return hasCode(err, ErrCodeConflict) || hasCode(err, ErrCodeIdempotencyConflict)
}

// IsRateLimited reports whether err is a rate-limit error (HTTP 429). Read
// APIError.RetryAfter for how long to wait before retrying.
func IsRateLimited(err error) bool {
	apiErr := asAPIError(err)
	if apiErr == nil {
		return false
	}
	return apiErr.StatusCode == 429 || apiErr.ErrorCode == ErrCodeTooManyRequests
}
