package linear

import (
	"context"
	"fmt"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// Issue retrieves a single issue by ID.
//
// Returns:
//   - Issue.ID: Issue UUID (always populated)
//   - Issue.Title: Issue title (always populated)
//   - Issue.Description: Markdown description (may be empty)
//   - Issue.Priority: 0-4 priority level (always populated)
//   - Issue.Estimate: Story point estimate (may be 0)
//   - Issue.Number: Issue number in team (always populated)
//   - Issue.URL: Linear web URL (always populated)
//   - Issue.State: Workflow state with ID, Name, Type (always populated)
//   - Issue.Team: Team with ID, Name, Key (always populated)
//   - Issue.Assignee: User with ID, Name, DisplayName (nil if unassigned)
//   - Issue.CreatedAt: Creation timestamp (always populated)
//   - Issue.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if issue not found or query fails
//
// Permissions Required: Read
//
// Related: [Issues], [IssueCreate], [IssueUpdate]
func (c *Client) Issue(ctx context.Context, id string) (*intgraphql.GetIssue_Issue, error) {
	resp, err := c.gqlClient.GetIssue(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("issue query", err)
	}
	return &resp.Issue, nil
}

// Issues retrieves a paginated list of issues.
//
// Parameters:
//   - first: Number of issues to return (pointer, nil = server default ~50)
//   - after: Cursor for pagination (pointer, nil = start from beginning)
//
// Returns:
//   - Issues.Nodes: Array of issues (may be empty)
//   - Issues.PageInfo.HasNextPage: true if more results available
//   - Issues.PageInfo.EndCursor: Cursor for next page (pass to after parameter)
//   - error: Non-nil if query fails
//
// Pagination Pattern:
//  1. Call with first=50, after=nil for first page
//  2. Check HasNextPage
//  3. Call again with after=EndCursor for next page
//  4. Repeat until HasNextPage is false
//
// Permissions Required: Read
//
// Related: [Issue], [NewIssueIterator], [IssueSearch]
//
// Example:
//
//	// Get first 10 issues
//	first := int64(10)
//	issues, err := client.Issues(ctx, &first, nil)
//	if err != nil {
//	    return err
//	}
//	for _, issue := range issues.Nodes {
//	    fmt.Println(issue.Title)
//	}
//
// Example pagination:
//
//	cursor := (*string)(nil)
//	for {
//	    issues, err := client.Issues(ctx, &first, cursor)
//	    if err != nil { return err }
//	    // Process issues.Nodes
//	    if !issues.PageInfo.HasNextPage { break }
//	    cursor = issues.PageInfo.EndCursor
//	}
func (c *Client) Issues(ctx context.Context, first *int64, after *string) (*intgraphql.ListIssues_Issues, error) {
	resp, err := c.gqlClient.ListIssues(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("issues query", err)
	}
	return &resp.Issues, nil
}

// IssuesFiltered retrieves a paginated list of issues with filtering.
//
// Parameters:
//   - first: Number of issues to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//   - filter: Issue filter (team, assignee, state, priority, dates, labels)
//
// Returns:
//   - Issues.Nodes: Array of issues matching filter (may be empty)
//   - Issues.PageInfo.HasNextPage: true if more results available
//   - Issues.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Issues], [SearchIssues], [Issue]
func (c *Client) IssuesFiltered(ctx context.Context, first *int64, after *string, filter *intgraphql.IssueFilter) (*intgraphql.ListIssuesFiltered_Issues, error) {
	resp, err := c.gqlClient.ListIssuesFiltered(ctx, first, after, filter)
	if err != nil {
		return nil, wrapGraphQLError("issues filtered query", err)
	}
	return &resp.Issues, nil
}

// IssueCreate creates a new issue in Linear.
//
// Required Input Fields:
//   - TeamID: Team identifier (string, get from Teams() or Team())
//
// Common Optional Fields (all are pointers, nil = omitted):
//   - Title: Issue title (*string)
//   - Description: Markdown description (*string)
//   - Priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low (*int)
//   - AssigneeID: User to assign (*string, get from Users())
//   - StateID: Workflow state (*string, get from WorkflowStates())
//   - LabelIDs: Label identifiers ([]string)
//   - DueDate: Due date (*string, format: YYYY-MM-DD)
//
// Returns:
//   - Issue.ID: Created issue UUID (always populated)
//   - Issue.Number: Issue number in team (always populated)
//   - Issue.Title: Issue title (always populated)
//   - Issue.Team: Team relationship (always populated)
//   - Issue.State: Workflow state (always populated)
//   - error: Non-nil if mutation fails or Success is false
//
// Permissions Required: Write (or issues:create)
//
// Related: [IssueUpdate], [IssueDelete], [Issues]
//
// Example:
//
//	title := "Fix login bug"
//	desc := "Users can't log in on Safari"
//	priority := int64(linear.PriorityUrgent)
//
//	issue, err := client.IssueCreate(ctx, IssueCreateInput{
//	    TeamID: "team-uuid",
//	    Title: &title,
//	    Description: &desc,
//	    Priority: &priority,
//	})
//	if err != nil {
//	    return fmt.Errorf("create failed: %w", err)
//	}
//	log.Printf("Created issue #%d: %s", issue.Number, issue.ID)
func (c *Client) IssueCreate(ctx context.Context, input intgraphql.IssueCreateInput) (*intgraphql.CreateIssue_IssueCreate_Issue, error) {
	resp, err := c.gqlClient.CreateIssue(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("IssueCreate", err)
	}

	if !resp.IssueCreate.Success {
		return nil, errMutationFailed("IssueCreate")
	}

	return resp.IssueCreate.Issue, nil
}

// IssueUpdate updates an existing issue's fields.
//
// Parameters:
//   - id: Issue UUID to update (required)
//   - input: Fields to update (all fields are optional pointers)
//
// Input Fields (all optional, nil = unchanged):
//   - Title: New title (*string)
//   - Description: New description (*string)
//   - Priority: New priority 0-4 (*int)
//   - StateID: New workflow state (*string)
//   - AssigneeID: New assignee (*string, empty string to unassign)
//
// Returns:
//   - Updated issue with modified fields
//   - error: Non-nil if issue not found, permission denied, or mutation fails
//
// Permissions Required: Write
//
// Related: [IssueCreate], [IssueDelete], [Issue]
//
// Example:
//
//	updatedTitle := "Updated: Fix critical bug"
//	priority := int64(linear.PriorityUrgent)
//
//	updated, err := client.IssueUpdate(ctx, issueID, IssueUpdateInput{
//	    Title: &updatedTitle,
//	    Priority: &priority,
//	})
func (c *Client) IssueUpdate(ctx context.Context, id string, input intgraphql.IssueUpdateInput) (*intgraphql.UpdateIssue_IssueUpdate_Issue, error) {
	resp, err := c.gqlClient.UpdateIssue(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("IssueUpdate", err)
	}

	if !resp.IssueUpdate.Success {
		return nil, errMutationFailed("IssueUpdate")
	}

	return resp.IssueUpdate.Issue, nil
}

// IssueBatchUpdate updates multiple issues at once (max 50).
//
// Parameters:
//   - ids: Issue UUIDs to update (max 50)
//   - input: Update to apply to all issues
//
// Returns:
//   - IssueBatchPayload.Issues: Updated issues
//   - IssueBatchPayload.Success: true if successful
//   - error: Non-nil if mutation fails
//
// Permissions Required: Write
//
// Related: [IssueUpdate], [IssuesFiltered]
func (c *Client) IssueBatchUpdate(ctx context.Context, ids []string, input intgraphql.IssueUpdateInput) (*intgraphql.BatchUpdateIssues_IssueBatchUpdate, error) {
	resp, err := c.gqlClient.BatchUpdateIssues(ctx, ids, input)
	if err != nil {
		return nil, wrapGraphQLError("IssueBatchUpdate", err)
	}

	if !resp.IssueBatchUpdate.Success {
		return nil, errMutationFailed("IssueBatchUpdate")
	}

	return &resp.IssueBatchUpdate, nil
}

// IssueDelete permanently deletes an issue.
//
// Parameters:
//   - id: Issue UUID to delete (required)
//
// Returns:
//   - nil: Issue successfully deleted
//   - error: Non-nil if issue not found, permission denied, or deletion fails
//
// Warning: This is permanent and cannot be undone. The issue will be removed
// from all projects, cycles, and relationships.
//
// Permissions Required: Write
//
// Related: [IssueCreate], [IssueUpdate]
//
// Example:
//
//	if err := client.IssueDelete(ctx, issueID); err != nil {
//	    return fmt.Errorf("delete failed: %w", err)
//	}
func (c *Client) IssueDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.DeleteIssue(ctx, id)
	if err != nil {
		return wrapGraphQLError("IssueDelete", err)
	}

	if !resp.IssueDelete.Success {
		return errMutationFailed("IssueDelete")
	}

	return nil
}

