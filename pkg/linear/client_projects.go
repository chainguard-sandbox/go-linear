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

// ProjectArchive archives a project by ID.
//
// Note: Linear recommends using ProjectDelete instead of archive.
//
// Parameters:
//   - id: Project UUID to archive (required)
//
// Returns:
//   - nil: Project successfully archived
//   - error: Non-nil if archive fails or Success is false
//
// Permissions Required: Write
//
// Related: [ProjectUnarchive], [ProjectDelete]
func (c *Client) ProjectArchive(ctx context.Context, id string) error {
	resp, err := c.gqlClient.ArchiveProject(ctx, id)
	if err != nil {
		return wrapGraphQLError("ProjectArchive", err)
	}
	if !resp.ProjectArchive.Success {
		return errMutationFailed("ProjectArchive")
	}
	return nil
}

// ProjectUnarchive restores an archived project by ID.
//
// Parameters:
//   - id: Project UUID to restore (required)
//
// Returns:
//   - nil: Project successfully restored
//   - error: Non-nil if unarchive fails or Success is false
//
// Permissions Required: Write
//
// Related: [ProjectDelete], [ProjectCreate]
func (c *Client) ProjectUnarchive(ctx context.Context, id string) error {
	resp, err := c.gqlClient.UnarchiveProject(ctx, id)
	if err != nil {
		return wrapGraphQLError("ProjectUnarchive", err)
	}
	if !resp.ProjectUnarchive.Success {
		return errMutationFailed("ProjectUnarchive")
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

// ProjectUpdateCreate creates a status update for a project.
//
// Parameters:
//   - input: Status update creation parameters
//
// Required fields:
//   - ProjectID: UUID of the project (required)
//   - Body: Update body in markdown format (required)
//
// Optional fields:
//   - Health: Project health status (onTrack, atRisk, offTrack)
//
// Returns:
//   - Created project update with ID, body, health, and dates
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Related: [ProjectUpdateDelete], [ProjectUpdate], [ListProjectUpdates]
func (c *Client) ProjectUpdateCreate(ctx context.Context, input intgraphql.ProjectUpdateCreateInput) (*intgraphql.CreateProjectUpdate_ProjectUpdateCreate_ProjectUpdate, error) {
	resp, err := c.gqlClient.CreateProjectUpdate(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("ProjectUpdateCreate", err)
	}
	if !resp.ProjectUpdateCreate.Success {
		return nil, errMutationFailed("ProjectUpdateCreate")
	}
	return &resp.ProjectUpdateCreate.ProjectUpdate, nil
}

// GetProjectUpdate retrieves a single project status update by ID.
//
// Returns:
//   - ProjectUpdate with ID, body, health, user, dates, and URL
//   - error: Non-nil if update not found or query fails
//
// Permissions Required: Read
//
// Related: [ListProjectUpdates], [ProjectUpdateCreate]
func (c *Client) GetProjectUpdate(ctx context.Context, id string) (*intgraphql.GetProjectUpdate_ProjectUpdate, error) {
	resp, err := c.gqlClient.GetProjectUpdate(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("project update query", err)
	}
	return &resp.ProjectUpdate, nil
}

// ListProjectUpdates retrieves status updates for a project.
//
// Parameters:
//   - projectID: Project UUID (required)
//   - first: Number of updates to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Project with nested project updates list
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [ProjectUpdate], [ProjectUpdateCreate]
func (c *Client) ListProjectUpdates(ctx context.Context, projectID string, first *int64, after *string) (*intgraphql.ListProjectUpdates_Project, error) {
	resp, err := c.gqlClient.ListProjectUpdates(ctx, projectID, first, after)
	if err != nil {
		return nil, wrapGraphQLError("list project updates query", err)
	}
	return &resp.Project, nil
}

// ProjectUpdateDelete deletes a project status update by ID.
//
// Parameters:
//   - id: Project update UUID to delete (required)
//
// Returns:
//   - nil: Update successfully deleted
//   - error: Non-nil if delete fails or Success is false
//
// Permissions Required: Write
//
// Related: [ProjectUpdateCreate], [ProjectUpdate]
func (c *Client) ProjectUpdateDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.DeleteProjectUpdate(ctx, id)
	if err != nil {
		return wrapGraphQLError("ProjectUpdateDelete", err)
	}
	if !resp.ProjectUpdateDelete.Success {
		return errMutationFailed("ProjectUpdateDelete")
	}
	return nil
}

// ProjectStatuses retrieves all project statuses for the organization.
//
// Project statuses represent workflow states like "Backlog", "In Progress", "Completed".
//
// Returns:
//   - []ProjectStatus: Array of project statuses with ID, Name, Type, Color
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [ProjectUpdate]
func (c *Client) ProjectStatuses(ctx context.Context) ([]*intgraphql.ListProjectStatuses_Organization_ProjectStatuses, error) {
	resp, err := c.gqlClient.ListProjectStatuses(ctx)
	if err != nil {
		return nil, wrapGraphQLError("list project statuses query", err)
	}
	return resp.Organization.ProjectStatuses, nil
}
