package linear

import (
	"context"
	"net/http"
	"testing"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

func TestClient_TeamCreate(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"teamCreate":{"success":true,"team":{"id":"team-new","name":"New Team","key":"NEW","description":"Test team","createdAt":"2024-01-01T00:00:00.000Z"}}}}`))
	})

	name := "New Team"
	key := "NEW"
	desc := "Test team"

	team, err := client.TeamCreate(context.Background(), intgraphql.TeamCreateInput{
		Name:        name,
		Key:         &key,
		Description: &desc,
	})

	if err != nil {
		t.Fatalf("TeamCreate() error = %v", err)
	}

	if team.ID != "team-new" {
		t.Errorf("ID = %q, want %q", team.ID, "team-new")
	}
	if team.Name != "New Team" {
		t.Errorf("Name = %q, want %q", team.Name, "New Team")
	}
}

func TestClient_TeamUpdate(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"teamUpdate":{"success":true,"team":{"id":"team-123","name":"Updated Team","key":"UPD","description":"Updated","updatedAt":"2024-01-02T00:00:00.000Z"}}}}`))
	})

	updatedName := "Updated Team"
	_, err := client.TeamUpdate(context.Background(), "team-123", intgraphql.TeamUpdateInput{
		Name: &updatedName,
	})

	if err != nil {
		t.Fatalf("TeamUpdate() error = %v", err)
	}
}

func TestClient_TeamDelete(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"teamDelete":{"success":true}}}`))
	})

	err := client.TeamDelete(context.Background(), "team-123")
	if err != nil {
		t.Fatalf("TeamDelete() error = %v", err)
	}
}

func TestClient_TeamCreate_failureHandling(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"teamCreate":{"success":false}}}`))
	})

	name := "Team"
	_, err := client.TeamCreate(context.Background(), intgraphql.TeamCreateInput{
		Name: name,
	})

	if err == nil {
		t.Error("Expected error when success=false, got nil")
	}
}
