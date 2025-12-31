package linear

import (
	"fmt"
	"slices"
	"strings"
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
//  1. Request-based: Limits requests per time window (~2 req/sec sustained)
//  2. Complexity-based: Limits GraphQL query complexity per time window
//
// Both limits are tracked independently and reset at different times.
// Check both RequestsRemaining and ComplexityRemaining to determine capacity.
//
// The client automatically retries 429 errors with exponential backoff.
// You can also implement custom retry logic using the RetryAfter field.
//
// Example:
//
//	_, err := client.Issues(ctx, &first, nil)
//	if err != nil {
//	    var rateLimitErr *linear.RateLimitError
//	    if errors.As(err, &rateLimitErr) {
//	        log.Printf("Rate limited. Retry after %d seconds", rateLimitErr.RetryAfter)
//	        log.Printf("Requests: %d/%d remaining", rateLimitErr.RequestsRemaining, rateLimitErr.RequestsLimit)
//	        time.Sleep(time.Duration(rateLimitErr.RetryAfter) * time.Second)
//	        // Retry request...
//	    }
//	}
//
// Monitor rate limits proactively:
//
//	client, _ := linear.NewClient(apiKey,
//	    linear.WithRateLimitCallback(func(info *linear.RateLimitInfo) {
//	        if info.RequestsRemaining < 10 {
//	            log.Warn("Approaching rate limit")
//	        }
//	    }),
//	)
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

// AuthenticationError represents an authentication error (401).
//
// This typically indicates:
//   - Invalid or expired API key
//   - Missing Authorization header
//   - Invalid OAuth token
//   - API key revoked in Linear settings
//
// With WithCredentialProvider, the client automatically refreshes
// credentials on 401 errors and retries the request once.
//
// Example:
//
//	viewer, err := client.Viewer(ctx)
//	if err != nil {
//	    var authErr *linear.AuthenticationError
//	    if errors.As(err, &authErr) {
//	        log.Error("Authentication failed - check API key")
//	        // Verify key at: https://linear.app/settings/account/security
//	        return err
//	    }
//	}
//
// Automatic refresh with credential provider:
//
//	provider := &MySecretsProvider{...}
//	client, _ := linear.NewClient("", linear.WithCredentialProvider(provider))
//	// On 401: client calls provider.GetCredential() and retries automatically
type AuthenticationError struct {
	*LinearError
}

// ForbiddenError represents a forbidden/authorization error (403).
//
// This indicates the authenticated user lacks permission for the requested operation.
// Common causes:
//   - API key has Read permission but Write permission required
//   - Attempting to modify resources in different workspace
//   - Organization-level restrictions
//
// Check and update API key scopes at:
// https://linear.app/settings/account/security
//
// Example:
//
//	_, err := client.IssueCreate(ctx, input)
//	if err != nil {
//	    var forbiddenErr *linear.ForbiddenError
//	    if errors.As(err, &forbiddenErr) {
//	        log.Error("Permission denied - API key needs Write permission")
//	        return err
//	    }
//	}
type ForbiddenError struct {
	*LinearError
}

// Error pattern groups for HTTP status detection.
var (
	authErrorPatterns      = []string{`"code":401`, "AUTHENTICATION_ERROR"}
	forbiddenErrorPatterns = []string{`"code":403`, "FORBIDDEN"}
	rateLimitErrorPatterns = []string{`"code":429`, "RATELIMITED"}
)

// containsAny checks if value contains any of the patterns.
func containsAny(value string, patterns []string) bool {
	return slices.ContainsFunc(patterns, func(pattern string) bool {
		return strings.Contains(value, pattern)
	})
}

// wrapGraphQLError wraps a gqlgenc error in a LinearError with operation context.
// Extracts user-friendly error messages from verbose gqlgenc responses.
func wrapGraphQLError(operation string, err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Extract clean error message from gqlgenc JSON errors
	// gqlgenc returns: {"networkErrors":{"code":401,"message":"Response body {...}"},...}
	// We want to extract the actual GraphQL error message for better UX

	// Check for authentication errors (401)
	if containsAny(errStr, authErrorPatterns) {
		return &AuthenticationError{
			LinearError: &LinearError{
				Type:       ErrorTypeAuthenticationError,
				Message:    "invalid or expired API key",
				StatusCode: 401,
				wrapped:    err,
			},
		}
	}

	// Check for forbidden errors (403)
	if containsAny(errStr, forbiddenErrorPatterns) {
		return &ForbiddenError{
			LinearError: &LinearError{
				Type:       ErrorTypeForbidden,
				Message:    "permission denied - check API key scopes",
				StatusCode: 403,
				wrapped:    err,
			},
		}
	}

	// Check for rate limit errors (429)
	if containsAny(errStr, rateLimitErrorPatterns) {
		return &LinearError{
			Type:       ErrorTypeRateLimited,
			Message:    "rate limit exceeded",
			StatusCode: 429,
			wrapped:    err,
		}
	}

	// For other errors, just use operation context without verbose JSON
	return &LinearError{
		Type:       ErrorTypeGraphQLError,
		Message:    fmt.Sprintf("%s failed", operation),
		StatusCode: 200,
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
