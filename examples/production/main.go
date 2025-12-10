// Package main demonstrates production-ready usage of the go-linear client.
//
// This example shows best practices for:
//   - Structured logging with slog
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

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Get API key from environment
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		slog.Error("LINEAR_API_KEY environment variable not set")
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
			slog.Info("rate limit status",
				"requests_remaining", info.RequestsRemaining,
				"requests_limit", info.RequestsLimit,
				"complexity_remaining", info.ComplexityRemaining,
				"complexity_limit", info.ComplexityLimit,
			)

			// Alert if close to rate limit
			if info.RequestsRemaining < info.RequestsLimit/10 {
				slog.Warn("approaching rate limit",
					"requests_remaining", info.RequestsRemaining)
			}
		}),
		linear.WithTLSConfig(&tls.Config{
			MinVersion: tls.VersionTLS12,
		}),
	)
	if err != nil {
		slog.Error("failed to create client", "error", err)
		os.Exit(1)
	}

	// Run main logic
	exitCode := run(client)

	// Cleanup
	if err := client.Close(); err != nil {
		slog.Warn("failed to close client", "error", err)
	}

	os.Exit(exitCode)
}

func run(client *linear.Client) int {
	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		slog.Info("received shutdown signal")
		cancel()
	}()

	// Verify authentication
	viewer, err := client.Viewer(ctx)
	if err != nil {
		slog.Error("authentication failed", "error", err)
		return 1
	}
	slog.Info("authenticated", "user", viewer.Email)

	// Example: List issues with pagination using iterator
	if err := listIssuesWithIterator(ctx, client); err != nil {
		slog.Error("failed to list issues", "error", err)
		return 1
	}

	// Example: Create issue with timeout
	if err := createIssueWithTimeout(ctx, client); err != nil {
		slog.Error("failed to create issue", "error", err)
		return 1
	}

	slog.Info("all operations completed successfully")
	return 0
}

func listIssuesWithIterator(ctx context.Context, client *linear.Client) error {
	slog.Info("listing issues with automatic pagination")

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
		slog.Debug("issue retrieved",
			"id", issue.ID,
			"title", issue.Title,
			"number", issue.Number)

		// Process issues...

		// Limit for demo
		if count >= 10 {
			break
		}
	}

	slog.Info("issues retrieved", "count", count)
	return nil
}

func createIssueWithTimeout(ctx context.Context, client *linear.Client) error {
	slog.Info("creating issue with 10s timeout")

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

	slog.Info("issue created",
		"id", issue.ID,
		"number", issue.Number,
		"url", issue.URL)

	return nil
}