// SearchIssues searches for issues matching the search term.
//
// The new searchIssues API replaces the deprecated issueSearch endpoint.
// It uses full-text search with optional structured filtering.
//
// Parameters:
//   - term: Search text (required). Searches across issue titles, descriptions, and optionally comments.
//   - first: Number of results per page (nil = default ~50, max: 250)
//   - after: Pagination cursor from previous PageInfo.EndCursor (nil = first page)
//   - filter: Optional structured filters (assignee, state, priority, team, etc.)
//   - includeArchived: Include archived issues in results (default: false)
//
// Returns:
//   - SearchIssues with nodes (matching issues) and pageInfo (pagination)
//   - totalCount: Total number of matches (useful for showing "X of Y results")
//   - Error on failure (network, auth, invalid filter)
//
// Example (simple text search):
//
//	term := "bug"
//	first := int64(50)
//	issues, err := client.SearchIssues(ctx, term, &first, nil, nil, nil)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Found %d results\n", len(issues.Nodes))
//	for _, issue := range issues.Nodes {
//	    fmt.Printf("[%.0f] %s\n", issue.Number, issue.Title)
//	}
//
// Example (with filter):
//
//	term := "login"
//	first := int64(50)
//	priorityHigh := float64(2)
//	filter := &intgraphql.IssueFilter{
//	    Priority: &intgraphql.NullableNumberComparator{Eq: &priorityHigh},
//	}
//	issues, err := client.SearchIssues(ctx, term, &first, nil, filter, nil)
//
// Related: [Issues], [IssueCreate], [IssueUpdate]
func (c *Client) SearchIssues(ctx context.Context, term string, first *int64, after *string, filter *intgraphql.IssueFilter, includeArchived *bool) (*intgraphql.SearchIssues_SearchIssues, error) {
	resp, err := c.gqlClient.SearchIssues(ctx, term, first, after, filter, includeArchived)
	if err != nil {
		return nil, fmt.Errorf("issue search failed: %w", err)
	}
	return &resp.SearchIssues, nil
}

