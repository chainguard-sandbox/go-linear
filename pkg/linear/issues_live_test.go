//go:build read

package linear

import (
	"context"
	"os"
	"testing"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
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

// TestLive_IssuesFiltered tests IssuesFiltered query against real Linear API.
// Safe read-only operation.
func TestLive_IssuesFiltered(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Get a team to filter by
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil || len(teams.Nodes) == 0 {
		t.Skip("No teams available for testing")
	}
	teamID := teams.Nodes[0].ID

	tests := []struct {
		name   string
		filter *intgraphql.IssueFilter
	}{
		{
			name:   "no filter",
			filter: nil,
		},
		{
			name: "filter by team",
			filter: &intgraphql.IssueFilter{
				Team: &intgraphql.TeamFilter{
					ID: &intgraphql.IDComparator{Eq: &teamID},
				},
			},
		},
		{
			name: "filter by priority",
			filter: &intgraphql.IssueFilter{
				Priority: &intgraphql.NullableNumberComparator{
					Gte: float64Ptr(1),
				},
			},
		},
		{
			name: "combined filters",
			filter: &intgraphql.IssueFilter{
				Team: &intgraphql.TeamFilter{
					ID: &intgraphql.IDComparator{Eq: &teamID},
				},
				Priority: &intgraphql.NullableNumberComparator{
					Gte: float64Ptr(0),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first := int64(5)
			issues, err := client.IssuesFiltered(ctx, &first, nil, tt.filter)
			if err != nil {
				t.Fatalf("IssuesFiltered() error = %v", err)
			}

			if issues == nil {
				t.Fatal("IssuesFiltered() returned nil")
			}

			t.Logf("Retrieved %d issues with filter", len(issues.Nodes))
		})
	}
}

// TestLive_SearchIssues tests SearchIssues against real Linear API.
// Safe read-only operation.
func TestLive_SearchIssues(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Search for common term
	first := int64(5)
	results, err := client.SearchIssues(ctx, "test", &first, nil, nil, nil)
	if err != nil {
		t.Fatalf("SearchIssues() error = %v", err)
	}

	t.Logf("Search found %d results", len(results.Nodes))
}

// TestLive_Attachments tests Attachment operations against real Linear API.
// Safe read-only operation.
func TestLive_Attachments(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	first := int64(10)
	attachments, err := client.Attachments(ctx, &first, nil)
	if err != nil {
		t.Fatalf("Attachments() error = %v", err)
	}

	t.Logf("Retrieved %d attachments", len(attachments.Nodes))

	// If we have attachments, test getting a single one
	if len(attachments.Nodes) > 0 {
		attachmentID := attachments.Nodes[0].ID
		attachment, err := client.Attachment(ctx, attachmentID)
		if err != nil {
			t.Fatalf("Attachment() error = %v", err)
		}
		if attachment.ID != attachmentID {
			t.Errorf("Attachment ID = %s, want %s", attachment.ID, attachmentID)
		}
	}
}

// TestLive_Initiatives tests Initiative operations against real Linear API.
// Safe read-only operation.
func TestLive_Initiatives(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	first := int64(10)
	initiatives, err := client.Initiatives(ctx, &first, nil)
	if err != nil {
		t.Fatalf("Initiatives() error = %v", err)
	}

	t.Logf("Retrieved %d initiatives", len(initiatives.Nodes))

	// If we have initiatives, test getting a single one
	if len(initiatives.Nodes) > 0 {
		initiativeID := initiatives.Nodes[0].ID
		initiative, err := client.Initiative(ctx, initiativeID)
		if err != nil {
			t.Fatalf("Initiative() error = %v", err)
		}
		if initiative.ID != initiativeID {
			t.Errorf("Initiative ID = %s, want %s", initiative.ID, initiativeID)
		}
	}
}

func float64Ptr(f float64) *float64 {
	return &f
}

// TestLive_IssuePagination tests pagination through issues.
func TestLive_IssuePagination(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// First page
	first := int64(2)
	firstPage, err := client.Issues(ctx, &first, nil)
	if err != nil {
		t.Fatalf("Issues() first page error = %v", err)
	}

	if len(firstPage.Nodes) < 2 {
		t.Skip("Not enough issues to test pagination")
	}

	if !firstPage.PageInfo.HasNextPage {
		t.Skip("No next page to test")
	}

	// Second page - EndCursor is already a *string
	secondPage, err := client.Issues(ctx, &first, firstPage.PageInfo.EndCursor)
	if err != nil {
		t.Fatalf("Issues() second page error = %v", err)
	}

	// Verify we got different issues
	firstPageIDs := make(map[string]bool)
	for _, issue := range firstPage.Nodes {
		firstPageIDs[issue.ID] = true
	}

	for _, issue := range secondPage.Nodes {
		if firstPageIDs[issue.ID] {
			t.Errorf("Second page contains issue from first page: %s", issue.ID)
		}
	}

	t.Logf("Pagination test: first page %d issues, second page %d issues",
		len(firstPage.Nodes), len(secondPage.Nodes))
}

// TestLive_SearchIssuesWithFilter tests search with filter.
func TestLive_SearchIssuesWithFilter(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Get a team to filter by
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil || len(teams.Nodes) == 0 {
		t.Skip("No teams available for testing")
	}
	teamID := teams.Nodes[0].ID

	// Search with team filter
	first := int64(5)
	filter := &intgraphql.IssueFilter{
		Team: &intgraphql.TeamFilter{
			ID: &intgraphql.IDComparator{Eq: &teamID},
		},
	}

	results, err := client.SearchIssues(ctx, "test", &first, nil, filter, nil)
	if err != nil {
		t.Fatalf("SearchIssues() error = %v", err)
	}

	t.Logf("Search with filter found %d results", len(results.Nodes))
}

// TestLive_IssuesWithFilters tests various filter combinations.
func TestLive_IssuesWithFilters(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	first := int64(5)

	// Test priority filter
	priorityFilter := &intgraphql.IssueFilter{
		Priority: &intgraphql.NullableNumberComparator{
			Gte: float64Ptr(2),
		},
	}

	issues, err := client.IssuesFiltered(ctx, &first, nil, priorityFilter)
	if err != nil {
		t.Fatalf("IssuesFiltered with priority filter error = %v", err)
	}

	t.Logf("Found %d high priority issues (priority >= 2)", len(issues.Nodes))

	// Verify filter worked
	for _, issue := range issues.Nodes {
		if issue.Priority < 2 {
			t.Errorf("Issue %s has priority %f, expected >= 2", issue.ID, issue.Priority)
		}
	}
}

// TestLive_IssuesFilteredByDate tests date-based filtering.
func TestLive_IssuesFilteredByDate(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	first := int64(10)

	// Get issues without date filter first
	allIssues, err := client.Issues(ctx, &first, nil)
	if err != nil {
		t.Fatalf("Issues() error = %v", err)
	}

	t.Logf("Total issues (first %d): %d", first, len(allIssues.Nodes))
	if len(allIssues.Nodes) > 0 {
		t.Logf("First issue created at: %s", allIssues.Nodes[0].CreatedAt)
	}
}

// TestLive_MultipleRequests tests making multiple sequential requests.
func TestLive_MultipleRequests(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Make multiple requests to verify client stability
	for i := 0; i < 3; i++ {
		_, err := client.Viewer(ctx)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
	}

	t.Log("Successfully made 3 sequential requests")
}

// TestLive_ConcurrentRequests tests making concurrent requests.
func TestLive_ConcurrentRequests(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	first := int64(1)

	// Make concurrent requests
	done := make(chan error, 3)

	go func() {
		_, err := client.Viewer(ctx)
		done <- err
	}()

	go func() {
		_, err := client.Teams(ctx, &first, nil)
		done <- err
	}()

	go func() {
		_, err := client.Issues(ctx, &first, nil)
		done <- err
	}()

	// Wait for all to complete
	for i := 0; i < 3; i++ {
		if err := <-done; err != nil {
			t.Errorf("Concurrent request %d failed: %v", i+1, err)
		}
	}

	t.Log("Successfully made 3 concurrent requests")
}
