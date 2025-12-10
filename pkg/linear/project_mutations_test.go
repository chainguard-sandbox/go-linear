package linear

import (
	"context"
	"net/http"
	"testing"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

func TestClient_ProjectCreate(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"projectCreate":{"success":true,"project":{"id":"proj-new","name":"New Project","description":"Test project","state":"planned","createdAt":"2024-01-01T00:00:00.000Z","url":"https://linear.app/proj"}}}}`))
	})

	name := "New Project"
	desc := "Test project"

	project, err := client.ProjectCreate(context.Background(), intgraphql.ProjectCreateInput{
		Name:        name,
		Description: &desc,
	})

	if err != nil {
		t.Fatalf("ProjectCreate() error = %v", err)
	}

	if project.ID != "proj-new" {
		t.Errorf("ID = %q, want %q", project.ID, "proj-new")
	}
	if project.Name != "New Project" {
		t.Errorf("Name = %q, want %q", project.Name, "New Project")
	}
}

func TestClient_ProjectUpdate(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"projectUpdate":{"success":true,"project":{"id":"proj-123","name":"Updated Project","description":"Updated","state":"started","updatedAt":"2024-01-02T00:00:00.000Z"}}}}`))
	})

	updatedName := "Updated Project"
	_, err := client.ProjectUpdate(context.Background(), "proj-123", intgraphql.ProjectUpdateInput{
		Name: &updatedName,
	})

	if err != nil {
		t.Fatalf("ProjectUpdate() error = %v", err)
	}
}

func TestClient_ProjectDelete(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"projectDelete":{"success":true}}}`))
	})

	err := client.ProjectDelete(context.Background(), "proj-123")
	if err != nil {
		t.Fatalf("ProjectDelete() error = %v", err)
	}
}

func TestClient_ProjectCreate_failureHandling(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"projectCreate":{"success":false}}}`))
	})

	name := "Project"
	_, err := client.ProjectCreate(context.Background(), intgraphql.ProjectCreateInput{
		Name: name,
	})

	if err == nil {
		t.Error("Expected error when success=false, got nil")
	}
}
