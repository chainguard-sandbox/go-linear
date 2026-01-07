package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// Cycle retrieves a single development cycle (sprint) by ID.
//
// Returns:
//   - Cycle.ID: Cycle UUID (always populated)
//   - Cycle.Number: Cycle number in sequence (always populated)
//   - Cycle.Name: Cycle name (may be empty)
//   - Cycle.Description: Cycle description (may be empty)
//   - Cycle.StartsAt: Start date timestamp (always populated)
//   - Cycle.EndsAt: End date timestamp (always populated)
//   - Cycle.CompletedAt: Completion timestamp (nil if not completed)
//   - Cycle.Progress: Completion progress 0.0-1.0 (always populated)
//   - Cycle.Team: Team with ID, Name, Key (always populated)
//   - Cycle.CreatedAt: Creation timestamp (always populated)
//   - Cycle.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if cycle not found or query fails
//
// Permissions Required: Read
//
// Related: [Cycles]
func (c *Client) Cycle(ctx context.Context, id string) (*intgraphql.GetCycle_Cycle, error) {
	resp, err := c.gqlClient.GetCycle(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("cycle query", err)
	}
	return &resp.Cycle, nil
}

// Cycles retrieves a paginated list of development cycles (sprints).
//
// Parameters:
//   - first: Number of cycles to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Cycles.Nodes: Array of cycles (may be empty)
//   - Cycles.PageInfo.HasNextPage: true if more results available
//   - Cycles.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Cycle]
func (c *Client) Cycles(ctx context.Context, first *int64, after *string) (*intgraphql.ListCycles_Cycles, error) {
	resp, err := c.gqlClient.ListCycles(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("cycles query", err)
	}
	return &resp.Cycles, nil
}

// CyclesFiltered retrieves cycles matching the specified filter criteria.
//
// Parameters:
//   - first: Maximum number of cycles to return (nil = default 50)
//   - after: Cursor for pagination (nil = first page)
//   - filter: Filter criteria (nil = no filtering)
//
// Filter Options:
//   - CreatedAt, UpdatedAt, CompletedAt: Date range comparators
//   - StartsAt, EndsAt: Cycle date range comparators
//   - IsActive, IsFuture, IsPast: Boolean state filters
//   - IsNext, IsPrevious, IsInCooldown: Relative cycle filters
//   - Name: String comparator for cycle name
//   - Number: Number comparator for cycle number
//   - Team: Filter by team
//
// Returns:
//   - Cycles with ID, number, name, description, dates, progress, and team
//   - PageInfo for pagination (HasNextPage, EndCursor)
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Cycles], [Cycle]
func (c *Client) CyclesFiltered(ctx context.Context, first *int64, after *string, filter *intgraphql.CycleFilter) (*intgraphql.ListCyclesFiltered_Cycles, error) {
	resp, err := c.gqlClient.ListCyclesFiltered(ctx, first, after, filter)
	if err != nil {
		return nil, wrapGraphQLError("cycles filtered query", err)
	}
	return &resp.Cycles, nil
}

// CycleCreate creates a new development cycle (sprint).
//
// Parameters:
//   - input: Cycle creation parameters (name, description, start/end dates, team)
//
// Required fields:
//   - TeamID: UUID of the team (required)
//   - Name: Cycle name (required)
//   - StartsAt: Start date in ISO 8601 format (required)
//   - EndsAt: End date in ISO 8601 format (required)
//
// Optional fields:
//   - Description: Cycle description (nil = empty)
//
// Returns:
//   - Created cycle with ID, number, name, dates, and team
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	startsAt := "2025-01-06"
//	endsAt := "2025-01-17"
//	name := "Sprint 1"
//	cycle, err := client.CycleCreate(ctx, intgraphql.CycleCreateInput{
//	    TeamID:   teamID,
//	    Name:     &name,
//	    StartsAt: &startsAt,
//	    EndsAt:   &endsAt,
//	})
//
// Related: [CycleUpdate], [CycleArchive], [Cycles]
func (c *Client) CycleCreate(ctx context.Context, input intgraphql.CycleCreateInput) (*intgraphql.CreateCycle_CycleCreate_Cycle, error) {
	resp, err := c.gqlClient.CreateCycle(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("CycleCreate", err)
	}
	if !resp.CycleCreate.Success {
		return nil, errMutationFailed("CycleCreate")
	}
	return resp.CycleCreate.Cycle, nil
}

// CycleUpdate updates an existing development cycle.
//
// Parameters:
//   - id: Cycle UUID to update (required)
//   - input: Fields to update (all optional, nil = unchanged)
//
// Optional fields:
//   - Name: Update cycle name
//   - Description: Update description
//   - StartsAt: Update start date (ISO 8601 format)
//   - EndsAt: Update end date (ISO 8601 format)
//
// Returns:
//   - Updated cycle with new values
//   - error: Non-nil if update fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	updatedName := "Sprint 1 (Extended)"
//	cycle, err := client.CycleUpdate(ctx, cycleID, intgraphql.CycleUpdateInput{
//	    Name: &updatedName,
//	})
//
// Related: [CycleCreate], [CycleArchive], [Cycle]
func (c *Client) CycleUpdate(ctx context.Context, id string, input intgraphql.CycleUpdateInput) (*intgraphql.UpdateCycle_CycleUpdate_Cycle, error) {
	resp, err := c.gqlClient.UpdateCycle(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("CycleUpdate", err)
	}
	if !resp.CycleUpdate.Success {
		return nil, errMutationFailed("CycleUpdate")
	}
	return resp.CycleUpdate.Cycle, nil
}

// CycleArchive archives (soft-deletes) a development cycle.
//
// Archiving a cycle:
//   - Removes it from active cycle lists
//   - Issues remain associated with the cycle
//   - Cycle data is preserved for historical tracking
//   - Can be unarchived via Linear UI
//
// Parameters:
//   - id: Cycle UUID to archive (required)
//
// Returns:
//   - error: Non-nil if archive fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	err := client.CycleArchive(ctx, cycleID)
//	if err != nil {
//	    return fmt.Errorf("failed to archive cycle: %w", err)
//	}
//
// Related: [CycleCreate], [CycleUpdate], [Cycle]
func (c *Client) CycleArchive(ctx context.Context, id string) error {
	resp, err := c.gqlClient.ArchiveCycle(ctx, id)
	if err != nil {
		return wrapGraphQLError("CycleArchive", err)
	}
	if !resp.CycleArchive.Success {
		return errMutationFailed("CycleArchive")
	}
	return nil
}
