package linear

import (
	"context"
	"net/http"
	"testing"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

func TestClient_IssueLabelCreate(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueLabelCreate":{"success":true,"issueLabel":{"id":"l123","name":"bug","description":"Bug label","color":"#FF0000","createdAt":"2024-01-01T00:00:00.000Z"}}}}`))
	})

	name := "bug"
	input := intgraphql.IssueLabelCreateInput{
		Name: name,
	}

	label, err := client.IssueLabelCreate(context.Background(), input)
	if err != nil {
		t.Fatalf("IssueLabelCreate() error = %v", err)
	}

	if label.ID != "l123" {
		t.Errorf("ID = %q, want %q", label.ID, "l123")
	}
}

func TestClient_IssueLabelUpdate(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueLabelUpdate":{"success":true,"issueLabel":{"id":"l123","name":"updated","description":"Updated","color":"#00FF00","updatedAt":"2024-01-02T00:00:00.000Z"}}}}`))
	})

	name := "updated"
	input := intgraphql.IssueLabelUpdateInput{
		Name: &name,
	}

	label, err := client.IssueLabelUpdate(context.Background(), "l123", input)
	if err != nil {
		t.Fatalf("IssueLabelUpdate() error = %v", err)
	}

	if label.Name != "updated" {
		t.Errorf("Name = %q, want %q", label.Name, "updated")
	}
}

func TestClient_IssueLabelDelete(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueLabelDelete":{"success":true}}}`))
	})

	err := client.IssueLabelDelete(context.Background(), "l123")
	if err != nil {
		t.Fatalf("IssueLabelDelete() error = %v", err)
	}
}
