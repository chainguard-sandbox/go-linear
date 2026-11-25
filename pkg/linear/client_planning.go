package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// The Planning queries are read-only lists for planning and delivery
// tracking. Each selects the fields needed to join projects, initiatives,
// teams, milestones and project relations to one another and to see where
// each stands: status type, start and completion times, the latest status
// update, and visibility. Linked records are returned by ID only.
//
// They are separate from Projects, Initiatives, Teams and ProjectRelations,
// whose fields and query cost are unchanged.

// PlanningProjects retrieves a paginated list of projects with their status
// type, start, completion and cancellation times, target date, latest status
// update and team IDs.
//
// A project's teams are a nested connection of up to 50. When
// Teams.PageInfo.HasNextPage is true the project has more teams than the
// query returned.
//
// Permissions Required: Read
func (c *Client) PlanningProjects(ctx context.Context, first *int64, after *string) (*intgraphql.PlanningProjects_Projects, error) {
	resp, err := c.gqlClient.PlanningProjects(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("planning projects query", err)
	}
	return &resp.Projects, nil
}

// PlanningInitiatives retrieves a paginated list of initiatives with their
// status, visibility, target date, parent, latest status update, and whether
// each has sub-initiatives.
//
// SubInitiatives holds at most one node. It says whether the initiative has
// any sub-initiatives, not which.
//
// Permissions Required: Read
func (c *Client) PlanningInitiatives(ctx context.Context, first *int64, after *string) (*intgraphql.PlanningInitiatives_Initiatives, error) {
	resp, err := c.gqlClient.PlanningInitiatives(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("planning initiatives query", err)
	}
	return &resp.Initiatives, nil
}

// PlanningTeams retrieves a paginated list of teams with their visibility,
// which replaces the deprecated private flag.
//
// Permissions Required: Read
func (c *Client) PlanningTeams(ctx context.Context, first *int64, after *string) (*intgraphql.PlanningTeams_Teams, error) {
	resp, err := c.gqlClient.PlanningTeams(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("planning teams query", err)
	}
	return &resp.Teams, nil
}

// PlanningProjectRelations retrieves a paginated list of project relations
// with the milestone at each end, for a relation anchored to a milestone.
//
// Permissions Required: Read
func (c *Client) PlanningProjectRelations(ctx context.Context, first *int64, after *string) (*intgraphql.PlanningProjectRelations_ProjectRelations, error) {
	resp, err := c.gqlClient.PlanningProjectRelations(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("planning project relations query", err)
	}
	return &resp.ProjectRelations, nil
}

// PlanningProjectMilestones retrieves a paginated list of the milestones of
// every project, each with its status, target date and project ID.
//
// Permissions Required: Read
func (c *Client) PlanningProjectMilestones(ctx context.Context, first *int64, after *string) (*intgraphql.PlanningProjectMilestones_ProjectMilestones, error) {
	resp, err := c.gqlClient.PlanningProjectMilestones(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("planning project milestones query", err)
	}
	return &resp.ProjectMilestones, nil
}
