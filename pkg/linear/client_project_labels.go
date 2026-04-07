package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ProjectLabels retrieves a paginated list of project labels.
func (c *Client) ProjectLabels(ctx context.Context, first *int64, after *string) (*intgraphql.ListProjectLabels_ProjectLabels, error) {
	resp, err := c.gqlClient.ListProjectLabels(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("project labels query", err)
	}
	return &resp.ProjectLabels, nil
}

// ProjectLabel retrieves a single project label by ID.
func (c *Client) ProjectLabel(ctx context.Context, id string) (*intgraphql.GetProjectLabel_ProjectLabel, error) {
	resp, err := c.gqlClient.GetProjectLabel(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("project label query", err)
	}
	return &resp.ProjectLabel, nil
}

// ProjectLabelCreate creates a new project label.
func (c *Client) ProjectLabelCreate(ctx context.Context, input intgraphql.ProjectLabelCreateInput) (*intgraphql.ProjectLabelCreate_ProjectLabelCreate_ProjectLabel, error) {
	resp, err := c.gqlClient.ProjectLabelCreate(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("ProjectLabelCreate", err)
	}
	if !resp.ProjectLabelCreate.Success {
		return nil, errMutationFailed("ProjectLabelCreate")
	}
	return &resp.ProjectLabelCreate.ProjectLabel, nil
}

// ProjectLabelUpdate updates an existing project label.
func (c *Client) ProjectLabelUpdate(ctx context.Context, id string, input intgraphql.ProjectLabelUpdateInput) (*intgraphql.ProjectLabelUpdate_ProjectLabelUpdate_ProjectLabel, error) {
	resp, err := c.gqlClient.ProjectLabelUpdate(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("ProjectLabelUpdate", err)
	}
	if !resp.ProjectLabelUpdate.Success {
		return nil, errMutationFailed("ProjectLabelUpdate")
	}
	return &resp.ProjectLabelUpdate.ProjectLabel, nil
}

// ProjectLabelDelete deletes a project label by ID.
func (c *Client) ProjectLabelDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.ProjectLabelDelete(ctx, id)
	if err != nil {
		return wrapGraphQLError("ProjectLabelDelete", err)
	}
	if !resp.ProjectLabelDelete.Success {
		return errMutationFailed("ProjectLabelDelete")
	}
	return nil
}
