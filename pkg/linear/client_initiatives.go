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

// InitiativeUpdateCreate creates a status update for an initiative.
//
// Parameters:
//   - input: Status update creation parameters
//
// Required fields:
//   - InitiativeID: UUID of the initiative (required)
//   - Body: Update body in markdown format (required)
//
// Optional fields:
//   - Health: Initiative health status (onTrack, atRisk, offTrack)
//
// Returns:
//   - Created initiative update with ID, body, health, and dates
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Related: [InitiativeUpdateArchive], [GetInitiativeUpdate], [ListInitiativeUpdates]
func (c *Client) InitiativeUpdateCreate(ctx context.Context, input intgraphql.InitiativeUpdateCreateInput) (*intgraphql.CreateInitiativeUpdate_InitiativeUpdateCreate_InitiativeUpdate, error) {
	resp, err := c.gqlClient.CreateInitiativeUpdate(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("InitiativeUpdateCreate", err)
	}
	if !resp.InitiativeUpdateCreate.Success {
		return nil, errMutationFailed("InitiativeUpdateCreate")
	}
	return &resp.InitiativeUpdateCreate.InitiativeUpdate, nil
}

// GetInitiativeUpdate retrieves a single initiative status update by ID.
//
// Returns:
//   - InitiativeUpdate with ID, body, health, user, dates, and URL
//   - error: Non-nil if update not found or query fails
//
// Permissions Required: Read
//
// Related: [ListInitiativeUpdates], [InitiativeUpdateCreate]
func (c *Client) GetInitiativeUpdate(ctx context.Context, id string) (*intgraphql.GetInitiativeUpdate_InitiativeUpdate, error) {
	resp, err := c.gqlClient.GetInitiativeUpdate(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("initiative update query", err)
	}
	return &resp.InitiativeUpdate, nil
}

// ListInitiativeUpdates retrieves status updates for an initiative.
//
// Parameters:
//   - initiativeID: Initiative UUID (required)
//   - first: Number of updates to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Initiative with nested initiative updates list
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [GetInitiativeUpdate], [InitiativeUpdateCreate]
func (c *Client) ListInitiativeUpdates(ctx context.Context, initiativeID string, first *int64, after *string) (*intgraphql.ListInitiativeUpdates_Initiative, error) {
	resp, err := c.gqlClient.ListInitiativeUpdates(ctx, initiativeID, first, after)
	if err != nil {
		return nil, wrapGraphQLError("list initiative updates query", err)
	}
	return &resp.Initiative, nil
}

// InitiativeUpdateArchive archives an initiative status update by ID.
//
// Parameters:
//   - id: Initiative update UUID to archive (required)
//
// Returns:
//   - nil: Update successfully archived
//   - error: Non-nil if archive fails or Success is false
//
// Permissions Required: Write
//
// Related: [InitiativeUpdateCreate], [GetInitiativeUpdate]
func (c *Client) InitiativeUpdateArchive(ctx context.Context, id string) error {
	resp, err := c.gqlClient.ArchiveInitiativeUpdate(ctx, id)
	if err != nil {
		return wrapGraphQLError("InitiativeUpdateArchive", err)
	}
	if !resp.InitiativeUpdateArchive.Success {
		return errMutationFailed("InitiativeUpdateArchive")
	}
	return nil
}

// InitiativeDelete deletes an initiative by ID.
//
// Parameters:
//   - id: Initiative UUID to delete (required)
//
// Returns:
//   - nil: Initiative successfully deleted
//   - error: Non-nil if delete fails or Success is false
//
// Permissions Required: Write
//
// Related: [InitiativeCreate], [InitiativeUpdate]
func (c *Client) InitiativeDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.DeleteInitiative(ctx, id)
	if err != nil {
		return wrapGraphQLError("InitiativeDelete", err)
	}
	if !resp.InitiativeDelete.Success {
		return errMutationFailed("InitiativeDelete")
	}
	return nil
}

// InitiativeToProjectCreate links a project to an initiative.
//
// Parameters:
//   - input: Initiative-project link creation parameters
//
// Required fields:
//   - InitiativeID: UUID of the initiative (required)
//   - ProjectID: UUID of the project (required)
//
// Returns:
//   - Created link with ID, initiative, and project details
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Related: [InitiativeToProjectDelete]
func (c *Client) InitiativeToProjectCreate(ctx context.Context, input intgraphql.InitiativeToProjectCreateInput) (*intgraphql.CreateInitiativeToProject_InitiativeToProjectCreate_InitiativeToProject, error) {
	resp, err := c.gqlClient.CreateInitiativeToProject(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("InitiativeToProjectCreate", err)
	}
	if !resp.InitiativeToProjectCreate.Success {
		return nil, errMutationFailed("InitiativeToProjectCreate")
	}
	return &resp.InitiativeToProjectCreate.InitiativeToProject, nil
}

// InitiativeToProjectDelete unlinks a project from an initiative.
//
// Parameters:
//   - id: InitiativeToProject link UUID to delete (required)
//
// Returns:
//   - nil: Link successfully deleted
//   - error: Non-nil if delete fails or Success is false
//
// Permissions Required: Write
//
// Related: [InitiativeToProjectCreate]
func (c *Client) InitiativeToProjectDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.DeleteInitiativeToProject(ctx, id)
	if err != nil {
		return wrapGraphQLError("InitiativeToProjectDelete", err)
	}
	if !resp.InitiativeToProjectDelete.Success {
		return errMutationFailed("InitiativeToProjectDelete")
	}
	return nil
}

// ListInitiativeToProjects retrieves all initiative-project links.
//
// Parameters:
//   - first: Number of links to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Links with initiative and project details
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [InitiativeToProjectCreate], [InitiativeToProjectDelete]
func (c *Client) ListInitiativeToProjects(ctx context.Context, first *int64, after *string) (*intgraphql.ListInitiativeToProjects_InitiativeToProjects, error) {
	resp, err := c.gqlClient.ListInitiativeToProjects(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("list initiative to projects query", err)
	}
	return &resp.InitiativeToProjects, nil
}
