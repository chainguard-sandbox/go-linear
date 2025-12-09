package linear

import (
	"context"
	"net/http"
	"testing"
)

func TestClient_Document(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"document":{"id":"d123","title":"API Docs","content":"# Documentation","createdAt":"2024-01-01T00:00:00.000Z","updatedAt":"2024-01-02T00:00:00.000Z","slugId":"api-docs","color":"#fff","icon":"📄"}}}`))
	})

	doc, err := client.Document(context.Background(), "d123")
	if err != nil {
		t.Fatalf("Document() error = %v", err)
	}

	if doc.Title != "API Docs" {
		t.Errorf("Title = %q, want %q", doc.Title, "API Docs")
	}
}

func TestClient_Documents(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"documents":{"nodes":[{"id":"d1","title":"Doc","content":"","createdAt":"2024-01-01T00:00:00.000Z","updatedAt":"2024-01-02T00:00:00.000Z","slugId":"doc","color":"","icon":""}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`))
	})

	docs, err := client.Documents(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Documents() error = %v", err)
	}

	if len(docs.Nodes) != 1 {
		t.Errorf("len(Nodes) = %d, want 1", len(docs.Nodes))
	}
}

func TestClient_Template(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"template":{"id":"t123","name":"Bug Template","description":"Template","createdAt":"2024-01-01T00:00:00.000Z","updatedAt":"2024-01-02T00:00:00.000Z","type":"issue"}}}`))
	})

	tpl, err := client.Template(context.Background(), "t123")
	if err != nil {
		t.Fatalf("Template() error = %v", err)
	}

	if tpl.Name != "Bug Template" {
		t.Errorf("Name = %q, want %q", tpl.Name, "Bug Template")
	}
}

func TestClient_Templates(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"templates":[{"id":"t1","name":"Bug","description":"","createdAt":"2024-01-01T00:00:00.000Z","type":"issue"}]}}`))
	})

	tpls, err := client.Templates(context.Background())
	if err != nil {
		t.Fatalf("Templates() error = %v", err)
	}

	if len(tpls) != 1 {
		t.Errorf("len(Templates) = %d, want 1", len(tpls))
	}
}
