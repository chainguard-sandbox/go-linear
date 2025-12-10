package linear

import (
	"errors"
	"fmt"
	"testing"
)

func TestLinearError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrapped := wrapGraphQLError("IssueCreate", originalErr)

	// Verify error chain is preserved
	if !errors.Is(wrapped, originalErr) {
		t.Error("errors.Is() should find original error in chain")
	}

	// Verify errors.As() works
	var linearErr *LinearError
	if !errors.As(wrapped, &linearErr) {
		t.Fatal("errors.As() should find LinearError")
	}

	// Verify Unwrap() returns original
	if unwrapped := linearErr.Unwrap(); !errors.Is(unwrapped, originalErr) {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

func TestWrapGraphQLError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		err       error
		wantType  ErrorType
		wantNil   bool
	}{
		{
			name:      "nil error",
			operation: "IssueCreate",
			err:       nil,
			wantNil:   true,
		},
		{
			name:      "network error",
			operation: "TeamUpdate",
			err:       errors.New("network timeout"),
			wantType:  ErrorTypeGraphQLError,
		},
		{
			name:      "wrapped error",
			operation: "ProjectDelete",
			err:       fmt.Errorf("wrapped: %w", errors.New("base")),
			wantType:  ErrorTypeGraphQLError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapGraphQLError(tt.operation, tt.err)

			if tt.wantNil {
				if got != nil {
					t.Errorf("wrapGraphQLError() = %v, want nil", got)
				}
				return
			}

			var linearErr *LinearError
			if !errors.As(got, &linearErr) {
				t.Fatal("expected LinearError")
			}

			if linearErr.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", linearErr.Type, tt.wantType)
			}

			// Verify original error is in chain
			if !errors.Is(got, tt.err) {
				t.Error("errors.Is() should find original error")
			}
		})
	}
}

func TestErrMutationFailed(t *testing.T) {
	err := errMutationFailed("CommentCreate")

	var linearErr *LinearError
	if !errors.As(err, &linearErr) {
		t.Fatal("expected LinearError")
	}

	if linearErr.Type != ErrorTypeGraphQLError {
		t.Errorf("Type = %v, want %v", linearErr.Type, ErrorTypeGraphQLError)
	}

	if linearErr.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", linearErr.StatusCode)
	}

	// Unwrap should return nil (no underlying error)
	if linearErr.Unwrap() != nil {
		t.Errorf("Unwrap() = %v, want nil", linearErr.Unwrap())
	}
}

func TestLinearError_ErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		err  *LinearError
		want string
	}{
		{
			name: "with trace ID",
			err: &LinearError{
				Type:    ErrorTypeGraphQLError,
				Message: "operation failed",
				TraceID: "trace-123",
			},
			want: "linear: GraphqlError: operation failed [trace_id=trace-123]",
		},
		{
			name: "without trace ID",
			err: &LinearError{
				Type:    ErrorTypeNetworkError,
				Message: "connection refused",
			},
			want: "linear: NetworkError: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}
