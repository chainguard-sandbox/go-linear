package linear

import (
	"context"
	"os"
	"testing"
)

// TestLive_Issue tests Issue query against real Linear API.
// Safe read-only operation.
func TestLive_Issue(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()

	// First get an issue ID from the list
	issues, err := client.Issues(ctx, nil, nil)
	if err != nil {
		t.Fatalf("Issues() error = %v", err)
	}

	if len(issues.Nodes) == 0 {
		t.Skip("No issues found in workspace")
	}

	issueID := issues.Nodes[0].ID
	t.Logf("Testing with issue ID: %s", issueID)

	// Now fetch that specific issue
	issue, err := client.Issue(ctx, issueID)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if issue.ID != issueID {
		t.Errorf("ID = %q, want %q", issue.ID, issueID)
	}

	t.Logf("✓ Retrieved issue: %s", issue.Title)
}

// TestLive_Issues tests Issues query against real Linear API.
// Safe read-only operation.
func TestLive_Issues(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()

	// Get first 5 issues
	first := int64(5)
	issues, err := client.Issues(ctx, &first, nil)
	if err != nil {
		t.Fatalf("Issues() error = %v", err)
	}

	if issues == nil {
		t.Fatal("Issues() returned nil")
	}

	t.Logf("Retrieved %d issues", len(issues.Nodes))
	for i, issue := range issues.Nodes {
		t.Logf("  [%d] %s: %s", i+1, issue.ID, issue.Title)
	}
}
