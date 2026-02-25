package linear

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestErrorContext_UserMessage(t *testing.T) {
	tests := []struct {
		name string
		err  *ErrorContext
		want string
	}{
		{
			name: "message only",
			err: &ErrorContext{
				Message: "Something failed",
			},
			want: "Something failed",
		},
		{
			name: "message with suggestion",
			err: &ErrorContext{
				Message:    "Team not found",
				Suggestion: "List teams with: linear team list",
			},
			want: "Team not found. List teams with: linear team list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.UserMessage()
			if got != tt.want {
				t.Errorf("UserMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeError_LinearError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		wantClass     ErrorClass
		wantSeverity  ErrorSeverity
		wantMsgPrefix string
	}{
		{
			name: "invalid input",
			err: &LinearError{
				Type:    ErrorTypeInvalidInput,
				Message: "field validation failed",
			},
			wantClass:     ErrorClassUser,
			wantSeverity:  SeverityError,
			wantMsgPrefix: "Invalid input",
		},
		{
			name: "rate limited",
			err: &LinearError{
				Type:    ErrorTypeRateLimited,
				Message: "too many requests",
			},
			wantClass:     ErrorClassTransient,
			wantSeverity:  SeverityWarning,
			wantMsgPrefix: "Rate limit exceeded",
		},
		{
			name: "authentication error",
			err: &AuthenticationError{
				LinearError: &LinearError{
					Type:    ErrorTypeAuthenticationError,
					Message: "invalid token xyz123",
					TraceID: "trace-abc",
				},
			},
			wantClass:     ErrorClassConfiguration,
			wantSeverity:  SeverityError,
			wantMsgPrefix: "Authentication failed",
		},
		{
			name: "forbidden error",
			err: &ForbiddenError{
				LinearError: &LinearError{
					Type:    ErrorTypeForbidden,
					Message: "workspace access denied",
				},
			},
			wantClass:     ErrorClassPermission,
			wantSeverity:  SeverityError,
			wantMsgPrefix: "Permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeError("test_operation", tt.err)

			if result.Class != tt.wantClass {
				t.Errorf("Class = %v, want %v", result.Class, tt.wantClass)
			}
			if result.Severity != tt.wantSeverity {
				t.Errorf("Severity = %v, want %v", result.Severity, tt.wantSeverity)
			}
			if tt.wantMsgPrefix != "" && result.Message != tt.wantMsgPrefix {
				t.Errorf("Message = %q, want prefix %q", result.Message, tt.wantMsgPrefix)
			}
			// Internal error should always be preserved
			if result.Internal == nil {
				t.Error("Internal error should be preserved")
			}
		})
	}
}

func TestSanitizeError_RateLimitError(t *testing.T) {
	err := &RateLimitError{
		LinearError: &LinearError{
			Type:    ErrorTypeRateLimited,
			Message: "rate limit exceeded",
			TraceID: "trace-123",
		},
		RetryAfter:          30,
		RequestsRemaining:   0,
		RequestsLimit:       100,
		ComplexityRemaining: 500,
	}

	result := SanitizeError("test_op", err)

	// Should classify correctly
	if result.Class != ErrorClassTransient {
		t.Errorf("Class = %v, want %v", result.Class, ErrorClassTransient)
	}

	// Should provide retry guidance
	if result.Suggestion == "" {
		t.Error("Suggestion should not be empty for rate limit errors")
	}

	// Should preserve metadata for logging
	if result.Metadata == nil {
		t.Error("Metadata should be preserved")
	}
	if retryAfter, ok := result.Metadata["retry_after"].(int); !ok || retryAfter != 30 {
		t.Errorf("retry_after = %v, want 30", result.Metadata["retry_after"])
	}
}

func TestSanitizeError_GenericError(t *testing.T) {
	err := errors.New("some unexpected error")

	result := SanitizeError("test_operation", err)

	// Generic errors should be classified as internal
	if result.Class != ErrorClassInternal {
		t.Errorf("Class = %v, want %v", result.Class, ErrorClassInternal)
	}

	// Should provide generic message
	if result.Message != "Operation failed: test_operation" {
		t.Errorf("Message = %q, unexpected", result.Message)
	}

	// Should preserve original error
	if !errors.Is(result.Internal, err) {
		t.Error("Should preserve original error")
	}
}

func TestSanitizeError_AlreadySanitized(t *testing.T) {
	original := &ErrorContext{
		Class:     ErrorClassUser,
		Severity:  SeverityWarning,
		Message:   "Already sanitized",
		Operation: "original_op",
	}

	// Sanitizing an already-sanitized error should return it unchanged
	result := SanitizeError("new_op", original)

	if result != original {
		t.Error("Should return same ErrorContext if already sanitized")
	}
}

func TestErrorClass_String(t *testing.T) {
	tests := []struct {
		class ErrorClass
		want  string
	}{
		{ErrorClassUser, "user_error"},
		{ErrorClassPermission, "permission_error"},
		{ErrorClassTransient, "transient_error"},
		{ErrorClassInternal, "internal_error"},
		{ErrorClassConfiguration, "configuration_error"},
		{ErrorClass(999), "unknown_error"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.class.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrorSeverity_String(t *testing.T) {
	tests := []struct {
		severity ErrorSeverity
		want     string
	}{
		{SeverityInfo, "info"},
		{SeverityWarning, "warning"},
		{SeverityError, "error"},
		{SeverityCritical, "critical"},
		{ErrorSeverity(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.severity.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrorContext_LogFields(t *testing.T) {
	err := &ErrorContext{
		Class:     ErrorClassUser,
		Severity:  SeverityError,
		Message:   "Test error",
		Operation: "test_op",
		Input:     "test_input",
		TraceID:   "trace-123",
		Internal:  errors.New("internal details"),
		Metadata: map[string]any{
			"custom_field": "custom_value",
		},
	}

	fields := err.LogFields()

	// Check that all important fields are present
	expectedFields := map[string]bool{
		"error_class":    false,
		"error_severity": false,
		"operation":      false,
		"message":        false,
		"input":          false,
		"trace_id":       false,
		"internal_error": false,
		"custom_field":   false,
	}

	for i := 0; i < len(fields); i += 2 {
		key := fields[i].(string)
		if _, ok := expectedFields[key]; ok {
			expectedFields[key] = true
		}
	}

	for key, found := range expectedFields {
		if !found {
			t.Errorf("LogFields() missing key %q", key)
		}
	}
}

// TestErrorContext_Unwrap verifies error chain support
func TestErrorContext_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	wrappedErr := &ErrorContext{
		Message:  "Outer error",
		Internal: innerErr,
	}

	// Test errors.Is
	if !errors.Is(wrappedErr, innerErr) {
		t.Error("errors.Is should work with Unwrap")
	}

	// Test errors.As
	var target *ErrorContext
	if !errors.As(wrappedErr, &target) {
		t.Error("errors.As should work with ErrorContext")
	}
}

// TestSanitizeErrorForLogging_Credentials verifies credentials are NEVER logged
func TestSanitizeErrorForLogging_Credentials(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		mustNot  string // String that MUST NOT appear in output
		mustHave string // Redaction marker that MUST appear
	}{
		{
			name:     "Linear API key",
			input:    "authentication failed with key lin_api_abc123xyz789def456",
			mustNot:  "lin_api_abc123xyz789def456",
			mustHave: "[REDACTED_API_KEY]",
		},
		{
			name:     "Bearer token",
			input:    "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			mustNot:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			mustHave: "[REDACTED_TOKEN]",
		},
		{
			name:     "token parameter",
			input:    "request failed: token=abc123xyz789",
			mustNot:  "abc123xyz789",
			mustHave: "token=[REDACTED]",
		},
		{
			name:     "apiKey parameter",
			input:    "invalid apiKey=lin_api_secretkey123",
			mustNot:  "lin_api_secretkey123",
			mustHave: "[REDACTED",
		},
		{
			name:     "password in error",
			input:    "auth failed: password=super_secret_123",
			mustNot:  "super_secret_123",
			mustHave: "password=[REDACTED]",
		},
		{
			name:     "multiple credentials",
			input:    "failed with token=abc123 and lin_api_xyz789",
			mustNot:  "abc123",
			mustHave: "[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeErrorForLogging(tt.input)

			// CRITICAL: Must not contain the credential
			if strings.Contains(result, tt.mustNot) {
				t.Errorf("SECURITY FAILURE: Credential leaked in logs!\nInput: %s\nOutput: %s\nLeaked: %s",
					tt.input, result, tt.mustNot)
			}

			// Must contain redaction marker
			if !strings.Contains(result, tt.mustHave) {
				t.Errorf("Missing redaction marker\nInput: %s\nOutput: %s\nExpected: %s",
					tt.input, result, tt.mustHave)
			}
		})
	}
}

// TestErrorContext_LogFields_NoCredentials verifies LogFields never logs credentials
func TestErrorContext_LogFields_NoCredentials(t *testing.T) {
	// Create error with credential in internal error
	err := &ErrorContext{
		Message:   "Authentication failed",
		Operation: "test_auth",
		Internal:  errors.New("invalid API key: lin_api_SUPER_SECRET_KEY_123"),
	}

	fields := err.LogFields()

	// Convert fields to string for checking
	fieldsStr := fmt.Sprintf("%v", fields)

	// CRITICAL: API key must not appear in log fields
	if strings.Contains(fieldsStr, "lin_api_SUPER_SECRET_KEY_123") {
		t.Errorf("SECURITY FAILURE: API key leaked in LogFields!\nFields: %v", fields)
	}

	// Must contain redaction marker
	if !strings.Contains(fieldsStr, "[REDACTED_API_KEY]") {
		t.Errorf("Missing redaction in LogFields\nFields: %v", fields)
	}
}

// TestSanitizeErrorForLogging_PreservesStructure verifies non-credential data preserved
func TestSanitizeErrorForLogging_PreservesStructure(t *testing.T) {
	input := "request to api.linear.app failed: status=401, token=secret123, trace_id=abc-def-ghi"
	result := sanitizeErrorForLogging(input)

	// Should preserve non-credential parts
	if !strings.Contains(result, "api.linear.app") {
		t.Error("Should preserve endpoint URL")
	}
	if !strings.Contains(result, "status=401") {
		t.Error("Should preserve status code")
	}
	if !strings.Contains(result, "trace_id=abc-def-ghi") {
		t.Error("Should preserve trace ID")
	}

	// Should redact credential
	if strings.Contains(result, "secret123") {
		t.Error("Should redact token")
	}
}
