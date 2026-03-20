package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// FavoriteCreate stars an issue or project for quick access.
//
// Starred items appear in Linear's "Favorites" section for easy retrieval.
//
// Parameters:
//   - input: Favorite creation parameters (must include either IssueID or ProjectID)
//
// Returns:
//   - Created favorite with ID, type, and referenced resource
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Example (star an issue):
//
//	favorite, err := client.FavoriteCreate(ctx, intgraphql.FavoriteCreateInput{
//	    IssueID: &issueID,
//	})
//	fmt.Printf("Starred issue: %s\n", favorite.Issue.Title)
//
// Example (star a project):
//
//	favorite, err := client.FavoriteCreate(ctx, intgraphql.FavoriteCreateInput{
//	    ProjectID: &projectID,
//	})
//	fmt.Printf("Starred project: %s\n", favorite.Project.Name)
//
// Related: [FavoriteDelete]
func (c *Client) FavoriteCreate(ctx context.Context, input intgraphql.FavoriteCreateInput) (*intgraphql.FavoriteCreate_FavoriteCreate_Favorite, error) {
	resp, err := c.gqlClient.FavoriteCreate(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("FavoriteCreate", err)
	}
	if !resp.FavoriteCreate.Success {
		return nil, errMutationFailed("FavoriteCreate")
	}
	return &resp.FavoriteCreate.Favorite, nil
}

// FavoriteDelete removes a favorite (unstar).
//
// Parameters:
//   - id: Favorite UUID to delete (required)
//
// Returns:
//   - nil: Favorite successfully deleted (item unstarred)
//   - error: Non-nil if delete fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	err := client.FavoriteDelete(ctx, favoriteID)
//	if err != nil {
//	    return fmt.Errorf("failed to unstar: %w", err)
//	}
//
// Related: [FavoriteCreate]
func (c *Client) FavoriteDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.FavoriteDelete(ctx, id)
	if err != nil {
		return wrapGraphQLError("FavoriteDelete", err)
	}
	if !resp.FavoriteDelete.Success {
		return errMutationFailed("FavoriteDelete")
	}
	return nil
}
