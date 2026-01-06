// Package main demonstrates performance profiling with go-linear.
//
// This example shows:
//   - CPU profiling for identifying hot paths
//   - Memory profiling for analyzing allocations
//   - HTTP pprof endpoints for live profiling
//   - Execution tracing for goroutine analysis
//
// Usage:
//
//	# Basic profiling (generates profiles and exits)
//	LINEAR_API_KEY=xxx go run examples/profiling/main.go
//
//	# Analyze CPU profile
//	go tool pprof cpu.prof
//	# Commands: top10, list functionName, web
//
//	# Analyze memory allocations
//	go tool pprof -alloc_space mem.prof
//
//	# Analyze memory in-use
//	go tool pprof -inuse_space mem.prof
//
//	# View execution trace
//	go tool trace trace.out
//
//	# Live profiling (server mode)
//	LINEAR_API_KEY=xxx PROFILE_MODE=server go run examples/profiling/main.go
//	# Then in another terminal:
//	go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30  # CPU
//	go tool pprof http://localhost:6060/debug/pprof/heap               # Memory
//	go tool pprof http://localhost:6060/debug/pprof/goroutine          # Goroutines
//	go tool pprof http://localhost:6060/debug/pprof/block              # Blocking
//
// What to Look For:
//
// CPU Profile:
//   - transport.RoundTrip - Should dominate (network I/O)
//   - json.Unmarshal - JSON deserialization overhead
//   - io.ReadAll - Request body buffering
//
// Memory Profile:
//   - Look for allocations in hot paths (per-request allocations)
//   - Check iterator buffer sizes in pagination
//   - Identify opportunities for sync.Pool
//
// Related: examples/prometheus/main.go (metrics), examples/production/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync"
	"time"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		log.Fatal("LINEAR_API_KEY environment variable not set")
	}

	// Check if we're in server mode (live profiling)
	if os.Getenv("PROFILE_MODE") == "server" {
		runServerMode(apiKey)
		return
	}

	// File-based profiling mode
	runFileMode(apiKey)
}

// runFileMode generates profile files and exits
func runFileMode(apiKey string) {
	log.Println("Starting profiling run (file mode)...")
	log.Println("This will generate cpu.prof, mem.prof, and trace.out")

	// Start CPU profiling
	cpuFile, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatalf("Failed to create CPU profile: %v", err)
	}
	defer cpuFile.Close()

	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		log.Fatalf("Failed to start CPU profiling: %v", err)
	}
	defer pprof.StopCPUProfile()

	// Start execution trace
	traceFile, err := os.Create("trace.out")
	if err != nil {
		log.Fatalf("Failed to create trace file: %v", err)
	}
	defer traceFile.Close()

	if err := trace.Start(traceFile); err != nil {
		log.Fatalf("Failed to start trace: %v", err)
	}
	defer trace.Stop()

	// Enable block profiling (for channel/mutex contention analysis)
	runtime.SetBlockProfileRate(1)

	// Enable mutex profiling
	runtime.SetMutexProfileFraction(1)

	// Create client
	logger := linear.NewLogger()
	client, err := linear.NewClient(apiKey,
		linear.WithLogger(logger),
		linear.WithMetrics(),
		linear.WithRetry(3, 500*time.Millisecond, 30*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Run workload
	if err := runWorkload(client); err != nil {
		log.Fatalf("Workload failed: %v", err)
	}

	// Write heap profile
	memFile, err := os.Create("mem.prof")
	if err != nil {
		log.Fatalf("Failed to create memory profile: %v", err)
	}
	defer memFile.Close()

	runtime.GC() // Force GC to get accurate memory stats
	if err := pprof.WriteHeapProfile(memFile); err != nil {
		log.Fatalf("Failed to write memory profile: %v", err)
	}

	log.Println("\n✓ Profiling complete!")
	log.Println("\nAnalyze profiles with:")
	log.Println("  go tool pprof cpu.prof")
	log.Println("  go tool pprof -alloc_space mem.prof")
	log.Println("  go tool trace trace.out")
	log.Println("\nUseful pprof commands:")
	log.Println("  top10          - Show top 10 functions by time/allocations")
	log.Println("  list funcName  - Show annotated source for function")
	log.Println("  web            - Generate call graph (requires graphviz)")
	log.Println("  -http=:8080    - Start web UI")
	log.Println("\nLook for:")
	log.Println("  - Unexpected allocations in hot paths")
	log.Println("  - Functions spending excessive time in transport layer")
	log.Println("  - Opportunities for buffer reuse (sync.Pool)")
}

// runServerMode starts HTTP server with pprof endpoints
func runServerMode(apiKey string) {
	log.Println("Starting profiling server on :6060")
	log.Println("\nProfile endpoints available:")
	log.Println("  http://localhost:6060/debug/pprof/")
	log.Println("\nCapture profiles with:")
	log.Println("  go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30")
	log.Println("  go tool pprof http://localhost:6060/debug/pprof/heap")
	log.Println("  go tool pprof http://localhost:6060/debug/pprof/goroutine")
	log.Println("  go tool pprof http://localhost:6060/debug/pprof/block")
	log.Println("\nOr use web UI:")
	log.Println("  go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30")
	log.Println("\nPress Ctrl+C to stop")

	// Enable profiling
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)

	// Start pprof server (net/http/pprof automatically registers handlers)
	go func() {
		log.Fatal(http.ListenAndServe(":6060", nil))
	}()

	// Create client
	logger := linear.NewLogger()
	client, err := linear.NewClient(apiKey,
		linear.WithLogger(logger),
		linear.WithMetrics(),
		linear.WithRetry(3, 500*time.Millisecond, 30*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Run continuous workload in background
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := runWorkload(client); err != nil {
				log.Printf("Workload error: %v", err)
			}
		}
	}()

	// Keep server running
	select {}
}

// runWorkload performs typical API operations to generate profile data
func runWorkload(client *linear.Client) error {
	ctx := context.Background()

	log.Println("Running workload...")

	// Authenticate (lightweight operation)
	viewer, err := client.Viewer(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	log.Printf("Authenticated as: %s", viewer.Email)

	// Sequential operations (typical API usage)
	log.Println("Performing sequential operations...")

	// Get issues (pagination - tests iterator performance)
	first := int64(50)
	issues, err := client.Issues(ctx, &first, nil)
	if err != nil {
		log.Printf("Issues query failed: %v", err)
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

	// Search issues (more complex query)
	query := "is:issue"
	searchResults, err := client.SearchIssues(ctx, query, &first, nil, nil, nil)
	if err != nil {
		log.Printf("Search query failed: %v", err)
	} else {
		log.Printf("Search returned %d results", len(searchResults.Nodes))
	}

	// Concurrent operations (tests parallel request handling)
	log.Println("Performing concurrent operations...")

	var wg sync.WaitGroup
	concurrency := 5

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			pageSize := int64(10)
			_, err := client.Issues(ctx, &pageSize, nil)
			if err != nil {
				log.Printf("Concurrent request %d failed: %v", id, err)
			}
		}(i)
	}

	wg.Wait()

	// Test iterator (tests buffer allocation patterns)
	log.Println("Testing iterator...")
	iter := linear.NewIssueIterator(client, 25)

	count := 0
	maxItems := 100 // Limit iterations for profiling

	for count < maxItems {
		issue, err := iter.Next(ctx)
		if err != nil {
			// io.EOF or error
			break
		}
		if issue != nil {
			count++
		}
	}
	log.Printf("Iterator processed %d issues", count)

	// Force GC to see memory patterns
	runtime.GC()

	log.Println("✓ Workload complete")
	return nil
}
