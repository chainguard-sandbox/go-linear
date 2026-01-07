package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// Roadmap retrieves a single product roadmap by ID.
//
// Returns:
//   - Roadmap.ID: Roadmap UUID (always populated)
//   - Roadmap.Name: Roadmap name (always populated)
//   - Roadmap.Description: Roadmap description (may be empty)
//   - Roadmap.SlugID: URL-friendly identifier (always populated)
//   - Roadmap.CreatedAt: Creation timestamp (always populated)
//   - Roadmap.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if roadmap not found or query fails
//
// Permissions Required: Read
//
// Related: [Roadmaps]
func (c *Client) Roadmap(ctx context.Context, id string) (*intgraphql.GetRoadmap_Roadmap, error) {
	resp, err := c.gqlClient.GetRoadmap(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("roadmap query", err)
	}
	return &resp.Roadmap, nil
}

// Roadmaps retrieves a paginated list of product roadmaps.
//
// Parameters:
//   - first: Number of roadmaps to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Roadmaps.Nodes: Array of roadmaps (may be empty)
//   - Roadmaps.PageInfo.HasNextPage: true if more results available
//   - Roadmaps.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Roadmap]
func (c *Client) Roadmaps(ctx context.Context, first *int64, after *string) (*intgraphql.ListRoadmaps_Roadmaps, error) {
	resp, err := c.gqlClient.ListRoadmaps(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("roadmaps query", err)
	}
	return &resp.Roadmaps, nil
}
