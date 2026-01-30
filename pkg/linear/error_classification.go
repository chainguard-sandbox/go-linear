package linear

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/chainguard-dev/clog"
)

// ErrorClass categorizes errors by their characteristics for proper handling.
type ErrorClass int

const (
	// ErrorClassUser - User input error (fixable by user)
	// Examples: invalid team name, malformed date, missing required field
	// Action: Return clear message with correction guidance
	ErrorClassUser ErrorClass = iota

	// ErrorClassPermission - Authorization/permission error
	// Examples: API key lacks write permission, workspace access denied
	// Action: Return generic permission message, log details
	ErrorClassPermission

	// ErrorClassTransient - Temporary failure (retryable)
	// Examples: rate limits, 5xx errors, network timeouts
	// Action: Suggest retry, log for monitoring
	ErrorClassTransient

	// ErrorClassInternal - Internal service error (not user-fixable)
	// Examples: Linear API 500s, unexpected data format
	// Action: Generic message + support contact, log full details
	ErrorClassInternal

	// ErrorClassConfiguration - Configuration/setup error
	// Examples: invalid API key, network configuration
	// Action: Setup guidance message
	ErrorClassConfiguration
)

// ErrorSeverity indicates the impact level of an error.
type ErrorSeverity int

const (
	// SeverityInfo - Informational (non-blocking)
	SeverityInfo ErrorSeverity = iota
	// SeverityWarning - Warning (degraded but functional)
	SeverityWarning
	// SeverityError - Error (operation failed but recoverable)
	SeverityError
	// SeverityCritical - Critical (system-level failure)
	SeverityCritical
)

// ErrorContext holds rich context for internal logging while
// providing sanitized messages for external consumption.
//
//nolint:errname // ErrorContext is a context-holder that implements error, not a simple error type
type ErrorContext struct {
	// User-facing fields (safe to expose)
	Class    ErrorClass
	Severity ErrorSeverity
	Message  string // Sanitized, user-friendly message

	// Internal fields (for logging only)
	Operation  string         // e.g., "resolve_team", "issue_create"
	Input      string         // Sanitized input that caused error (no PII)
	Suggestion string         // User action to resolve (if applicable)
	TraceID    string         // For correlation with Linear API logs
	Internal   error          // Wrapped internal error (full details)
	Metadata   map[string]any // Additional debug context
}

// Error implements the error interface.
func (e *ErrorContext) Error() string {
	return e.Message
}

// Unwrap returns the wrapped error for errors.Is/As support.
func (e *ErrorContext) Unwrap() error {
	return e.Internal
}

// UserMessage returns a clean, actionable message for end users.
// Strips implementation details while providing guidance.
func (e *ErrorContext) UserMessage() string {
	msg := e.Message
	if e.Suggestion != "" {
		msg += ". " + e.Suggestion
	}
	return msg
}

// LogFields returns structured log fields for observability.
// Contains full context including internal details.
// SECURITY: Sanitizes credential-related errors before logging.
func (e *ErrorContext) LogFields() []any {
	fields := []any{
		"error_class", e.Class.String(),
		"error_severity", e.Severity.String(),
		"operation", e.Operation,
		"message", e.Message,
	}

	if e.Input != "" {
		fields = append(fields, "input", e.Input)
	}
	if e.TraceID != "" {
		fields = append(fields, "trace_id", e.TraceID)
	}
	if e.Internal != nil {
		// SECURITY: Sanitize internal error before logging
		// Never log API keys, tokens, or credentials
		sanitizedErr := sanitizeErrorForLogging(e.Internal.Error())
		fields = append(fields, "internal_error", sanitizedErr)
	}
	for k, v := range e.Metadata {
		fields = append(fields, k, v)
	}

	return fields
}

// sanitizeErrorForLogging removes credentials from error strings before logging.
// CRITICAL: API keys, tokens, passwords should NEVER appear in logs.
func sanitizeErrorForLogging(errStr string) string {
	// Pattern: lin_api_* (Linear API keys)
	if strings.Contains(errStr, "lin_api_") {
		// Replace entire key with placeholder
		// Use regex to handle any format: lin_api_xxx, lin_api_...
		re := regexp.MustCompile(`lin_api_[a-zA-Z0-9]+`)
		errStr = re.ReplaceAllString(errStr, "[REDACTED_API_KEY]")
	}

	// Pattern: Bearer tokens in Authorization headers
	if strings.Contains(errStr, "Bearer ") {
		re := regexp.MustCompile(`Bearer [a-zA-Z0-9\-_.]+`)
		errStr = re.ReplaceAllString(errStr, "Bearer [REDACTED_TOKEN]")
	}

	// Pattern: Authorization: <token>
	if strings.Contains(errStr, "Authorization:") || strings.Contains(errStr, "authorization:") {
		re := regexp.MustCompile(`(?i)authorization:\s*[^\s,]+`)
		errStr = re.ReplaceAllString(errStr, "Authorization: [REDACTED]")
	}

	// Pattern: token=xxx, apiKey=xxx, api_key=xxx
	if strings.Contains(errStr, "token=") || strings.Contains(errStr, "apiKey=") || strings.Contains(errStr, "api_key=") {
		re := regexp.MustCompile(`(?i)(token|apiKey|api_key)=\S+`)
		errStr = re.ReplaceAllString(errStr, "$1=[REDACTED]")
	}

	// Pattern: password=xxx, passwd=xxx
	if strings.Contains(errStr, "password=") || strings.Contains(errStr, "passwd=") {
		re := regexp.MustCompile(`(?i)(password|passwd)=\S+`)
		errStr = re.ReplaceAllString(errStr, "$1=[REDACTED]")
	}

	return errStr
}

