package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// Team retrieves a single team by ID.
//
// Returns:
//   - Team.ID: Team UUID (always populated)
//   - Team.Name: Team name (always populated)
//   - Team.Key: Team identifier used in URLs (always populated)
//   - Team.Description: Team description (may be empty)
//   - Team.Private: Whether team is private (always populated)
//   - Team.Timezone: Team timezone (always populated)
//   - Team.Icon: Team icon emoji (may be empty)
//   - Team.Color: Team color hex code (may be empty)
//   - Team.CreatedAt: Creation timestamp (always populated)
//   - Team.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if team not found or query fails
//
// Permissions Required: Read
//
// Related: [Teams], [TeamMemberships]
func (c *Client) Team(ctx context.Context, id string) (*intgraphql.GetTeam_Team, error) {
	resp, err := c.gqlClient.GetTeam(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("team query", err)
	}
	return &resp.Team, nil
}

// Teams retrieves a paginated list of teams.
//
// Parameters:
//   - first: Number of teams to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Teams.Nodes: Array of teams (may be empty)
//   - Teams.PageInfo.HasNextPage: true if more results available
//   - Teams.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Team], [NewTeamIterator]
func (c *Client) Teams(ctx context.Context, first *int64, after *string) (*intgraphql.ListTeams_Teams, error) {
	resp, err := c.gqlClient.ListTeams(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("teams query", err)
	}
	return &resp.Teams, nil
}

// TeamsFiltered returns teams matching the specified filter.
func (c *Client) TeamsFiltered(ctx context.Context, first *int64, after *string, filter *intgraphql.TeamFilter) (*intgraphql.ListTeamsFiltered_Teams, error) {
	resp, err := c.gqlClient.ListTeamsFiltered(ctx, first, after, filter)
	if err != nil {
		return nil, wrapGraphQLError("teams filtered query", err)
	}
	return &resp.Teams, nil
}

// TeamCreate creates a new team.
func (c *Client) TeamCreate(ctx context.Context, input intgraphql.TeamCreateInput) (*intgraphql.CreateTeam_TeamCreate_Team, error) {
	resp, err := c.gqlClient.CreateTeam(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("TeamCreate", err)
	}

	if !resp.TeamCreate.Success {
		return nil, errMutationFailed("TeamCreate")
	}

	return resp.TeamCreate.Team, nil
}

// TeamUpdate updates an existing team.
func (c *Client) TeamUpdate(ctx context.Context, id string, input intgraphql.TeamUpdateInput) (*intgraphql.UpdateTeam_TeamUpdate_Team, error) {
	resp, err := c.gqlClient.UpdateTeam(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("TeamUpdate", err)
	}

	if !resp.TeamUpdate.Success {
		return nil, errMutationFailed("TeamUpdate")
	}

	return resp.TeamUpdate.Team, nil
}

// TeamDelete deletes a team by ID.
func (c *Client) TeamDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.DeleteTeam(ctx, id)
	if err != nil {
		return wrapGraphQLError("TeamDelete", err)
	}

	if !resp.TeamDelete.Success {
		return errMutationFailed("TeamDelete")
	}

	return nil
}

// TeamUnarchive restores an archived team by ID.
//
// Parameters:
//   - id: Team UUID to restore (required)
//
// Returns:
//   - nil: Team successfully restored
//   - error: Non-nil if unarchive fails or Success is false
//
// Permissions Required: Write
//
// Related: [TeamDelete], [TeamCreate]
func (c *Client) TeamUnarchive(ctx context.Context, id string) error {
	resp, err := c.gqlClient.UnarchiveTeam(ctx, id)
	if err != nil {
		return wrapGraphQLError("TeamUnarchive", err)
	}
	if !resp.TeamUnarchive.Success {
		return errMutationFailed("TeamUnarchive")
	}
	return nil
}

// TeamMemberships retrieves a paginated list of team memberships.
//
// Parameters:
//   - teamID: Team UUID (required)
//   - first: Number of memberships to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Team with memberships containing user info
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [TeamMembershipCreate], [TeamMembershipDelete]
func (c *Client) TeamMemberships(ctx context.Context, teamID string, first *int64, after *string) (*intgraphql.ListTeamMemberships_Team, error) {
	resp, err := c.gqlClient.ListTeamMemberships(ctx, teamID, first, after)
	if err != nil {
		return nil, wrapGraphQLError("team memberships query", err)
	}
	return &resp.Team, nil
}

// TeamMembershipCreate adds a user to a team.
//
// Parameters:
//   - input: Membership creation parameters (teamId, userId)
//
// Returns:
//   - Created membership with team and user details
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Admin
//
// Related: [TeamMembershipDelete], [Team]
func (c *Client) TeamMembershipCreate(ctx context.Context, input intgraphql.TeamMembershipCreateInput) (*intgraphql.AddTeamMember_TeamMembershipCreate_TeamMembership, error) {
	resp, err := c.gqlClient.AddTeamMember(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("TeamMembershipCreate", err)
	}
	if !resp.TeamMembershipCreate.Success {
		return nil, errMutationFailed("TeamMembershipCreate")
	}
	return resp.TeamMembershipCreate.TeamMembership, nil
}

// TeamMembershipDelete removes a user from a team.
//
// Parameters:
//   - id: TeamMembership UUID to delete (required)
//
// Returns:
//   - nil: Membership successfully deleted
//   - error: Non-nil if delete fails or Success is false
//
// Permissions Required: Admin
//
// Related: [TeamMembershipCreate], [Team]
func (c *Client) TeamMembershipDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.RemoveTeamMember(ctx, id)
	if err != nil {
		return wrapGraphQLError("TeamMembershipDelete", err)
	}
	if !resp.TeamMembershipDelete.Success {
		return errMutationFailed("TeamMembershipDelete")
	}
	return nil
}
