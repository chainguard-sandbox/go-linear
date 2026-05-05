package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ProjectRelations retrieves a paginated list of project relations.
func (c *Client) ProjectRelations(ctx context.Context, first *int64, after *string) (*intgraphql.ListProjectRelations_ProjectRelations, error) {
	resp, err := c.gqlClient.ListProjectRelations(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("project relations query", err)
	}
	return &resp.ProjectRelations, nil
}

// ProjectRelationCreate creates a new project relation.
func (c *Client) ProjectRelationCreate(ctx context.Context, input intgraphql.ProjectRelationCreateInput) (*intgraphql.ProjectRelationCreate_ProjectRelationCreate_ProjectRelation, error) {
	resp, err := c.gqlClient.ProjectRelationCreate(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("ProjectRelationCreate", err)
	}
	if !resp.ProjectRelationCreate.Success {
		return nil, errMutationFailed("ProjectRelationCreate")
	}
	return &resp.ProjectRelationCreate.ProjectRelation, nil
}

// ProjectRelationUpdate updates an existing project relation.
func (c *Client) ProjectRelationUpdate(ctx context.Context, id string, input intgraphql.ProjectRelationUpdateInput) (*intgraphql.ProjectRelationUpdate_ProjectRelationUpdate_ProjectRelation, error) {
	resp, err := c.gqlClient.ProjectRelationUpdate(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("ProjectRelationUpdate", err)
	}
	if !resp.ProjectRelationUpdate.Success {
		return nil, errMutationFailed("ProjectRelationUpdate")
	}
	return &resp.ProjectRelationUpdate.ProjectRelation, nil
}

// ProjectRelationDelete deletes a project relation by ID.
func (c *Client) ProjectRelationDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.ProjectRelationDelete(ctx, id)
	if err != nil {
		return wrapGraphQLError("ProjectRelationDelete", err)
	}
	if !resp.ProjectRelationDelete.Success {
		return errMutationFailed("ProjectRelationDelete")
	}
	return nil
}
