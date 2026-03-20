// Package main demonstrates production-ready usage of the go-linear client.
//
// This example shows best practices for:
//   - Structured logging with clog
//   - Automatic retry with exponential backoff
//   - Rate limit monitoring
//   - Graceful shutdown
//   - Context cancellation and timeouts
//   - Error handling
package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chainguard-dev/clog"

	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

func main() {
	// Setup structured logging with clog (wraps slog)
	logger := clog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx := context.Background()

	// Get API key from environment
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		logger.ErrorContext(ctx, "LINEAR_API_KEY environment variable not set")
		os.Exit(1)
	}

	// Create production client with:
	// - Structured logging
	// - Retry with exponential backoff (5 retries, 500ms-60s backoff)
	// - Rate limit monitoring callback
	// - TLS 1.2+ enforcement
	client, err := linear.NewClient(apiKey,
		linear.WithLogger(logger),
		linear.WithRetry(5, 500*time.Millisecond, 60*time.Second),
		linear.WithTimeout(30*time.Second),
		linear.WithRateLimitCallback(func(info *linear.RateLimitInfo) {
			// Monitor rate limits - integrate with metrics system
			logger.InfoContext(ctx, "rate limit status",
				"requests_remaining", info.RequestsRemaining,
				"requests_limit", info.RequestsLimit,
				"complexity_remaining", info.ComplexityRemaining,
				"complexity_limit", info.ComplexityLimit,
			)

			// Alert if close to rate limit
			if info.RequestsRemaining < info.RequestsLimit/10 {
				logger.WarnContext(ctx, "approaching rate limit",
					"requests_remaining", info.RequestsRemaining)
			}
		}),
		linear.WithTLSConfig(&tls.Config{
			MinVersion: tls.VersionTLS12,
		}),
	)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create client", "error", err)
		os.Exit(1)
	}

	// Run main logic
	exitCode := run(ctx, logger, client)

	// Cleanup
	if err := client.Close(); err != nil {
		logger.WarnContext(ctx, "failed to close client", "error", err)
	}

	os.Exit(exitCode)
}

func run(ctx context.Context, logger *clog.Logger, client *linear.Client) int {
	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.InfoContext(ctx, "received shutdown signal")
		cancel()
	}()

	// Verify authentication
	viewer, err := client.Viewer(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "authentication failed", "error", err)
		return 1
	}
	logger.InfoContext(ctx, "authenticated", "user", viewer.Email)

	// Example: List issues with pagination using iterator
	if err := listIssuesWithIterator(ctx, logger, client); err != nil {
		logger.ErrorContext(ctx, "failed to list issues", "error", err)
		return 1
	}

	// Example: Create issue with timeout
	if err := createIssueWithTimeout(ctx, logger, client); err != nil {
		logger.ErrorContext(ctx, "failed to create issue", "error", err)
		return 1
	}

	logger.InfoContext(ctx, "all operations completed successfully")
	return 0
}

func listIssuesWithIterator(ctx context.Context, logger *clog.Logger, client *linear.Client) error {
	logger.InfoContext(ctx, "listing issues with automatic pagination")

	iter := linear.NewIssueIterator(client, 50)
	count := 0

	for {
		issue, err := iter.Next(ctx)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("iteration failed: %w", err)
		}

		count++
		logger.DebugContext(ctx, "issue retrieved",
			"id", issue.ID,
			"title", issue.Title,
			"number", issue.Number)

		// Process issues...

		// Limit for demo
		if count >= 10 {
			break
		}
	}

	logger.InfoContext(ctx, "issues retrieved", "count", count)
	return nil
}

func createIssueWithTimeout(ctx context.Context, logger *clog.Logger, client *linear.Client) error {
	logger.InfoContext(ctx, "creating issue with 10s timeout")

	// Create context with timeout for this specific operation
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Get team ID (in production, cache this)
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to get teams: %w", err)
	}
	if len(teams.Nodes) == 0 {
		return fmt.Errorf("no teams found")
	}
	teamID := teams.Nodes[0].ID

	title := "Production monitoring alert"
	description := "Automated issue created from production system"
	priority := int64(2) // High priority

	issue, err := client.IssueCreate(ctx, linear.IssueCreateInput{
		TeamID:      teamID,
		Title:       &title,
		Description: &description,
		Priority:    &priority,
	})
	if err != nil {
		// Use errors.As to check for specific error types
		var rateLimitErr *linear.RateLimitError
		if errors.As(err, &rateLimitErr) {
			return fmt.Errorf("rate limited, retry after %d seconds: %w",
				rateLimitErr.RetryAfter, err)
		}

		var authErr *linear.AuthenticationError
		if errors.As(err, &authErr) {
			return fmt.Errorf("authentication failed: %w", err)
		}

		return fmt.Errorf("create failed: %w", err)
	}

	logger.InfoContext(ctx, "issue created",
		"id", issue.ID,
		"number", issue.Number,
		"url", issue.URL)

	return nil
}
