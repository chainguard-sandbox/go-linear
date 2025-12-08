package linear

import (
	"context"
	"net/http"
	"testing"

	intgraphql "github.com/eslerm/go-linear/internal/graphql"
)

func TestClient_IssueCreate(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueCreate":{"success":true,"issue":{"id":"new123","title":"New Issue","description":"Created","priority":0,"createdAt":"2024-01-01T00:00:00.000Z","number":1,"url":"https://linear.app/test/TEST-1","state":{"id":"s1","name":"Todo"},"team":{"id":"t1","name":"Team","key":"TEAM"}}}}}`))
	})

	teamID := "team-123"
	title := "Test Issue"
	input := intgraphql.IssueCreateInput{
		TeamID: teamID,
		Title:  &title,
	}

	issue, err := client.IssueCreate(context.Background(), input)
	if err != nil {
		t.Fatalf("IssueCreate() error = %v", err)
	}

	if issue.ID != "new123" {
		t.Errorf("ID = %q, want %q", issue.ID, "new123")
	}
}

func TestClient_IssueUpdate(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueUpdate":{"success":true,"issue":{"id":"i123","title":"Updated","description":"Changed","priority":1,"updatedAt":"2024-01-02T00:00:00.000Z"}}}}`))
	})

	title := "Updated Title"
	input := intgraphql.IssueUpdateInput{
		Title: &title,
	}

	issue, err := client.IssueUpdate(context.Background(), "i123", input)
	if err != nil {
		t.Fatalf("IssueUpdate() error = %v", err)
	}

	if issue.Title != "Updated" {
		t.Errorf("Title = %q, want %q", issue.Title, "Updated")
	}
}

func TestClient_IssueDelete(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueDelete":{"success":true}}}`))
	})

	err := client.IssueDelete(context.Background(), "i123")
	if err != nil {
		t.Fatalf("IssueDelete() error = %v", err)
	}
}
