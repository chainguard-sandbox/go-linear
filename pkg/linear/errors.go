package linear

import (
	"fmt"
	"time"
)

// ErrorType represents the type of error returned by the Linear API.
type ErrorType string

// Error types matching the Linear API responses.
const (
	ErrorTypeInvalidInput        ErrorType = "InvalidInput"
	ErrorTypeRateLimited         ErrorType = "Ratelimited"
	ErrorTypeNetworkError        ErrorType = "NetworkError"
	ErrorTypeAuthenticationError ErrorType = "AuthenticationError"
	ErrorTypeForbidden           ErrorType = "Forbidden"
	ErrorTypeGraphQLError        ErrorType = "GraphqlError"
	ErrorTypeUnknown             ErrorType = "Unknown"
	ErrorTypeInternalError       ErrorType = "InternalError"
)

// LinearError represents an error from the Linear API.
type LinearError struct {
	Type          ErrorType
	Message       string
	StatusCode    int
	Query         string
	Variables     map[string]any
	TraceID       string
	GraphQLErrors []GraphQLError
	wrapped       error // Underlying error (for errors.Unwrap)
}

// Error implements the error interface.
func (e *LinearError) Error() string {
	if e.TraceID != "" {
		return fmt.Sprintf("linear: %s: %s [trace_id=%s]", e.Type, e.Message, e.TraceID)
	}
	return fmt.Sprintf("linear: %s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error for errors.Unwrap and errors.As support.
func (e *LinearError) Unwrap() error {
	return e.wrapped
}

// GraphQLError represents a GraphQL error from the API response.
type GraphQLError struct {
	Message    string
	Path       []any
	Extensions map[string]any
}

// RateLimitError represents a rate limiting error with retry information.
//
// The Linear API uses two types of rate limiting:
//  1. Request-based: Limits requests per time window
//  2. Complexity-based: Limits GraphQL query complexity per time window
//
// Both limits are tracked independently and reset at different times.
// Check both RequestsRemaining and ComplexityRemaining to determine capacity.
type RateLimitError struct {
	*LinearError
	RetryAfter int // Seconds until retry recommended

	// Request-based rate limiting
	RequestsLimit     int       // Total requests allowed in window
	RequestsRemaining int       // Requests remaining in current window
	RequestsResetAt   time.Time // When request limit resets

	// Complexity-based rate limiting
	ComplexityLimit     int       // Total complexity allowed in window
	ComplexityRemaining int       // Complexity remaining in current window
	ComplexityResetAt   time.Time // When complexity limit resets
}

// Error implements the error interface.
func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("%s (retry after %ds)", e.LinearError.Error(), e.RetryAfter)
	}
	return e.LinearError.Error()
}

// AuthenticationError represents an authentication error.
//
// This typically indicates:
//   - Invalid or expired API key
//   - Missing Authorization header
//   - Invalid OAuth token
type AuthenticationError struct {
	*LinearError
}

// ForbiddenError represents a forbidden/authorization error.
//
// This indicates the authenticated user lacks permission for the requested operation.
// Check API key scopes at https://linear.app/settings/account/security
type ForbiddenError struct {
	*LinearError
}

// wrapGraphQLError wraps a gqlgenc error in a LinearError with operation context.
// Preserves the original error for errors.As() and errors.Unwrap() support.
func wrapGraphQLError(operation string, err error) error {
	if err == nil {
		return nil
	}

	// gqlgenc wraps GraphQL errors - preserve chain for errors.As()
	return &LinearError{
		Type:       ErrorTypeGraphQLError,
		Message:    fmt.Sprintf("%s failed", operation),
		StatusCode: 200, // HTTP success but GraphQL error
		wrapped:    err,
	}
}

// errMutationFailed creates an error for mutations that return success=false.
func errMutationFailed(operation string) error {
	return &LinearError{
		Type:       ErrorTypeGraphQLError,
		Message:    fmt.Sprintf("%s: mutation returned success=false", operation),
		StatusCode: 200,
		wrapped:    nil, // No underlying error for success=false case
	}
}
