package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/chainguard-dev/clog"
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// ErrorHandler handles errors at the command boundary.
// Provides the interface between internal errors and user-facing output.
type ErrorHandler struct {
	logger *clog.Logger
	// IsMCP indicates if running in MCP mode (affects output format)
	IsMCP bool
}

// NewErrorHandler creates an error handler for commands.
func NewErrorHandler(logger *clog.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
		IsMCP:  false, // TODO: detect from environment or flag
	}
}

// Handle processes an error and returns a user-friendly error.
// Logs full context internally, returns sanitized message externally.
//
// Usage in commands:
//
//	func runCreate(cmd *cobra.Command, client *linear.Client) error {
//	    eh := cli.NewErrorHandler(getLogger(cmd))
//	    result, err := client.IssueCreate(ctx, input)
//	    if err != nil {
//	        return eh.Handle(cmd.Context(), "issue_create", err)
//	    }
//	    ...
//	}
func (h *ErrorHandler) Handle(ctx context.Context, operation string, err error) error {
	if err == nil {
		return nil
	}

	// Convert to ErrorContext for sanitization
	errCtx := h.toErrorContext(operation, err)

	// Log full details (including internal error)
	h.logError(ctx, errCtx)

	// Return sanitized error based on context
	if h.IsMCP {
		return h.mcpError(errCtx)
	}
	return h.cliError(errCtx)
}

// toErrorContext converts various error types to ErrorContext.
func (h *ErrorHandler) toErrorContext(operation string, err error) *linear.ErrorContext {
	// Handle resolver errors specially
	var resErr *resolver.ResolutionError
	if errors.As(err, &resErr) {
		return resErr.ToErrorContext(operation)
	}

	// Use linear's sanitization for other errors
	return linear.SanitizeError(operation, err)
}

// logError logs the full error context for observability.
func (h *ErrorHandler) logError(ctx context.Context, errCtx *linear.ErrorContext) {
	if h.logger == nil {
		return
	}

	// Choose log level based on severity
	logFunc := h.logger.ErrorContext
	switch errCtx.Severity {
	case linear.SeverityInfo:
		logFunc = h.logger.InfoContext
	case linear.SeverityWarning:
		logFunc = h.logger.WarnContext
	case linear.SeverityError, linear.SeverityCritical:
		// Use default ErrorContext
	}

	// Log with full context
	logFunc(ctx, errCtx.Message, errCtx.LogFields()...)
}

// cliError formats error for CLI output (human-readable).
func (h *ErrorHandler) cliError(errCtx *linear.ErrorContext) error {
	msg := errCtx.UserMessage()

	// Add emoji indicators for severity (optional, can be disabled)
	switch errCtx.Severity {
	case linear.SeverityInfo:
		// No emoji for info
	case linear.SeverityWarning:
		msg = "⚠️  " + msg
	case linear.SeverityError, linear.SeverityCritical:
		msg = "❌ " + msg
	}

	return fmt.Errorf("%s", msg)
}

// mcpError formats error for MCP output (AI agent-friendly).
// Strips emojis and formatting that could confuse LLMs.
func (h *ErrorHandler) mcpError(errCtx *linear.ErrorContext) error {
	// MCP output: plain text, structured
	msg := errCtx.UserMessage()

	// Add machine-readable prefix for classification
	// Helps AI agents understand error type
	switch errCtx.Class {
	case linear.ErrorClassUser:
		msg = "[INPUT_ERROR] " + msg
	case linear.ErrorClassPermission:
		msg = "[PERMISSION_ERROR] " + msg
	case linear.ErrorClassTransient:
		msg = "[RETRY_ERROR] " + msg
	case linear.ErrorClassConfiguration:
		msg = "[CONFIG_ERROR] " + msg
	case linear.ErrorClassInternal:
		msg = "[ERROR] " + msg
	}

	return fmt.Errorf("%s", msg)
}

// HandleWithSuggestion handles errors and adds custom suggestions.
// Useful for command-specific guidance.
//
// Usage:
//
//	return eh.HandleWithSuggestion(ctx, "issue_create", err,
//	    "Try specifying --team explicitly")
func (h *ErrorHandler) HandleWithSuggestion(ctx context.Context, operation string, err error, suggestion string) error {
	if err == nil {
		return nil
	}

	errCtx := h.toErrorContext(operation, err)
	// Override suggestion with command-specific one
	errCtx.Suggestion = suggestion

	h.logError(ctx, errCtx)

	if h.IsMCP {
		return h.mcpError(errCtx)
	}
	return h.cliError(errCtx)
}

// getLogger extracts logger from command context (if available).
func getLogger(cmd *cobra.Command) *clog.Logger {
	// TODO: Implement context-based logger extraction
	// For now, check if logger is in context
	if logger, ok := cmd.Context().Value("logger").(*clog.Logger); ok {
		return logger
	}
	return nil
}
