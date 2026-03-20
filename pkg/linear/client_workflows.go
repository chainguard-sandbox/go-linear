package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// WorkflowState retrieves a single workflow state by ID.
//
// Returns:
//   - WorkflowState.ID: State UUID (always populated)
//   - WorkflowState.Name: State name (always populated)
//   - WorkflowState.Description: State description (may be empty)
//   - WorkflowState.Type: State type (unstarted/started/completed/canceled)
//   - WorkflowState.Color: State color hex code (always populated)
//   - WorkflowState.Position: Sort position (always populated)
//   - WorkflowState.CreatedAt: Creation timestamp (always populated)
//   - WorkflowState.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if state not found or query fails
//
// Permissions Required: Read
//
// Related: [WorkflowStates]
func (c *Client) WorkflowState(ctx context.Context, id string) (*intgraphql.GetWorkflowState_WorkflowState, error) {
	resp, err := c.gqlClient.GetWorkflowState(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("workflow state query", err)
	}
	return &resp.WorkflowState, nil
}

// WorkflowStates retrieves a paginated list of workflow states.
//
// Parameters:
//   - first: Number of states to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - WorkflowStates.Nodes: Array of states (may be empty)
//   - WorkflowStates.PageInfo.HasNextPage: true if more results available
//   - WorkflowStates.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [WorkflowState]
func (c *Client) WorkflowStates(ctx context.Context, first *int64, after *string) (*intgraphql.ListWorkflowStates_WorkflowStates, error) {
	resp, err := c.gqlClient.ListWorkflowStates(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("workflow states query", err)
	}
	return &resp.WorkflowStates, nil
}

// WorkflowStatesFiltered returns workflow states matching the specified filter.
func (c *Client) WorkflowStatesFiltered(ctx context.Context, first *int64, after *string, filter *intgraphql.WorkflowStateFilter) (*intgraphql.ListWorkflowStatesFiltered_WorkflowStates, error) {
	resp, err := c.gqlClient.ListWorkflowStatesFiltered(ctx, first, after, filter)
	if err != nil {
		return nil, wrapGraphQLError("workflow states filtered query", err)
	}
	return &resp.WorkflowStates, nil
}
