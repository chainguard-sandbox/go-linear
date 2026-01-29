package resolver

import (
	"fmt"
	"strings"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// ResolutionError represents a name-to-ID resolution failure.
// Provides user-friendly messages while preserving internal context.
type ResolutionError struct {
	EntityType string // "team", "user", "label", etc.
	Input      string // What the user provided (safe to expose)
	Reason     string // User-facing reason
	Internal   error  // Underlying error (for logging)
}

// Error implements the error interface.
func (e *ResolutionError) Error() string {
	switch e.Reason {
	case "fetch failed":
		return fmt.Sprintf("failed to fetch %s", e.EntityType)
	case "empty input":
		return fmt.Sprintf("%s name/ID cannot be empty", e.EntityType)
	case "not found":
		return fmt.Sprintf("%s not found: %s", e.EntityType, e.Input)
	default:
		if e.Input != "" {
			return fmt.Sprintf("%s %s: %s", e.EntityType, e.Reason, e.Input)
		}
		return fmt.Sprintf("%s %s", e.EntityType, e.Reason)
	}
}

// Unwrap returns the wrapped error.
func (e *ResolutionError) Unwrap() error {
	return e.Internal
}

// ToErrorContext converts ResolutionError to linear.ErrorContext.
func (e *ResolutionError) ToErrorContext(operation string) *linear.ErrorContext {
	return &linear.ErrorContext{
		Class:      linear.ErrorClassUser,
		Severity:   linear.SeverityError,
		Message:    e.Error(),
		Operation:  operation,
		Input:      sanitizeInput(e.Input), // Remove potential PII
		Suggestion: e.suggestion(),
		Internal:   e.Internal,
	}
}

// suggestion provides actionable guidance based on error type.
func (e *ResolutionError) suggestion() string {
	// For fetch errors, suggest checking connectivity
	if e.Reason == "fetch failed" {
		return "Check your network connection and API key"
	}

	switch e.EntityType {
	case "team":
		return "List available teams with: linear team list"
	case "user":
		return "Check the user email or name. Use 'me' for yourself"
	case "label":
		return "List available labels with: linear label list"
	case "project":
		return "List available projects with: linear project list"
	case "workflow state":
		return "List available states with: linear state list"
	case "cycle":
		return "List available cycles with: linear cycle list"
	case "issue":
		return "Check the issue identifier (e.g., ENG-123)"
	case "initiative":
		return "List available initiatives with: linear initiative list"
	case "document":
		return "List available documents with: linear document list"
	default:
		return fmt.Sprintf("Verify the %s exists", e.EntityType)
	}
}

// sanitizeInput removes potential PII from user input.
// Preserves structure for debugging while protecting privacy.
func sanitizeInput(input string) string {
	// If input looks like email, redact domain but keep structure
	if len(input) > 0 && strings.Contains(input, "@") {
		return "***@***"
	}
	// For non-email, show first 3 chars + length
	if len(input) <= 3 {
		return input
	}
	return fmt.Sprintf("%s... (%d chars)", input[:3], len(input))
}

// newResolutionError creates a ResolutionError for "not found" cases.
func newNotFoundError(entityType, input string, fetchErr error) error {
	return &ResolutionError{
		EntityType: entityType,
		Input:      input,
		Reason:     "not found",
		Internal:   fetchErr,
	}
}

// newAmbiguousError creates a ResolutionError for ambiguous matches.
func newAmbiguousError(entityType, input string, matches []string) error {
	// Don't expose all matches (could contain PII)
	// Just indicate ambiguity and suggest being more specific
	matchCount := len(matches)
	return &ResolutionError{
		EntityType: entityType,
		Input:      input,
		Reason:     fmt.Sprintf("ambiguous (%d matches)", matchCount),
		Internal:   fmt.Errorf("ambiguous: %d matches found", matchCount),
	}
}

// newFetchError creates a ResolutionError for fetch failures.
func newFetchError(entityType string, fetchErr error) error {
	// Check if it's already a Linear API error
	var linearErr *linear.LinearError
	if err, ok := fetchErr.(*linear.LinearError); ok {
		linearErr = err
	}

	return &ResolutionError{
		EntityType: entityType,
		Input:      "",
		Reason:     "fetch failed",
		Internal:   linearErr,
	}
}
