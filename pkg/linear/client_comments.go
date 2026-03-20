package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// Comment retrieves a single comment by ID.
//
// Returns:
//   - Comment.ID: Comment UUID (always populated)
//   - Comment.Body: Comment text in markdown (always populated)
//   - Comment.URL: Linear web URL to comment (always populated)
//   - Comment.User: Comment author with ID, Name (always populated)
//   - Comment.Issue: Parent issue with ID, Title (always populated)
//   - Comment.CreatedAt: Creation timestamp (always populated)
//   - Comment.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if comment not found or query fails
//
// Permissions Required: Read
//
// Related: [Comments], [CommentCreate], [CommentUpdate]
func (c *Client) Comment(ctx context.Context, id string) (*intgraphql.GetComment_Comment, error) {
	defaultLimit := int64(50)
	resp, err := c.gqlClient.GetComment(ctx, id, &defaultLimit, nil)
	if err != nil {
		return nil, wrapGraphQLError("comment query", err)
	}
	return &resp.Comment, nil
}

// Comments retrieves a paginated list of comments.
//
// Parameters:
//   - first: Number of comments to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Comments.Nodes: Array of comments (may be empty)
//   - Comments.PageInfo.HasNextPage: true if more results available
//   - Comments.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Comment], [NewCommentIterator]
func (c *Client) Comments(ctx context.Context, first *int64, after *string) (*intgraphql.ListComments_Comments, error) {
	resp, err := c.gqlClient.ListComments(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("comments query", err)
	}
	return &resp.Comments, nil
}

// CommentsFiltered returns comments matching the specified filter.
func (c *Client) CommentsFiltered(ctx context.Context, first *int64, after *string, filter *intgraphql.CommentFilter) (*intgraphql.ListCommentsFiltered_Comments, error) {
	resp, err := c.gqlClient.ListCommentsFiltered(ctx, first, after, filter)
	if err != nil {
		return nil, wrapGraphQLError("comments filtered query", err)
	}
	return &resp.Comments, nil
}

// CommentCreate creates a new comment on an issue.
//
// Required Input Fields:
//   - IssueID: Issue to comment on (*string, get from Issues())
//   - Body: Comment text in markdown (*string)
//
// Returns:
//   - Comment.ID: Created comment UUID (always populated)
//   - Comment.Body: Comment text (always populated)
//   - Comment.CreatedAt: Creation timestamp (always populated)
//   - error: Non-nil if issue not found or mutation fails
//
// Permissions Required: Write (or comments:create)
//
// Related: [CommentUpdate], [CommentDelete], [Comments]
//
// Example:
//
//	issueID := "issue-uuid"
//	body := "This looks like a duplicate of #123"
//
//	comment, err := client.CommentCreate(ctx, CommentCreateInput{
//	    IssueID: &issueID,
//	    Body: &body,
//	})
func (c *Client) CommentCreate(ctx context.Context, input intgraphql.CommentCreateInput) (*intgraphql.CreateComment_CommentCreate_Comment, error) {
	resp, err := c.gqlClient.CreateComment(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("CommentCreate", err)
	}

	if !resp.CommentCreate.Success {
		return nil, errMutationFailed("CommentCreate")
	}

	return &resp.CommentCreate.Comment, nil
}

// CommentUpdate updates an existing comment.
func (c *Client) CommentUpdate(ctx context.Context, id string, input intgraphql.CommentUpdateInput) (*intgraphql.UpdateComment_CommentUpdate_Comment, error) {
	resp, err := c.gqlClient.UpdateComment(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("CommentUpdate", err)
	}

	if !resp.CommentUpdate.Success {
		return nil, errMutationFailed("CommentUpdate")
	}

	return &resp.CommentUpdate.Comment, nil
}

// CommentDelete deletes a comment by ID.
func (c *Client) CommentDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.DeleteComment(ctx, id)
	if err != nil {
		return wrapGraphQLError("CommentDelete", err)
	}

	if !resp.CommentDelete.Success {
		return errMutationFailed("CommentDelete")
	}

	return nil
}

// GetCommentWithChildren retrieves a comment with configurable child limit and pagination.
func (c *Client) GetCommentWithChildren(ctx context.Context, id string, childrenLimit *int64, childrenAfter *string) (*intgraphql.GetComment_Comment, error) {
	resp, err := c.gqlClient.GetComment(ctx, id, childrenLimit, childrenAfter)
	if err != nil {
		return nil, wrapGraphQLError("comment query", err)
	}
	return &resp.Comment, nil
}
