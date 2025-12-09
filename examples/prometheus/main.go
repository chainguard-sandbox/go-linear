// Package main demonstrates Prometheus metrics integration with go-linear.
//
// This example shows:
//   - Enabling Prometheus metrics collection
//   - Exposing metrics at /metrics endpoint
//   - Using the Linear client with automatic metric tracking
//   - Viewing metrics in Prometheus format
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/eslerm/go-linear/pkg/linear"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		log.Fatal("LINEAR_API_KEY environment variable not set")
	}

	// Enable Linear Prometheus metrics
	// Metrics are automatically registered with default Prometheus registry
	linear.EnableMetrics()

	// Start Prometheus metrics endpoint at :2112/metrics
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())

		server := &http.Server{
			Addr:         ":2112",
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		}

		log.Printf("Metrics available at http://localhost:2112/metrics")
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()

	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create client with metrics enabled
	client, err := linear.NewClient(apiKey,
		linear.WithLogger(logger),
		linear.WithMetrics(), // Enable metrics collection
		linear.WithRetry(3, 500*time.Millisecond, 30*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Run main logic
	if err := run(client); err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Cleanup
	if err := client.Close(); err != nil {
		log.Printf("Warning: failed to close client: %v", err)
	}
}

func run(client *linear.Client) error {
	ctx := context.Background()

	// Verify authentication (metrics tracked automatically)
	viewer, err := client.Viewer(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	log.Printf("Authenticated as: %s", viewer.Email)

	// Make some API calls to generate metrics
	log.Println("Making API calls to generate metrics...")

	// Get issues (tracked in linear_requests_total, linear_request_duration_seconds)
	first := int64(10)
	issues, err := client.Issues(ctx, &first, nil)
	if err != nil {
		log.Printf("Issues query failed: %v", err)
		// Error tracked in linear_errors_total
	} else {
		log.Printf("Retrieved %d issues", len(issues.Nodes))
	}

	// Get teams
	teams, err := client.Teams(ctx, &first, nil)
	if err != nil {
		log.Printf("Teams query failed: %v", err)
	} else {
		log.Printf("Retrieved %d teams", len(teams.Nodes))
	}

	// Metrics are now available at http://localhost:2112/metrics
	log.Println("\nMetrics available at http://localhost:2112/metrics")
	log.Println("\nExample metrics:")
	log.Println("  linear_requests_total{operation=\"graphql\",status_code=\"200\"} 3")
	log.Println("  linear_request_duration_seconds_sum{operation=\"graphql\"} 0.45")
	log.Println("  linear_rate_limit_remaining{limit_type=\"requests\"} 115")
	log.Println("")
	log.Println("Curl the endpoint to see all metrics:")
	log.Println("  curl http://localhost:2112/metrics | grep linear_")
	log.Println("")
	log.Println("In Prometheus, query with:")
	log.Println("  rate(linear_requests_total[5m])")
	log.Println("  histogram_quantile(0.95, rate(linear_request_duration_seconds_bucket[5m]))")
	log.Println("")
	log.Println("Press Ctrl+C to exit...")

	// Keep server running
	select {}
}
