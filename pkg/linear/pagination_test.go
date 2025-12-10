package linear

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
)

func TestIssueIterator(t *testing.T) {
	var callCount atomic.Int32

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		page := callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")

		// First page
		if page == 1 {
			_, _ = w.Write([]byte(`{"data":{"issues":{"nodes":[{"id":"i1","title":"Issue 1","description":"","priority":0,"createdAt":"2024-01-01T00:00:00.000Z","number":1,"url":"","state":{"id":"s1","name":"Todo"},"team":{"id":"t1","name":"Team","key":"T"}},{"id":"i2","title":"Issue 2","description":"","priority":0,"createdAt":"2024-01-01T00:00:00.000Z","number":2,"url":"","state":{"id":"s1","name":"Todo"},"team":{"id":"t1","name":"Team","key":"T"}}],"pageInfo":{"hasNextPage":true,"endCursor":"cursor2"}}}}`))
			return
		}

		// Second page (last)
		_, _ = w.Write([]byte(`{"data":{"issues":{"nodes":[{"id":"i3","title":"Issue 3","description":"","priority":0,"createdAt":"2024-01-01T00:00:00.000Z","number":3,"url":"","state":{"id":"s1","name":"Todo"},"team":{"id":"t1","name":"Team","key":"T"}}],"pageInfo":{"hasNextPage":false,"endCursor":"cursor3"}}}}`))
	})

	iter := NewIssueIterator(client, 2)
	var issues []string

	for {
		issue, err := iter.Next(context.Background())
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("iterator error: %v", err)
		}
		issues = append(issues, issue.Title)
	}

	if len(issues) != 3 {
		t.Errorf("got %d issues, want 3", len(issues))
	}

	want := []string{"Issue 1", "Issue 2", "Issue 3"}
	for i, title := range want {
		if issues[i] != title {
			t.Errorf("issue[%d] = %q, want %q", i, issues[i], title)
		}
	}
}

func TestTeamIterator_Creation(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Verify iterator can be created
	iter := NewTeamIterator(client, 50)
	if iter == nil {
		t.Fatal("NewTeamIterator() returned nil")
	}

	if iter.first != 50 {
		t.Errorf("pageSize = %d, want 50", iter.first)
	}

	// Default pageSize
	iter2 := NewTeamIterator(client, 0)
	if iter2.first != 50 {
		t.Errorf("default pageSize = %d, want 50", iter2.first)
	}
}

func TestProjectIterator_Creation(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Verify iterator can be created
	iter := NewProjectIterator(client, 100)
	if iter == nil {
		t.Fatal("NewProjectIterator() returned nil")
	}

	if iter.first != 100 {
		t.Errorf("pageSize = %d, want 100", iter.first)
	}
}

func TestCommentIterator_Creation(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Verify iterator can be created
	iter := NewCommentIterator(client, 25)
	if iter == nil {
		t.Fatal("NewCommentIterator() returned nil")
	}

	if iter.first != 25 {
		t.Errorf("pageSize = %d, want 25", iter.first)
	}
}