// IssueAddLabel adds a label to an issue.
//
// Simpler alternative to IssueUpdate with AddedLabelIds array.
// Does not require fetching existing labels first.
//
// Parameters:
//   - id: Issue UUID (required)
//   - labelID: Label UUID to add (required)
//
// Returns:
//   - Updated issue with labels collection
//   - error: Non-nil if operation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	issue, err := client.IssueAddLabel(ctx, issueID, labelID)
//	if err != nil {
//	    return fmt.Errorf("failed to add label: %w", err)
//	}
//	fmt.Printf("Issue now has %d labels\n", len(issue.Labels.Nodes))
//
// Related: [IssueRemoveLabel], [IssueLabelCreate], [IssueUpdate]
func (c *Client) IssueAddLabel(ctx context.Context, id, labelID string) (*intgraphql.IssueAddLabel_IssueAddLabel_Issue, error) {
	resp, err := c.gqlClient.IssueAddLabel(ctx, id, labelID)
	if err != nil {
		return nil, wrapGraphQLError("IssueAddLabel", err)
	}
	if !resp.IssueAddLabel.Success {
		return nil, errMutationFailed("IssueAddLabel")
	}
	return resp.IssueAddLabel.Issue, nil
}

// IssueRemoveLabel removes a label from an issue.
//
// Simpler alternative to IssueUpdate with array manipulation.
// Does not require fetching and filtering existing labels.
//
// Parameters:
//   - id: Issue UUID (required)
//   - labelID: Label UUID to remove (required)
//
// Returns:
//   - Updated issue with labels collection
//   - error: Non-nil if operation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	issue, err := client.IssueRemoveLabel(ctx, issueID, labelID)
//	if err != nil {
//	    return fmt.Errorf("failed to remove label: %w", err)
//	}
//	fmt.Printf("Issue now has %d labels\n", len(issue.Labels.Nodes))
//
// Related: [IssueAddLabel], [IssueLabelDelete], [IssueUpdate]
func (c *Client) IssueRemoveLabel(ctx context.Context, id, labelID string) (*intgraphql.IssueRemoveLabel_IssueRemoveLabel_Issue, error) {
	resp, err := c.gqlClient.IssueRemoveLabel(ctx, id, labelID)
	if err != nil {
		return nil, wrapGraphQLError("IssueRemoveLabel", err)
	}
	if !resp.IssueRemoveLabel.Success {
		return nil, errMutationFailed("IssueRemoveLabel")
	}
	return resp.IssueRemoveLabel.Issue, nil
}

