package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// Template retrieves a single issue template by ID.
//
// Returns:
//   - Template.ID: Template UUID (always populated)
//   - Template.Name: Template name (always populated)
//   - Template.Description: Template description (may be empty)
//   - Template.Type: Template type (always populated)
//   - Template.CreatedAt: Creation timestamp (always populated)
//   - Template.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if template not found or query fails
//
// Permissions Required: Read
//
// Related: [Templates]
func (c *Client) Template(ctx context.Context, id string) (*intgraphql.GetTemplate_Template, error) {
	resp, err := c.gqlClient.GetTemplate(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("template query", err)
	}
	return &resp.Template, nil
}

// Templates retrieves all templates.
func (c *Client) Templates(ctx context.Context) ([]*intgraphql.ListTemplates_Templates, error) {
	resp, err := c.gqlClient.ListTemplates(ctx)
	if err != nil {
		return nil, wrapGraphQLError("templates query", err)
	}
	return resp.Templates, nil
}