func (c ErrorClass) String() string {
	switch c {
	case ErrorClassUser:
		return "user_error"
	case ErrorClassPermission:
		return "permission_error"
	case ErrorClassTransient:
		return "transient_error"
	case ErrorClassInternal:
		return "internal_error"
	case ErrorClassConfiguration:
		return "configuration_error"
	default:
		return "unknown_error"
	}
}

func (s ErrorSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// SanitizeError converts internal errors to user-safe ErrorContext.
// This is the primary entry point for error sanitization.
func SanitizeError(operation string, err error) *ErrorContext {
	if err == nil {
		return nil
	}

	// Handle already-sanitized errors
	var errCtx *ErrorContext
	if errors.As(err, &errCtx) {
		return errCtx
	}

	// Handle LinearError types
	var linearErr *LinearError
	if errors.As(err, &linearErr) {
		return sanitizeLinearError(operation, linearErr)
	}

	// Handle rate limit errors specially
	var rateLimitErr *RateLimitError
	if errors.As(err, &rateLimitErr) {
		return sanitizeRateLimitError(operation, rateLimitErr)
	}

	// Handle authentication errors
	var authErr *AuthenticationError
	if errors.As(err, &authErr) {
		return &ErrorContext{
			Class:      ErrorClassConfiguration,
			Severity:   SeverityError,
			Message:    "Authentication failed",
			Operation:  operation,
			Suggestion: "Check your API key at https://linear.app/settings/account/security",
			TraceID:    authErr.TraceID,
			Internal:   err,
		}
	}

	// Handle forbidden errors
	var forbiddenErr *ForbiddenError
	if errors.As(err, &forbiddenErr) {
		return &ErrorContext{
			Class:      ErrorClassPermission,
			Severity:   SeverityError,
			Message:    "Permission denied",
			Operation:  operation,
			Suggestion: "Your API key may lack required permissions",
			TraceID:    forbiddenErr.TraceID,
			Internal:   err,
		}
	}

	// Default: classify as internal error with generic message
	return &ErrorContext{
		Class:     ErrorClassInternal,
		Severity:  SeverityError,
		Message:   fmt.Sprintf("Operation failed: %s", operation),
		Operation: operation,
		Internal:  err,
	}
}

// sanitizeLinearError converts LinearError to ErrorContext.
func sanitizeLinearError(operation string, err *LinearError) *ErrorContext {
	ctx := &ErrorContext{
		Operation: operation,
		TraceID:   err.TraceID,
		Internal:  err,
	}

	switch err.Type {
	case ErrorTypeInvalidInput:
		ctx.Class = ErrorClassUser
		ctx.Severity = SeverityError
		ctx.Message = "Invalid input"
		ctx.Suggestion = "Check your command flags and try again"

	case ErrorTypeRateLimited:
		ctx.Class = ErrorClassTransient
		ctx.Severity = SeverityWarning
		ctx.Message = "Rate limit exceeded"
		ctx.Suggestion = "Wait a moment and try again"

	case ErrorTypeAuthenticationError:
		ctx.Class = ErrorClassConfiguration
		ctx.Severity = SeverityError
		ctx.Message = "Authentication failed"
		ctx.Suggestion = "Verify your API key"

	case ErrorTypeForbidden:
		ctx.Class = ErrorClassPermission
		ctx.Severity = SeverityError
		ctx.Message = "Permission denied"
		ctx.Suggestion = "Check API key permissions"

	case ErrorTypeNetworkError:
		ctx.Class = ErrorClassTransient
		ctx.Severity = SeverityError
		ctx.Message = "Network error"
		ctx.Suggestion = "Check your connection and retry"

	case ErrorTypeGraphQLError, ErrorTypeUnknown, ErrorTypeInternalError:
		ctx.Class = ErrorClassInternal
		ctx.Severity = SeverityError
		ctx.Message = "Request failed"
	}

	return ctx
}

// sanitizeRateLimitError provides specific guidance for rate limits.
func sanitizeRateLimitError(operation string, err *RateLimitError) *ErrorContext {
	suggestion := "Wait a moment and try again"
	if err.RetryAfter > 0 {
		suggestion = fmt.Sprintf("Retry after %d seconds", err.RetryAfter)
	}

	return &ErrorContext{
		Class:      ErrorClassTransient,
		Severity:   SeverityWarning,
		Message:    "Rate limit exceeded",
		Operation:  operation,
		Suggestion: suggestion,
		TraceID:    err.TraceID,
		Internal:   err,
		Metadata: map[string]any{
			"requests_remaining":   err.RequestsRemaining,
			"requests_limit":       err.RequestsLimit,
			"complexity_remaining": err.ComplexityRemaining,
			"retry_after":          err.RetryAfter,
		},
	}
}

// LogAndReturn logs the full error context and returns a sanitized error.
// This is the recommended pattern for error boundaries.
//
// Usage:
//
//	if err != nil {
//	    return linear.LogAndReturn(ctx, logger, "issue_create", err)
//	}
func LogAndReturn(ctx context.Context, logger *clog.Logger, operation string, err error) error {
	if err == nil {
		return nil
	}

	errCtx := SanitizeError(operation, err)

	// Log full context (including internal details)
	if logger != nil {
		logFunc := logger.ErrorContext
		switch errCtx.Severity {
		case SeverityInfo:
			logFunc = logger.InfoContext
		case SeverityWarning:
			logFunc = logger.WarnContext
		case SeverityError, SeverityCritical:
			// Use default ErrorContext
		}

		logFunc(ctx, errCtx.Message, errCtx.LogFields()...)
	}

	// Return sanitized error
	return errCtx
}
