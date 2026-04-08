package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// CustomView retrieves a single custom view by ID.
//
// Returns:
//   - CustomView with ID, Name, Description, Icon, Color, FilterData, ModelName, Shared, Creator, Owner, Team
//   - error: Non-nil if custom view not found or query fails
//
// Permissions Required: Read
//
// Related: [CustomViews], [CustomViewCreate], [CustomViewUpdate]
func (c *Client) CustomView(ctx context.Context, id string) (*intgraphql.GetCustomView_CustomView, error) {
	resp, err := c.gqlClient.GetCustomView(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("custom view query", err)
	}
	return &resp.CustomView, nil
}

// CustomViews retrieves a paginated list of custom views.
//
// Parameters:
//   - first: Number of custom views to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - CustomViews.Nodes: Array of custom views (may be empty)
//   - CustomViews.PageInfo.HasNextPage: true if more results available
//   - CustomViews.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [CustomView], [CustomViewCreate]
func (c *Client) CustomViews(ctx context.Context, first *int64, after *string) (*intgraphql.ListCustomViews_CustomViews, error) {
	resp, err := c.gqlClient.ListCustomViews(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("custom views query", err)
	}
	return &resp.CustomViews, nil
}

// CustomViewCreate creates a new custom view.
//
// Required fields:
//   - Name: Custom view name
//
// Optional fields:
//   - Description: Custom view description
//   - Icon: Custom view icon
//   - Color: Icon color
//   - FilterData: Issue filter as IssueFilter
//   - Shared: Whether shared with organization
//   - TeamID: Associated team UUID
//
// Returns:
//   - Created custom view with ID, Name, and other fields
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Related: [CustomView], [CustomViewUpdate], [CustomViewDelete]
func (c *Client) CustomViewCreate(ctx context.Context, input intgraphql.CustomViewCreateInput) (*intgraphql.CreateCustomView_CustomViewCreate_CustomView, error) {
	resp, err := c.gqlClient.CreateCustomView(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("CustomViewCreate", err)
	}

	if !resp.CustomViewCreate.Success {
		return nil, errMutationFailed("CustomViewCreate")
	}

	return &resp.CustomViewCreate.CustomView, nil
}

// CustomViewUpdate updates an existing custom view.
//
// Parameters:
//   - id: Custom view UUID to update (required)
//   - input: Fields to update (all optional, nil = unchanged)
//
// Returns:
//   - Updated custom view with new values
//   - error: Non-nil if update fails or Success is false
//
// Permissions Required: Write
//
// Related: [CustomView], [CustomViewCreate], [CustomViewDelete]
func (c *Client) CustomViewUpdate(ctx context.Context, id string, input intgraphql.CustomViewUpdateInput) (*intgraphql.UpdateCustomView_CustomViewUpdate_CustomView, error) {
	resp, err := c.gqlClient.UpdateCustomView(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("CustomViewUpdate", err)
	}

	if !resp.CustomViewUpdate.Success {
		return nil, errMutationFailed("CustomViewUpdate")
	}

	return &resp.CustomViewUpdate.CustomView, nil
}

// CustomViewDelete deletes a custom view by ID.
//
// Parameters:
//   - id: Custom view UUID to delete (required)
//
// Returns:
//   - nil: Custom view successfully deleted
//   - error: Non-nil if delete fails or Success is false
//
// Permissions Required: Write
//
// Related: [CustomView], [CustomViewCreate]
func (c *Client) CustomViewDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.DeleteCustomView(ctx, id)
	if err != nil {
		return wrapGraphQLError("CustomViewDelete", err)
	}

	if !resp.CustomViewDelete.Success {
		return errMutationFailed("CustomViewDelete")
	}

	return nil
}
