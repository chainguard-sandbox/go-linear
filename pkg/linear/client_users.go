package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// User retrieves a single user by ID.
//
// Returns:
//   - User.ID: User UUID (always populated)
//   - User.Name: Full name (always populated)
//   - User.Email: Email address (always populated)
//   - User.DisplayName: Display/nick name (always populated)
//   - User.Admin: Whether user is org admin (always populated)
//   - User.Active: Whether account is active (always populated)
//   - User.AvatarURL: Avatar image URL (may be empty)
//   - User.CreatedAt: Account creation timestamp (always populated)
//   - User.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if user not found or query fails
//
// Permissions Required: Read
//
// Related: [Users], [Viewer]
func (c *Client) User(ctx context.Context, id string) (*intgraphql.GetUser_User, error) {
	resp, err := c.gqlClient.GetUser(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("user query", err)
	}
	return &resp.User, nil
}

// Users retrieves a paginated list of users in the organization.
//
// Parameters:
//   - first: Number of users to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Users.Nodes: Array of users (may be empty)
//   - Users.PageInfo.HasNextPage: true if more results available
//   - Users.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [User], [Viewer]
func (c *Client) Users(ctx context.Context, first *int64, after *string) (*intgraphql.ListUsers_Users, error) {
	resp, err := c.gqlClient.ListUsers(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("users query", err)
	}
	return &resp.Users, nil
}

// UsersFiltered returns users matching the specified filter.
func (c *Client) UsersFiltered(ctx context.Context, first *int64, after *string, filter *intgraphql.UserFilter) (*intgraphql.ListUsersFiltered_Users, error) {
	resp, err := c.gqlClient.ListUsersFiltered(ctx, first, after, filter)
	if err != nil {
		return nil, wrapGraphQLError("users filtered query", err)
	}
	return &resp.Users, nil
}
