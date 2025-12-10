package linear

import (
	"context"
	"net/http"
	"testing"
)

func TestClient_IssueSearch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Return minimal valid search response
		_, _ = w.Write([]byte(`{"data":{"issueSearch":{"nodes":[],"pageInfo":{"hasNextPage":false}}}}`))
	})

	first := int64(10)
	results, err := client.IssueSearch(context.Background(), "test query", &first, nil)
	if err != nil {
		t.Fatalf("IssueSearch() error = %v", err)
	}

	if results == nil {
		t.Fatal("IssueSearch() returned nil")
	}

	// Verify method is callable and returns expected type
	if results.PageInfo.HasNextPage {
		t.Error("Expected hasNextPage=false")
	}
}
