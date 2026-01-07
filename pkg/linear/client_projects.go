package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// Project retrieves a single project by ID.
//
// Returns:
//   - Project.ID: Project UUID (always populated)
//   - Project.Name: Project name (always populated)
//   - Project.Description: Project description (may be empty)
//   - Project.State: Project state (backlog/planned/started/completed/canceled)
//   - Project.Progress: Completion progress 0.0-1.0 (always populated)
//   - Project.TargetDate: Target completion date (may be empty)
//   - Project.URL: Linear web URL (always populated)
//   - Project.Icon: Project icon (may be empty)
//   - Project.Color: Project color hex code (may be empty)
//   - Project.Lead: Project lead user with ID, Name, Email (may be nil)
//   - Project.CreatedAt: Creation timestamp (always populated)
//   - Project.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if project not found or query fails
//
// Permissions Required: Read
//
// Related: [Projects], [ProjectCreate], [ProjectUpdate]
func (c *Client) Project(ctx context.Context, id string) (*intgraphql.GetProject_Project, error) {
	resp, err := c.gqlClient.GetProject(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("project query", err)
	}
	return &resp.Project, nil
}

// Projects retrieves a paginated list of projects.
//
// Parameters:
//   - first: Number of projects to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Projects.Nodes: Array of projects (may be empty)
//   - Projects.PageInfo.HasNextPage: true if more results available
//   - Projects.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Project], [NewProjectIterator]
func (c *Client) Projects(ctx context.Context, first *int64, after *string) (*intgraphql.ListProjects_Projects, error) {
	resp, err := c.gqlClient.ListProjects(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("projects query", err)
	}
	return &resp.Projects, nil
}

// ProjectsFiltered returns projects matching the specified filter.
func (c *Client) ProjectsFiltered(ctx context.Context, first *int64, after *string, filter *intgraphql.ProjectFilter) (*intgraphql.ListProjectsFiltered_Projects, error) {
	resp, err := c.gqlClient.ListProjectsFiltered(ctx, first, after, filter)
	if err != nil {
		return nil, wrapGraphQLError("projects filtered query", err)
	}
	return &resp.Projects, nil
}

// ProjectCreate creates a new project.
func (c *Client) ProjectCreate(ctx context.Context, input intgraphql.ProjectCreateInput) (*intgraphql.CreateProject_ProjectCreate_Project, error) {
	resp, err := c.gqlClient.CreateProject(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("ProjectCreate", err)
	}

	if !resp.ProjectCreate.Success {
		return nil, errMutationFailed("ProjectCreate")
	}

	return resp.ProjectCreate.Project, nil
}

// ProjectUpdate updates an existing project.
func (c *Client) ProjectUpdate(ctx context.Context, id string, input intgraphql.ProjectUpdateInput) (*intgraphql.UpdateProject_ProjectUpdate_Project, error) {
	resp, err := c.gqlClient.UpdateProject(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("ProjectUpdate", err)
	}

	if !resp.ProjectUpdate.Success {
		return nil, errMutationFailed("ProjectUpdate")
	}

	return resp.ProjectUpdate.Project, nil
}

// ProjectDelete deletes a project by ID.
func (c *Client) ProjectDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.DeleteProject(ctx, id)
	if err != nil {
		return wrapGraphQLError("ProjectDelete", err)
	}

	if !resp.ProjectDelete.Success {
		return errMutationFailed("ProjectDelete")
	}

	return nil
}

// ProjectMilestoneCreate creates a new milestone within a project.
//
// Milestones represent phases or stages within a project (e.g., "Q1 2025", "Beta Launch", "v1.0 Release").
//
// Parameters:
//   - input: Milestone creation parameters
//
// Required fields:
//   - ProjectID: UUID of the project (required)
//   - Name: Milestone name (required)
//
// Optional fields:
//   - Description: Milestone description
//   - TargetDate: Due date in ISO 8601 format (YYYY-MM-DD)
//   - SortOrder: Display order (lower numbers first)
//
// Returns:
//   - Created milestone with ID, name, dates, and project
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	name := "Beta Launch"
//	targetDate := "2025-03-01"
//	milestone, err := client.ProjectMilestoneCreate(ctx, intgraphql.ProjectMilestoneCreateInput{
//	    ProjectID:  &projectID,
//	    Name:       &name,
//	    TargetDate: &targetDate,
//	})
//
// Related: [ProjectMilestoneUpdate], [ProjectMilestoneDelete], [Projects]
func (c *Client) ProjectMilestoneCreate(ctx context.Context, input intgraphql.ProjectMilestoneCreateInput) (*intgraphql.ProjectMilestoneCreate_ProjectMilestoneCreate_ProjectMilestone, error) {
	resp, err := c.gqlClient.ProjectMilestoneCreate(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("ProjectMilestoneCreate", err)
	}
	if !resp.ProjectMilestoneCreate.Success {
		return nil, errMutationFailed("ProjectMilestoneCreate")
	}
	return &resp.ProjectMilestoneCreate.ProjectMilestone, nil
}

// ProjectMilestoneUpdate updates an existing project milestone.
//
// Parameters:
//   - id: Milestone UUID to update (required)
//   - input: Fields to update (all optional, nil = unchanged)
//
// Optional fields:
//   - Name: Update milestone name
//   - Description: Update description
//   - TargetDate: Update target date (ISO 8601 format)
//   - SortOrder: Update display order
//
// Returns:
//   - Updated milestone with new values
//   - error: Non-nil if update fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	updatedTargetDate := "2025-04-01"
//	milestone, err := client.ProjectMilestoneUpdate(ctx, milestoneID, intgraphql.ProjectMilestoneUpdateInput{
//	    TargetDate: &updatedTargetDate,
//	})
//
// Related: [ProjectMilestoneCreate], [ProjectMilestoneDelete]
func (c *Client) ProjectMilestoneUpdate(ctx context.Context, id string, input intgraphql.ProjectMilestoneUpdateInput) (*intgraphql.ProjectMilestoneUpdate_ProjectMilestoneUpdate_ProjectMilestone, error) {
	resp, err := c.gqlClient.ProjectMilestoneUpdate(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("ProjectMilestoneUpdate", err)
	}
	if !resp.ProjectMilestoneUpdate.Success {
		return nil, errMutationFailed("ProjectMilestoneUpdate")
	}
	return &resp.ProjectMilestoneUpdate.ProjectMilestone, nil
}

// ProjectMilestoneDelete deletes a project milestone.
//
// Parameters:
//   - id: Milestone UUID to delete (required)
//
// Returns:
//   - nil: Milestone successfully deleted
//   - error: Non-nil if delete fails or Success is false
//
// Note: Issues associated with the milestone are not deleted.
//
// Permissions Required: Write
//
// Example:
//
//	err := client.ProjectMilestoneDelete(ctx, milestoneID)
//	if err != nil {
//	    return fmt.Errorf("failed to delete milestone: %w", err)
//	}
//
// Related: [ProjectMilestoneCreate], [ProjectMilestoneUpdate]
func (c *Client) ProjectMilestoneDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.ProjectMilestoneDelete(ctx, id)
	if err != nil {
		return wrapGraphQLError("ProjectMilestoneDelete", err)
	}
	if !resp.ProjectMilestoneDelete.Success {
		return errMutationFailed("ProjectMilestoneDelete")
	}
	return nil
}
