//go:build write

package linear

import (
	"context"
	"os"
	"testing"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// TestLive_IssueMutations_FullCycle tests complete create/update/delete cycle.
func TestLive_IssueMutations_FullCycle(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()

	// Get team ID
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil {
		t.Fatalf("Teams() error = %v", err)
	}
	if len(teams.Nodes) == 0 {
		t.Fatal("No teams found")
	}

	teamID := teams.Nodes[0].ID

	// Create test issue
	title := "[SDK-TEST] Integration Test Issue"
	desc := "Created by go-linear integration tests"

	issue, err := client.IssueCreate(ctx, intgraphql.IssueCreateInput{
		TeamID:      teamID,
		Title:       &title,
		Description: &desc,
	})

	if err != nil {
		t.Fatalf("IssueCreate() error = %v", err)
	}

	t.Logf("✓ Created issue: %s", issue.ID)

	// Update the issue
	updatedTitle := "[SDK-TEST] Updated Issue"
	_, err = client.IssueUpdate(ctx, issue.ID, intgraphql.IssueUpdateInput{
		Title: &updatedTitle,
	})

	if err != nil {
		t.Fatalf("IssueUpdate() error = %v", err)
	}

	t.Logf("✓ Updated issue")

	// Cleanup
	if err := client.IssueDelete(ctx, issue.ID, nil); err != nil {
		t.Errorf("Cleanup failed: %v", err)
	} else {
		t.Logf("✓ Cleaned up issue: %s", issue.ID)
	}
}

// TestLive_IssueCreate_Minimal tests creating an issue with minimal fields.
// Tests that omitempty works - nil fields should not be sent to API.
func TestLive_IssueCreate_Minimal(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()

	// Get team ID
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil {
		t.Fatalf("Teams() error = %v", err)
	}
	if len(teams.Nodes) == 0 {
		t.Fatal("No teams found")
	}

	teamID := teams.Nodes[0].ID
	title := "[SDK-TEST] Minimal Issue"

	// Create issue with only required fields
	// DueDate, SLAType, and other optional fields are nil
	// With omitempty, they should NOT be sent in JSON
	issue, err := client.IssueCreate(ctx, intgraphql.IssueCreateInput{
		TeamID: teamID,
		Title:  &title,
		// DueDate: nil - should be omitted from JSON
		// SLAType: nil - should be omitted from JSON
	})

	if err != nil {
		t.Fatalf("IssueCreate() error = %v", err)
	}

	if issue.ID == "" {
		t.Error("Created issue has empty ID")
	}

	t.Logf("✓ Created issue: %s (%s)", issue.Title, issue.ID)

	// Cleanup: Delete the test issue
	t.Cleanup(func() {
		if err := client.IssueDelete(ctx, issue.ID, nil); err != nil {
			t.Errorf("Cleanup failed: %v", err)
		} else {
			t.Logf("✓ Cleaned up issue: %s", issue.ID)
		}
	})
}
