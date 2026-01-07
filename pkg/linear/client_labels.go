package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// IssueLabel retrieves a single issue label by ID.
//
// Returns:
//   - IssueLabel.ID: Label UUID (always populated)
//   - IssueLabel.Name: Label name (always populated)
//   - IssueLabel.Description: Label description (may be empty)
//   - IssueLabel.Color: Label color hex code (always populated)
//   - IssueLabel.CreatedAt: Creation timestamp (always populated)
//   - IssueLabel.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if label not found or query fails
//
// Permissions Required: Read
//
// Related: [IssueLabels], [IssueLabelCreate]
func (c *Client) IssueLabel(ctx context.Context, id string) (*intgraphql.GetLabel_IssueLabel, error) {
	resp, err := c.gqlClient.GetLabel(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("label query", err)
	}
	return &resp.IssueLabel, nil
}

// IssueLabels retrieves a paginated list of issue labels.
//
// Parameters:
//   - first: Number of labels to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - IssueLabels.Nodes: Array of labels (may be empty)
//   - IssueLabels.PageInfo.HasNextPage: true if more results available
//   - IssueLabels.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [IssueLabel], [IssueLabelCreate]
func (c *Client) IssueLabels(ctx context.Context, first *int64, after *string) (*intgraphql.ListLabels_IssueLabels, error) {
	resp, err := c.gqlClient.ListLabels(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("labels query", err)
	}
	return &resp.IssueLabels, nil
}

// IssueLabelsFiltered returns labels matching the specified filter.
func (c *Client) IssueLabelsFiltered(ctx context.Context, first *int64, after *string, filter *intgraphql.IssueLabelFilter) (*intgraphql.ListLabelsFiltered_IssueLabels, error) {
	resp, err := c.gqlClient.ListLabelsFiltered(ctx, first, after, filter)
	if err != nil {
		return nil, wrapGraphQLError("labels filtered query", err)
	}
	return &resp.IssueLabels, nil
}

// IssueLabelCreate creates a new label.
func (c *Client) IssueLabelCreate(ctx context.Context, input intgraphql.IssueLabelCreateInput) (*intgraphql.CreateLabel_IssueLabelCreate_IssueLabel, error) {
	resp, err := c.gqlClient.CreateLabel(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("IssueLabelCreate", err)
	}

	if !resp.IssueLabelCreate.Success {
		return nil, errMutationFailed("IssueLabelCreate")
	}

	return &resp.IssueLabelCreate.IssueLabel, nil
}

// IssueLabelUpdate updates an existing label.
func (c *Client) IssueLabelUpdate(ctx context.Context, id string, input intgraphql.IssueLabelUpdateInput) (*intgraphql.UpdateLabel_IssueLabelUpdate_IssueLabel, error) {
	resp, err := c.gqlClient.UpdateLabel(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("IssueLabelUpdate", err)
	}

	if !resp.IssueLabelUpdate.Success {
		return nil, errMutationFailed("IssueLabelUpdate")
	}

	return &resp.IssueLabelUpdate.IssueLabel, nil
}

// IssueLabelDelete deletes a label by ID.
func (c *Client) IssueLabelDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.DeleteLabel(ctx, id)
	if err != nil {
		return wrapGraphQLError("IssueLabelDelete", err)
	}

	if !resp.IssueLabelDelete.Success {
		return errMutationFailed("IssueLabelDelete")
	}

	return nil
}
