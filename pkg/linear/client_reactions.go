package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ReactionCreate adds an emoji reaction to an issue or comment.
//
// Reactions provide quick feedback without creating a comment.
//
// Parameters:
//   - input: Reaction parameters (emoji, and either IssueID or CommentID)
//
// Returns:
//   - Created reaction with emoji and user
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Example (react to issue):
//
//	reaction, err := client.ReactionCreate(ctx, intgraphql.ReactionCreateInput{
//	    Emoji:   "👍",
//	    IssueID: &issueID,
//	})
//	fmt.Printf("%s reacted with %s\n", reaction.User.Name, reaction.Emoji)
//
// Example (react to comment):
//
//	reaction, err := client.ReactionCreate(ctx, intgraphql.ReactionCreateInput{
//	    Emoji:     "❤️",
//	    CommentID: &commentID,
//	})
//
// Related: [ReactionDelete]
func (c *Client) ReactionCreate(ctx context.Context, input intgraphql.ReactionCreateInput) (*intgraphql.ReactionCreate_ReactionCreate_Reaction, error) {
	resp, err := c.gqlClient.ReactionCreate(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("ReactionCreate", err)
	}
	if !resp.ReactionCreate.Success {
		return nil, errMutationFailed("ReactionCreate")
	}
	return &resp.ReactionCreate.Reaction, nil
}

// ReactionDelete removes a reaction from an issue or comment.
//
// Parameters:
//   - id: Reaction UUID to delete (required)
//
// Returns:
//   - nil: Reaction successfully deleted
//   - error: Non-nil if delete fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	err := client.ReactionDelete(ctx, reactionID)
//	if err != nil {
//	    return fmt.Errorf("failed to remove reaction: %w", err)
//	}
//
// Related: [ReactionCreate]
func (c *Client) ReactionDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.ReactionDelete(ctx, id)
	if err != nil {
		return wrapGraphQLError("ReactionDelete", err)
	}
	if !resp.ReactionDelete.Success {
		return errMutationFailed("ReactionDelete")
	}
	return nil
}
