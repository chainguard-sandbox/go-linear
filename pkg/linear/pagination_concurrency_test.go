package linear

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
)

// TestIssueIterator_Concurrent verifies iterator is safe for concurrent use
func TestIssueIterator_Concurrent(t *testing.T) {
	var callCount atomic.Int32
	var issuesSeen sync.Map

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		page := callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")

		// Return 10 issues across 2 pages
		if page == 1 {
			_, _ = w.Write([]byte(`{"data":{"issues":{"nodes":[
				{"id":"i1","title":"Issue 1","description":"","priority":0,"createdAt":"2024-01-01T00:00:00.000Z","number":1,"url":"","state":{"id":"s1","name":"Todo"},"team":{"id":"t1","name":"Team","key":"T"}},
				{"id":"i2","title":"Issue 2","description":"","priority":0,"createdAt":"2024-01-01T00:00:00.000Z","number":2,"url":"","state":{"id":"s1","name":"Todo"},"team":{"id":"t1","name":"Team","key":"T"}},
				{"id":"i3","title":"Issue 3","description":"","priority":0,"createdAt":"2024-01-01T00:00:00.000Z","number":3,"url":"","state":{"id":"s1","name":"Todo"},"team":{"id":"t1","name":"Team","key":"T"}},
				{"id":"i4","title":"Issue 4","description":"","priority":0,"createdAt":"2024-01-01T00:00:00.000Z","number":4,"url":"","state":{"id":"s1","name":"Todo"},"team":{"id":"t1","name":"Team","key":"T"}},
				{"id":"i5","title":"Issue 5","description":"","priority":0,"createdAt":"2024-01-01T00:00:00.000Z","number":5,"url":"","state":{"id":"s1","name":"Todo"},"team":{"id":"t1","name":"Team","key":"T"}}
			],"pageInfo":{"hasNextPage":true,"endCursor":"cursor2"}}}}`))
			return
		}

		_, _ = w.Write([]byte(`{"data":{"issues":{"nodes":[
			{"id":"i6","title":"Issue 6","description":"","priority":0,"createdAt":"2024-01-01T00:00:00.000Z","number":6,"url":"","state":{"id":"s1","name":"Todo"},"team":{"id":"t1","name":"Team","key":"T"}},
			{"id":"i7","title":"Issue 7","description":"","priority":0,"createdAt":"2024-01-01T00:00:00.000Z","number":7,"url":"","state":{"id":"s1","name":"Todo"},"team":{"id":"t1","name":"Team","key":"T"}},
			{"id":"i8","title":"Issue 8","description":"","priority":0,"createdAt":"2024-01-01T00:00:00.000Z","number":8,"url":"","state":{"id":"s1","name":"Todo"},"team":{"id":"t1","name":"Team","key":"T"}},
			{"id":"i9","title":"Issue 9","description":"","priority":0,"createdAt":"2024-01-01T00:00:00.000Z","number":9,"url":"","state":{"id":"s1","name":"Todo"},"team":{"id":"t1","name":"Team","key":"T"}},
			{"id":"i10","title":"Issue 10","description":"","priority":0,"createdAt":"2024-01-01T00:00:00.000Z","number":10,"url":"","state":{"id":"s1","name":"Todo"},"team":{"id":"t1","name":"Team","key":"T"}}
		],"pageInfo":{"hasNextPage":false,"endCursor":"cursor3"}}}}`))
	})

	// Share ONE iterator across 5 concurrent goroutines
	iter := NewIssueIterator(client, 5)
	var wg sync.WaitGroup

	for i := range 5 {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for {
				issue, err := iter.Next(context.Background())
				if errors.Is(err, io.EOF) {
					return
				}
				if err != nil {
					t.Errorf("Worker %d: unexpected error: %v", workerID, err)
					return
				}

				// Track which issues we've seen (should be unique)
				if _, loaded := issuesSeen.LoadOrStore(issue.ID, workerID); loaded {
					t.Errorf("Worker %d: duplicate issue %s", workerID, issue.ID)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all 10 issues were seen exactly once
	count := 0
	issuesSeen.Range(func(key, value any) bool {
		count++
		return true
	})

	if count != 10 {
		t.Errorf("Saw %d unique issues, want 10", count)
	}
}

// TestIterator_Concurrency verifies iterators work correctly when shared
func TestIterator_Concurrency(t *testing.T) {
	// Simple test: verify mutex prevents panics, not full integration
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Return empty result to avoid GraphQL struct issues
		_, _ = w.Write([]byte(`{"data":{"issues":{"nodes":[],"pageInfo":{"hasNextPage":false}}}}`))
	})

	iter := NewIssueIterator(client, 10)
	var wg sync.WaitGroup

	// Multiple goroutines calling Next() concurrently
	// With mutex: no panic. Without mutex: data race/panic.
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 3 {
				_, err := iter.Next(context.Background())
				if err != nil {
					return
				}
			}
		}()
	}

	wg.Wait()
	// If we get here without panic, mutex is working
}

// TestTeamIterator_Concurrent verifies TeamIterator mutex protection
func TestTeamIterator_Concurrent(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"teams":{"nodes":[],"pageInfo":{"hasNextPage":false}}}}`))
	})

	iter := NewTeamIterator(client, 10)
	var wg sync.WaitGroup

	for range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = iter.Next(context.Background())
		}()
	}

	wg.Wait()
}

// TestProjectIterator_Concurrent verifies ProjectIterator mutex protection
func TestProjectIterator_Concurrent(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"projects":{"nodes":[],"pageInfo":{"hasNextPage":false}}}}`))
	})

	iter := NewProjectIterator(client, 10)
	var wg sync.WaitGroup

	for range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = iter.Next(context.Background())
		}()
	}

	wg.Wait()
}

// TestCommentIterator_Concurrent verifies CommentIterator mutex protection
func TestCommentIterator_Concurrent(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"comments":{"nodes":[],"pageInfo":{"hasNextPage":false}}}}`))
	})

	iter := NewCommentIterator(client, 10)
	var wg sync.WaitGroup

	for range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = iter.Next(context.Background())
		}()
	}

	wg.Wait()
}
