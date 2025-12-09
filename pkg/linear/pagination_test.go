package linear

import (
	"context"
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

	for iter.Next(context.Background()) {
		issue := iter.Issue()
		if issue != nil {
			issues = append(issues, issue.Title)
		}
	}

	if err := iter.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
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
