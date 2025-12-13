// Package status provides the status command for viewing API rate limits.
package status

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// ClientFactory is a function that creates a Linear client.
type ClientFactory func() (*linear.Client, error)

// rateLimitCapture stores the most recent rate limit info from callbacks.
type rateLimitCapture struct {
	mu         sync.RWMutex
	info       *linear.RateLimitInfo
	capturedAt time.Time
}

func (r *rateLimitCapture) update(info *linear.RateLimitInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.info = info
	r.capturedAt = time.Now()
}

func (r *rateLimitCapture) get() (*linear.RateLimitInfo, time.Time) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.info, r.capturedAt
}

// NewStatusCommand creates the status command.
func NewStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show Linear API rate limit status",
		Long: `Display current rate limit quotas and usage.

Makes a lightweight API call (Viewer query) to fetch current rate limit headers.

Output shows:
  - Request-based limits (remaining / total)
  - Complexity-based limits (remaining / total)
  - Reset times

Examples:
  # Show status in table format
  go-linear status

  # JSON format for scripting
  go-linear status --output=json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd)
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runStatus(cmd *cobra.Command) error {
	ctx := context.Background()

	// Get API key
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("LINEAR_API_KEY environment variable required")
	}

	// Capture rate limit info via callback
	capture := &rateLimitCapture{}

	// Create client with rate limit callback
	client, err := linear.NewClient(apiKey,
		linear.WithRateLimitCallback(capture.update),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	// Make a lightweight request to trigger rate limit headers
	_, err = client.Viewer(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch rate limit info: %w", err)
	}

	// Get captured info
	info, timestamp := capture.get()
	if info == nil {
		return fmt.Errorf("no rate limit information available")
	}

	// Format output
	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), info, true)
	case "table":
		displayTable(cmd, info, timestamp)
		return nil
	default:
		return fmt.Errorf("unknown output format: %s", output)
	}
}

func displayTable(cmd *cobra.Command, info *linear.RateLimitInfo, timestamp time.Time) {
	fmt.Fprintln(cmd.OutOrStdout(), "LINEAR API RATE LIMIT STATUS")
	fmt.Fprintln(cmd.OutOrStdout(), "")

	// Request-based limits
	reqPercent := float64(info.RequestsRemaining) / float64(info.RequestsLimit) * 100
	fmt.Fprintln(cmd.OutOrStdout(), "REQUEST-BASED LIMITS")
	fmt.Fprintf(cmd.OutOrStdout(), "  Remaining: %d / %d (%.1f%%)\n",
		info.RequestsRemaining, info.RequestsLimit, reqPercent)
	if !info.RequestsReset.IsZero() {
		resetIn := time.Until(info.RequestsReset)
		fmt.Fprintf(cmd.OutOrStdout(), "  Resets in: %s (at %s)\n",
			resetIn.Round(time.Second), info.RequestsReset.Format(time.RFC3339))
	}
	fmt.Fprintln(cmd.OutOrStdout(), "")

	// Complexity-based limits
	complexPercent := float64(info.ComplexityRemaining) / float64(info.ComplexityLimit) * 100
	fmt.Fprintln(cmd.OutOrStdout(), "COMPLEXITY-BASED LIMITS")
	fmt.Fprintf(cmd.OutOrStdout(), "  Remaining: %d / %d (%.1f%%)\n",
		info.ComplexityRemaining, info.ComplexityLimit, complexPercent)
	if !info.ComplexityReset.IsZero() {
		resetIn := time.Until(info.ComplexityReset)
		fmt.Fprintf(cmd.OutOrStdout(), "  Resets in: %s (at %s)\n",
			resetIn.Round(time.Second), info.ComplexityReset.Format(time.RFC3339))
	}
	fmt.Fprintln(cmd.OutOrStdout(), "")

	// Metadata
	fmt.Fprintf(cmd.OutOrStdout(), "Last updated: %s\n", timestamp.Format(time.RFC3339))
}
