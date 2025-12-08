package linear

import (
	"context"
	"net/http"
	"testing"
)

func TestClient_Organization(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"organization":{"id":"org123","name":"Test Org","urlKey":"testorg","logoUrl":"https://example.com/logo.png","createdAt":"2024-01-01T00:00:00.000Z","updatedAt":"2024-01-02T00:00:00.000Z","roadmapEnabled":true}}}`))
	})

	org, err := client.Organization(context.Background())
	if err != nil {
		t.Fatalf("Organization() error = %v", err)
	}

	if org.Name != "Test Org" {
		t.Errorf("Name = %q, want %q", org.Name, "Test Org")
	}
}

func TestClient_IssueLabel(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueLabel":{"id":"l123","name":"bug","description":"Bug label","color":"#FF0000","createdAt":"2024-01-01T00:00:00.000Z","updatedAt":"2024-01-02T00:00:00.000Z"}}}`))
	})

	label, err := client.IssueLabel(context.Background(), "l123")
	if err != nil {
		t.Fatalf("IssueLabel() error = %v", err)
	}

	if label.Name != "bug" {
		t.Errorf("Name = %q, want %q", label.Name, "bug")
	}
}

func TestClient_IssueLabels(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueLabels":{"nodes":[{"id":"l1","name":"bug","description":"Bugs","color":"#F00","createdAt":"2024-01-01T00:00:00.000Z"}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`))
	})

	labels, err := client.IssueLabels(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("IssueLabels() error = %v", err)
	}

	if len(labels.Nodes) != 1 {
		t.Errorf("len(Nodes) = %d, want 1", len(labels.Nodes))
	}
}

func TestClient_WorkflowState(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"workflowState":{"id":"ws123","name":"In Progress","description":"Work in progress","type":"started","color":"#F2C94C","position":2.0,"createdAt":"2024-01-01T00:00:00.000Z","updatedAt":"2024-01-02T00:00:00.000Z"}}}`))
	})

	state, err := client.WorkflowState(context.Background(), "ws123")
	if err != nil {
		t.Fatalf("WorkflowState() error = %v", err)
	}

	if state.Name != "In Progress" {
		t.Errorf("Name = %q, want %q", state.Name, "In Progress")
	}
}

func TestClient_WorkflowStates(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"workflowStates":{"nodes":[{"id":"ws1","name":"Todo","description":"To do","type":"unstarted","color":"#ccc","position":1.0,"createdAt":"2024-01-01T00:00:00.000Z"}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`))
	})

	states, err := client.WorkflowStates(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("WorkflowStates() error = %v", err)
	}

	if len(states.Nodes) != 1 {
		t.Errorf("len(Nodes) = %d, want 1", len(states.Nodes))
	}
}
