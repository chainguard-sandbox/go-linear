//go:build write

package linear

import (
	"context"
	"os"
	"testing"

	intgraphql "github.com/eslerm/go-linear/internal/graphql"
)

// TestLive_CommentMutations tests comment create and delete.
// NOTE: Mindful of comment limits on test instances.
func TestLive_CommentMutations(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()

	// Get an existing issue to comment on
	issues, err := client.Issues(ctx, nil, nil)
	if err != nil || len(issues.Nodes) == 0 {
		t.Skip("No issues available for comment test")
	}

	issueID := issues.Nodes[0].ID

	// Create comment
	body := "[SDK-TEST] Test comment - will be deleted"

	comment, err := client.CommentCreate(ctx, intgraphql.CommentCreateInput{
		IssueID: &issueID,
		Body:    &body,
	})

	if err != nil {
		t.Fatalf("CommentCreate() error = %v", err)
	}

	t.Logf("✓ Created comment: %s", comment.ID)

	// Cleanup immediately
	if err := client.CommentDelete(ctx, comment.ID); err != nil {
		t.Errorf("Cleanup failed: %v", err)
	} else {
		t.Logf("✓ Cleaned up comment: %s", comment.ID)
	}
}
