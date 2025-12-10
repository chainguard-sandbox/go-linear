package linear

import (
	"context"
	"net/http"
	"testing"
)

func TestClient_SearchIssues(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Return minimal valid search response
		_, _ = w.Write([]byte(`{"data":{"searchIssues":{"nodes":[],"pageInfo":{"hasNextPage":false,"endCursor":null},"totalCount":0}}}`))
	})

	first := int64(10)
	results, err := client.SearchIssues(context.Background(), "test query", &first, nil, nil, nil)
	if err != nil {
		t.Fatalf("SearchIssues() error = %v", err)
	}

	if results == nil {
		t.Fatal("SearchIssues() returned nil")
	}

	// Verify method is callable and returns expected type
	if results.PageInfo.HasNextPage {
		t.Error("Expected hasNextPage=false")
	}

	if results.TotalCount != 0 {
		t.Errorf("Expected totalCount=0, got %.0f", results.TotalCount)
	}
}
