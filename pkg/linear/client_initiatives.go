package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// Initiative retrieves a single strategic initiative by ID.
//
// Returns:
//   - Initiative.ID: Initiative UUID (always populated)
//   - Initiative.Name: Initiative name (always populated)
//   - Initiative.Description: Initiative description (may be empty)
//   - Initiative.TargetDate: Target completion date (may be empty)
//   - Initiative.SortOrder: Sort position (always populated)
//   - Initiative.Icon: Initiative icon (may be empty)
//   - Initiative.Color: Initiative color hex code (may be empty)
//   - Initiative.CreatedAt: Creation timestamp (always populated)
//   - Initiative.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if initiative not found or query fails
//
// Permissions Required: Read
//
// Related: [Initiatives]
func (c *Client) Initiative(ctx context.Context, id string) (*intgraphql.GetInitiative_Initiative, error) {
	resp, err := c.gqlClient.GetInitiative(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("initiative query", err)
	}
	return &resp.Initiative, nil
}

// Initiatives retrieves a paginated list of strategic initiatives.
//
// Parameters:
//   - first: Number of initiatives to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Initiatives.Nodes: Array of initiatives (may be empty)
//   - Initiatives.PageInfo.HasNextPage: true if more results available
//   - Initiatives.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Initiative]
func (c *Client) Initiatives(ctx context.Context, first *int64, after *string) (*intgraphql.ListInitiatives_Initiatives, error) {
	resp, err := c.gqlClient.ListInitiatives(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("initiatives query", err)
	}
	return &resp.Initiatives, nil
}

// InitiativesFiltered retrieves initiatives with optional filtering.
//
// Filters:
//   - CreatedAt, TargetDate: Date range filters
//   - Creator, Owner: Filter by users
//   - ID, Name, SlugID: Identity and text filters
//   - Health, Status: State filters (onTrack/atRisk/offTrack, Planned/Active/Completed)
//
// Permissions Required: Read
//
// Related: [Initiatives], [Initiative]
func (c *Client) InitiativesFiltered(ctx context.Context, first *int64, after *string, filter *intgraphql.InitiativeFilter) (*intgraphql.ListInitiativesFiltered_Initiatives, error) {
	resp, err := c.gqlClient.ListInitiativesFiltered(ctx, first, after, filter)
	if err != nil {
		return nil, wrapGraphQLError("initiatives filtered query", err)
	}
	return &resp.Initiatives, nil
}

// InitiativeCreate creates a new initiative.
func (c *Client) InitiativeCreate(ctx context.Context, input intgraphql.InitiativeCreateInput) (*intgraphql.CreateInitiative_InitiativeCreate_Initiative, error) {
	resp, err := c.gqlClient.CreateInitiative(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("InitiativeCreate", err)
	}
	if !resp.InitiativeCreate.Success {
		return nil, errMutationFailed("InitiativeCreate")
	}
	return &resp.InitiativeCreate.Initiative, nil
}

// InitiativeUpdate updates an existing initiative.
func (c *Client) InitiativeUpdate(ctx context.Context, id string, input intgraphql.InitiativeUpdateInput) (*intgraphql.UpdateInitiative_InitiativeUpdate_Initiative, error) {
	resp, err := c.gqlClient.UpdateInitiative(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("InitiativeUpdate", err)
	}
	if !resp.InitiativeUpdate.Success {
		return nil, errMutationFailed("InitiativeUpdate")
	}
	return &resp.InitiativeUpdate.Initiative, nil
}
