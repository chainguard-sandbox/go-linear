// Package errors provides error types for the Linear API client.
package errors

import (
	"fmt"
	"time"
)

// ErrorType represents the type of error returned by the Linear API.
type ErrorType string

// Error types matching the Linear TypeScript SDK.
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
}

// Error implements the error interface.
func (e *LinearError) Error() string {
	if e.TraceID != "" {
		return fmt.Sprintf("linear: %s: %s [trace_id=%s]", e.Type, e.Message, e.TraceID)
	}
	return fmt.Sprintf("linear: %s: %s", e.Type, e.Message)
}

// GraphQLError represents a GraphQL error from the API response.
type GraphQLError struct {
	Message    string
	Path       []any
	Extensions map[string]any
}

// RateLimitError represents a rate limiting error with retry information.
type RateLimitError struct {
	*LinearError
	RetryAfter int

	// Request-based rate limiting
	RequestsLimit     int
	RequestsRemaining int
	RequestsResetAt   time.Time

	// Complexity-based rate limiting
	ComplexityLimit     int
	ComplexityRemaining int
	ComplexityResetAt   time.Time
}

// Error implements the error interface.
func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("%s (retry after %d seconds)", e.LinearError.Error(), e.RetryAfter)
	}
	return e.LinearError.Error()
}

// AuthenticationError represents an authentication error.
type AuthenticationError struct {
	*LinearError
}

// ForbiddenError represents a forbidden/authorization error.
type ForbiddenError struct {
	*LinearError
}
