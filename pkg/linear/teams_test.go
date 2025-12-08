package linear

import (
	"context"
	"net/http"
	"testing"
)

func TestClient_Team(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"team":{"id":"t123","name":"Engineering","key":"ENG","description":"Eng team","createdAt":"2024-01-01T00:00:00.000Z","updatedAt":"2024-01-02T00:00:00.000Z","private":false,"timezone":"UTC","icon":"🔧","color":"#FF6900"}}}`))
	})

	team, err := client.Team(context.Background(), "t123")
	if err != nil {
		t.Fatalf("Team() error = %v", err)
	}

	if team.ID != "t123" {
		t.Errorf("ID = %q, want %q", team.ID, "t123")
	}
	if team.Key != "ENG" {
		t.Errorf("Key = %q, want %q", team.Key, "ENG")
	}
}

func TestClient_Teams(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"teams":{"nodes":[{"id":"t1","name":"Eng","key":"ENG","description":"Team","createdAt":"2024-01-01T00:00:00.000Z","private":false,"icon":"🔧","color":"#FF6900"}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`))
	})

	teams, err := client.Teams(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Teams() error = %v", err)
	}

	if len(teams.Nodes) != 1 {
		t.Errorf("len(Nodes) = %d, want 1", len(teams.Nodes))
	}
}

func TestClient_Project(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"project":{"id":"p123","name":"Q1 Goals","description":"Goals","state":"started","createdAt":"2024-01-01T00:00:00.000Z","updatedAt":"2024-01-02T00:00:00.000Z","startedAt":"2024-01-01T00:00:00.000Z","targetDate":"2024-03-31","progress":0.5,"url":"https://linear.app/proj","icon":"📅","color":"#0F0","lead":{"id":"u1","name":"Lead","email":"lead@example.com"}}}}`))
	})

	project, err := client.Project(context.Background(), "p123")
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}

	if project.ID != "p123" {
		t.Errorf("ID = %q, want %q", project.ID, "p123")
	}
	if project.State != "started" {
		t.Errorf("State = %q, want %q", project.State, "started")
	}
}

func TestClient_Projects(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"projects":{"nodes":[{"id":"p1","name":"Q1","description":"Goals","state":"started","createdAt":"2024-01-01T00:00:00.000Z","updatedAt":"2024-01-02T00:00:00.000Z","targetDate":"2024-03-31","progress":0.5,"url":"https://linear.app/p1","icon":"📅","color":"#0F0"}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`))
	})

	projects, err := client.Projects(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Projects() error = %v", err)
	}

	if len(projects.Nodes) != 1 {
		t.Errorf("len(Nodes) = %d, want 1", len(projects.Nodes))
	}
}
