package linear

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// Viewer returns the currently authenticated user information.
//
// Use this to:
//   - Verify API authentication is working
//   - Get authenticated user details
//   - Check user permissions (admin status)
//
// Returns:
//   - Viewer.ID: User UUID (always populated)
//   - Viewer.Name: Full name (always populated)
//   - Viewer.Email: Email address (always populated)
//   - Viewer.DisplayName: Display/nick name (always populated)
//   - Viewer.Admin: Whether user is org admin (always populated)
//   - error: Non-nil if authentication fails (401) or network error
//
// Permissions Required: Read
//
// Related: [Users], [User]
//
// Example:
//
//	viewer, err := client.Viewer(ctx)
//	if err != nil {
//	    return fmt.Errorf("auth failed: %w", err)
//	}
//	log.Printf("Authenticated as: %s", viewer.Email)
func (c *Client) Viewer(ctx context.Context) (*intgraphql.Viewer_Viewer, error) {
	if c.config.Transport.TracingEnabled {
		var span trace.Span
		ctx, span = startSpan(ctx, "Viewer")
		defer span.End()
	}

	resp, err := c.gqlClient.Viewer(ctx)
	if err != nil {
		return nil, wrapGraphQLError("viewer query", err)
	}
	return &resp.Viewer, nil
}

// Organization retrieves the workspace organization information.
//
// Returns:
//   - Organization.ID: Organization UUID (always populated)
//   - Organization.Name: Organization name (always populated)
//   - Organization.URLKey: URL identifier (always populated)
//   - Organization.LogoURL: Organization logo URL (may be empty)
//   - Organization.RoadmapEnabled: Whether roadmaps are enabled (always populated)
//   - Organization.CreatedAt: Creation timestamp (always populated)
//   - Organization.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Teams], [Users]
func (c *Client) Organization(ctx context.Context) (*intgraphql.GetOrganization_Organization, error) {
	resp, err := c.gqlClient.GetOrganization(ctx)
	if err != nil {
		return nil, wrapGraphQLError("organization query", err)
	}
	return &resp.Organization, nil
}
