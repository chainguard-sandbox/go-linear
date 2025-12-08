package linear

import (
	"context"
	"net/http"
	"testing"
)

func TestClient_Cycle(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"cycle":{"id":"cyc123","number":5,"name":"Sprint 5","description":"Q1 Sprint","startsAt":"2024-01-01T00:00:00.000Z","endsAt":"2024-01-14T00:00:00.000Z","completedAt":"2024-01-14T00:00:00.000Z","createdAt":"2024-01-01T00:00:00.000Z","updatedAt":"2024-01-02T00:00:00.000Z","progress":0.75,"team":{"id":"t1","name":"Eng","key":"ENG"}}}}`))
	})

	cycle, err := client.Cycle(context.Background(), "cyc123")
	if err != nil {
		t.Fatalf("Cycle() error = %v", err)
	}

	if cycle.Number != 5 {
		t.Errorf("Number = %f, want %f", cycle.Number, 5.0)
	}
	if cycle.Name == nil || *cycle.Name != "Sprint 5" {
		name := ""
		if cycle.Name != nil {
			name = *cycle.Name
		}
		t.Errorf("Name = %q, want %q", name, "Sprint 5")
	}
}

func TestClient_Cycles(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"cycles":{"nodes":[{"id":"c1","number":1,"name":"Sprint 1","startsAt":"2024-01-01T00:00:00.000Z","endsAt":"2024-01-14T00:00:00.000Z","createdAt":"2024-01-01T00:00:00.000Z","progress":1.0}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`))
	})

	cycles, err := client.Cycles(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Cycles() error = %v", err)
	}

	if len(cycles.Nodes) != 1 {
		t.Errorf("len(Nodes) = %d, want 1", len(cycles.Nodes))
	}
}

func TestClient_Roadmap(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"roadmap":{"id":"rm123","name":"2024 Roadmap","description":"Annual roadmap","createdAt":"2024-01-01T00:00:00.000Z","updatedAt":"2024-01-02T00:00:00.000Z","slugId":"2024-roadmap"}}}`))
	})

	roadmap, err := client.Roadmap(context.Background(), "rm123")
	if err != nil {
		t.Fatalf("Roadmap() error = %v", err)
	}

	if roadmap.Name != "2024 Roadmap" {
		t.Errorf("Name = %q, want %q", roadmap.Name, "2024 Roadmap")
	}
}

func TestClient_Roadmaps(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"roadmaps":{"nodes":[{"id":"r1","name":"Roadmap 1","description":"","createdAt":"2024-01-01T00:00:00.000Z","updatedAt":"2024-01-02T00:00:00.000Z","slugId":"roadmap-1"}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`))
	})

	roadmaps, err := client.Roadmaps(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Roadmaps() error = %v", err)
	}

	if len(roadmaps.Nodes) != 1 {
		t.Errorf("len(Nodes) = %d, want 1", len(roadmaps.Nodes))
	}
}
