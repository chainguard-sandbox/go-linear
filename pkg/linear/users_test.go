package linear

import (
	"context"
	"net/http"
	"testing"
)

func TestClient_User(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"user":{"id":"u123","name":"Test User","email":"test@example.com","displayName":"testuser","createdAt":"2024-01-01T00:00:00.000Z","updatedAt":"2024-01-02T00:00:00.000Z","admin":false,"active":true,"avatarUrl":""}}}`))
	})

	user, err := client.User(context.Background(), "u123")
	if err != nil {
		t.Fatalf("User() error = %v", err)
	}

	if user.ID != "u123" {
		t.Errorf("ID = %q, want %q", user.ID, "u123")
	}
}

func TestClient_Users(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"users":{"nodes":[{"id":"u1","name":"User","email":"u@example.com","displayName":"user","createdAt":"2024-01-01T00:00:00.000Z","admin":false,"active":true,"avatarUrl":""}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`))
	})

	users, err := client.Users(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Users() error = %v", err)
	}

	if len(users.Nodes) != 1 {
		t.Errorf("len(Nodes) = %d, want 1", len(users.Nodes))
	}
}

func TestClient_Comment(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"comment":{"id":"c123","body":"Test comment","createdAt":"2024-01-01T00:00:00.000Z","updatedAt":"2024-01-02T00:00:00.000Z","url":"https://linear.app/comment/c123","user":{"id":"u1","name":"User"},"issue":{"id":"i1","title":"Issue"}}}}`))
	})

	comment, err := client.Comment(context.Background(), "c123")
	if err != nil {
		t.Fatalf("Comment() error = %v", err)
	}

	if comment.ID != "c123" {
		t.Errorf("ID = %q, want %q", comment.ID, "c123")
	}
}

func TestClient_Comments(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"comments":{"nodes":[{"id":"c1","body":"Comment","createdAt":"2024-01-01T00:00:00.000Z","updatedAt":"2024-01-02T00:00:00.000Z","url":"https://linear.app/c1"}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`))
	})

	comments, err := client.Comments(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Comments() error = %v", err)
	}

	if len(comments.Nodes) != 1 {
		t.Errorf("len(Nodes) = %d, want 1", len(comments.Nodes))
	}
}
