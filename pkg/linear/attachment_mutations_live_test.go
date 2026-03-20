//go:build write

package linear

import (
	"context"
	"os"
	"testing"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// TestLive_AttachmentMutations_FullCycle tests complete attachment lifecycle.
// Tests: create issue → attach URLs → list/filter attachments → delete attachments → cleanup issue
func TestLive_AttachmentMutations_FullCycle(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Setup: Get team ID
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil {
		t.Fatalf("Teams() error = %v", err)
	}
	if len(teams.Nodes) == 0 {
		t.Fatal("No teams found")
	}
	teamID := teams.Nodes[0].ID

	// Setup: Create test issue
	title := "[SDK-TEST] Attachment Integration Test"
	desc := "Created by go-linear attachment integration tests"

	issue, err := client.IssueCreate(ctx, intgraphql.IssueCreateInput{
		TeamID:      teamID,
		Title:       &title,
		Description: &desc,
	})
	if err != nil {
		t.Fatalf("IssueCreate() error = %v", err)
	}
	t.Logf("✓ Created test issue: %s", issue.ID)

	// Cleanup issue at end
	defer func() {
		if err := client.IssueDelete(ctx, issue.ID, nil); err != nil {
			t.Errorf("Failed to cleanup issue: %v", err)
		} else {
			t.Logf("✓ Cleaned up issue: %s", issue.ID)
		}
	}()

	// Test 1: Create custom attachment
	attachment1, err := client.AttachmentCreate(ctx, intgraphql.AttachmentCreateInput{
		IssueID:  issue.ID,
		Title:    "Test Doc 1",
		Subtitle: strPtr("First test document"),
		URL:      "https://example.com/doc1",
	})
	if err != nil {
		t.Fatalf("AttachmentCreate() error = %v", err)
	}
	t.Logf("✓ Created attachment 1: %s", attachment1.ID)

	// Test 2: Link URL attachment
	linkTitle := "Test Doc 2"
	attachment2, err := client.AttachmentLinkURL(ctx, issue.ID, "https://example.com/doc2", &linkTitle)
	if err != nil {
		t.Fatalf("AttachmentLinkURL() error = %v", err)
	}
	t.Logf("✓ Created attachment 2: %s", attachment2.ID)

	// Test 3: Create third attachment for filtering tests
	attachment3, err := client.AttachmentCreate(ctx, intgraphql.AttachmentCreateInput{
		IssueID: issue.ID,
		Title:   "Planning Document",
		URL:     "https://example.com/planning",
	})
	if err != nil {
		t.Fatalf("AttachmentCreate() error = %v", err)
	}
	t.Logf("✓ Created attachment 3: %s", attachment3.ID)

	// Test 4: Get single attachment
	retrieved, err := client.Attachment(ctx, attachment1.ID)
	if err != nil {
		t.Fatalf("Attachment() error = %v", err)
	}
	if retrieved.ID != attachment1.ID {
		t.Errorf("Attachment() returned wrong ID: got %s, want %s", retrieved.ID, attachment1.ID)
	}
	if retrieved.Title != "Test Doc 1" {
		t.Errorf("Attachment() returned wrong title: got %s, want %s", retrieved.Title, "Test Doc 1")
	}
	t.Logf("✓ Retrieved attachment by ID")

	// Test 5: List all attachments (should include our 3)
	first := int64(50)
	allAttachments, err := client.Attachments(ctx, &first, nil)
	if err != nil {
		t.Fatalf("Attachments() error = %v", err)
	}

	foundCount := 0
	for _, att := range allAttachments.Nodes {
		if att.ID == attachment1.ID || att.ID == attachment2.ID || att.ID == attachment3.ID {
			foundCount++
		}
	}
	if foundCount != 3 {
		t.Errorf("Attachments() found %d/3 test attachments", foundCount)
	}
	t.Logf("✓ Listed attachments, found %d total (%d test attachments)", len(allAttachments.Nodes), foundCount)

	// Test 6: Filter by title
	titleFilter := "Planning"
	filterBuilder := &intgraphql.AttachmentFilter{
		Title: &intgraphql.StringComparator{
			ContainsIgnoreCase: &titleFilter,
		},
	}

	filtered, err := client.AttachmentsFiltered(ctx, &first, nil, filterBuilder)
	if err != nil {
		t.Fatalf("AttachmentsFiltered() error = %v", err)
	}

	foundPlanning := false
	for _, att := range filtered.Nodes {
		if att.ID == attachment3.ID {
			foundPlanning = true
			break
		}
	}
	if !foundPlanning {
		t.Errorf("AttachmentsFiltered() did not find 'Planning Document'")
	}
	t.Logf("✓ Filtered attachments by title, found %d matches", len(filtered.Nodes))

	// Cleanup: Delete attachments
	if err := client.AttachmentDelete(ctx, attachment1.ID); err != nil {
		t.Errorf("Failed to delete attachment 1: %v", err)
	} else {
		t.Logf("✓ Deleted attachment 1: %s", attachment1.ID)
	}

	if err := client.AttachmentDelete(ctx, attachment2.ID); err != nil {
		t.Errorf("Failed to delete attachment 2: %v", err)
	} else {
		t.Logf("✓ Deleted attachment 2: %s", attachment2.ID)
	}

	if err := client.AttachmentDelete(ctx, attachment3.ID); err != nil {
		t.Errorf("Failed to delete attachment 3: %v", err)
	} else {
		t.Logf("✓ Deleted attachment 3: %s", attachment3.ID)
	}
}

// TestLive_AttachmentLinkTypes tests different attachment link types.
func TestLive_AttachmentLinkTypes(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Setup: Get team and create issue
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil {
		t.Fatalf("Teams() error = %v", err)
	}
	if len(teams.Nodes) == 0 {
		t.Fatal("No teams found")
	}

	title := "[SDK-TEST] Link Types Test"
	issue, err := client.IssueCreate(ctx, intgraphql.IssueCreateInput{
		TeamID: teams.Nodes[0].ID,
		Title:  &title,
	})
	if err != nil {
		t.Fatalf("IssueCreate() error = %v", err)
	}

	t.Cleanup(func() {
		client.IssueDelete(ctx, issue.ID, nil)
		t.Logf("✓ Cleaned up issue: %s", issue.ID)
	})

	// Test AttachmentLinkURL
	t.Run("AttachmentLinkURL", func(t *testing.T) {
		linkTitle := "Example Link"
		attachment, err := client.AttachmentLinkURL(ctx, issue.ID, "https://example.com", &linkTitle)
		if err != nil {
			t.Fatalf("AttachmentLinkURL() error = %v", err)
		}
		if attachment.ID == "" {
			t.Error("Attachment has empty ID")
		}
		t.Logf("✓ Created URL link: %s", attachment.ID)

		// Cleanup
		t.Cleanup(func() {
			client.AttachmentDelete(ctx, attachment.ID)
		})
	})

	// Test AttachmentLinkGitHubPR (may fail if GitHub integration not configured)
	t.Run("AttachmentLinkGitHubPR", func(t *testing.T) {
		attachment, err := client.AttachmentLinkGitHubPR(ctx, issue.ID, "https://github.com/owner/repo/pull/1")
		if err != nil {
			t.Logf("⚠ AttachmentLinkGitHubPR() error (expected if GitHub not configured): %v", err)
			t.Skip("Skipping GitHub PR link test - integration may not be configured")
		}
		if attachment.ID == "" {
			t.Error("Attachment has empty ID")
		}
		t.Logf("✓ Created GitHub PR link: %s", attachment.ID)

		// Cleanup
		t.Cleanup(func() {
			client.AttachmentDelete(ctx, attachment.ID)
		})
	})

	// Test AttachmentLinkSlack (may fail if Slack integration not configured)
	t.Run("AttachmentLinkSlack", func(t *testing.T) {
		attachment, err := client.AttachmentLinkSlack(ctx, issue.ID, "https://workspace.slack.com/archives/C123/p456")
		if err != nil {
			t.Logf("⚠ AttachmentLinkSlack() error (expected if Slack not configured): %v", err)
			t.Skip("Skipping Slack link test - integration may not be configured")
		}
		if attachment.ID == "" {
			t.Error("Attachment has empty ID")
		}
		t.Logf("✓ Created Slack link: %s", attachment.ID)

		// Cleanup
		t.Cleanup(func() {
			client.AttachmentDelete(ctx, attachment.ID)
		})
	})
}

// strPtr returns a pointer to the string value.
func strPtr(s string) *string {
	return &s
}