// IssueRelationCreate creates a relationship between two issues.
//
// Relationship types:
//   - "blocks": This issue blocks another issue
//   - "blocked": This issue is blocked by another
//   - "duplicate": This issue is a duplicate of another
//   - "related": This issue is related to another
//
// Parameters:
//   - input: Relation parameters (issueId, relatedIssueId, type)
//
// Returns:
//   - Created relation with both issues and type
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	relation, err := client.IssueRelationCreate(ctx, intgraphql.IssueRelationCreateInput{
//	    IssueID:        &currentIssueID,
//	    RelatedIssueID: &blockerIssueID,
//	    Type:           "blocks",
//	})
//	fmt.Printf("Issue %s blocks %s\n", relation.Issue.Title, relation.RelatedIssue.Title)
//
// Related: [IssueRelationUpdate], [IssueRelationDelete]
func (c *Client) IssueRelationCreate(ctx context.Context, input intgraphql.IssueRelationCreateInput) (*intgraphql.IssueRelationCreate_IssueRelationCreate_IssueRelation, error) {
	resp, err := c.gqlClient.IssueRelationCreate(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("IssueRelationCreate", err)
	}
	if !resp.IssueRelationCreate.Success {
		return nil, errMutationFailed("IssueRelationCreate")
	}
	return &resp.IssueRelationCreate.IssueRelation, nil
}

// IssueRelationUpdate updates an existing relationship between issues.
//
// Use this to change the relationship type (e.g., from "related" to "blocks").
//
// Parameters:
//   - id: IssueRelation UUID to update (required)
//   - input: Fields to update (type is the main updatable field)
//
// Returns:
//   - Updated relation with new type
//   - error: Non-nil if update fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	updatedType := "blocks"
//	relation, err := client.IssueRelationUpdate(ctx, relationID, intgraphql.IssueRelationUpdateInput{
//	    Type: &updatedType,
//	})
//
// Related: [IssueRelationCreate], [IssueRelationDelete]
func (c *Client) IssueRelationUpdate(ctx context.Context, id string, input intgraphql.IssueRelationUpdateInput) (*intgraphql.IssueRelationUpdate_IssueRelationUpdate_IssueRelation, error) {
	resp, err := c.gqlClient.IssueRelationUpdate(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("IssueRelationUpdate", err)
	}
	if !resp.IssueRelationUpdate.Success {
		return nil, errMutationFailed("IssueRelationUpdate")
	}
	return &resp.IssueRelationUpdate.IssueRelation, nil
}

// IssueRelationDelete deletes a relationship between issues.
//
// Parameters:
//   - id: IssueRelation UUID to delete (required)
//
// Returns:
//   - nil: Relation successfully deleted
//   - error: Non-nil if delete fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	err := client.IssueRelationDelete(ctx, relationID)
//	if err != nil {
//	    return fmt.Errorf("failed to delete relation: %w", err)
//	}
//
// Related: [IssueRelationCreate], [IssueRelationUpdate]
func (c *Client) IssueRelationDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.IssueRelationDelete(ctx, id)
	if err != nil {
		return wrapGraphQLError("IssueRelationDelete", err)
	}
	if !resp.IssueRelationDelete.Success {
		return errMutationFailed("IssueRelationDelete")
	}
	return nil
}

// IssueSuggestions retrieves AI suggestions for an issue.
//
// Parameters:
//   - issueID: Issue UUID (required)
//   - first: Number of suggestions to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Issue with nested suggestions list
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Issue]
func (c *Client) IssueSuggestions(ctx context.Context, issueID string, first *int64, after *string) (*intgraphql.GetIssueSuggestionsForIssue_Issue, error) {
	resp, err := c.gqlClient.GetIssueSuggestionsForIssue(ctx, issueID, first, after)
	if err != nil {
		return nil, wrapGraphQLError("issue suggestions query", err)
	}
	return &resp.Issue, nil
}
