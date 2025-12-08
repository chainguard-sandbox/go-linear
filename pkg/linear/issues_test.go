package linear

import (
	"context"
	"net/http"
	"testing"
)

func TestClient_Issue(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issue":{"id":"i123","title":"Test Issue","description":"Test","priority":2,"estimate":5,"createdAt":"2024-01-01T00:00:00.000Z","updatedAt":"2024-01-02T00:00:00.000Z","number":123,"url":"https://linear.app/test/TEST-123","state":{"id":"s1","name":"In Progress","type":"started"},"team":{"id":"t1","name":"Eng","key":"ENG"},"assignee":{"id":"u1","name":"User","displayName":"user"}}}}`))
	})

	issue, err := client.Issue(context.Background(), "i123")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if issue.ID != "i123" {
		t.Errorf("ID = %q, want %q", issue.ID, "i123")
	}
	if issue.Title != "Test Issue" {
		t.Errorf("Title = %q, want %q", issue.Title, "Test Issue")
	}
}

func TestClient_Issues(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issues":{"nodes":[{"id":"i1","title":"Issue 1","description":"Desc","priority":1,"createdAt":"2024-01-01T00:00:00.000Z","number":1,"url":"https://linear.app/test/TEST-1","state":{"id":"s1","name":"Todo"},"team":{"id":"t1","name":"Eng","key":"ENG"}}],"pageInfo":{"hasNextPage":true,"endCursor":"cursor1"}}}}`))
	})

	issues, err := client.Issues(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Issues() error = %v", err)
	}

	if len(issues.Nodes) != 1 {
		t.Errorf("len(Nodes) = %d, want 1", len(issues.Nodes))
	}
	if !issues.PageInfo.HasNextPage {
		t.Error("HasNextPage should be true")
	}
}
