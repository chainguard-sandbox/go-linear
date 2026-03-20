package reaction

import (
	"fmt"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewCreateCommand creates the reaction create command.
func NewCreateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Add an emoji reaction to an issue or comment",
		Long: `Add emoji reaction. Safe operation. Must specify exactly one of: --issue, --comment.

Required: --emoji (single emoji like thumbs-up)

Example: go-linear reaction create --issue=ENG-123 --emoji=thumbs-up

Related: reaction_delete, comment_create`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runCreate(cmd, client)
		},
	}

	_ = cmd.MarkFlagRequired("emoji")
	cmd.Flags().String("emoji", "", "Single emoji character (required)")
	cmd.Flags().String("issue", "", "Issue identifier or UUID")
	cmd.Flags().String("comment", "", "Comment UUID")

	return cmd
}

func runCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	emoji, _ := cmd.Flags().GetString("emoji")
	issueID, _ := cmd.Flags().GetString("issue")
	commentID, _ := cmd.Flags().GetString("comment")

	// Validate exactly one resource type specified
	if issueID == "" && commentID == "" {
		return fmt.Errorf("must specify one of: --issue or --comment")
	}
	if issueID != "" && commentID != "" {
		return fmt.Errorf("must specify exactly one of --issue or --comment, not both")
	}

	// Validate emoji is a single character
	if utf8.RuneCountInString(emoji) != 1 {
		return fmt.Errorf("emoji must be a single character, got: %s", emoji)
	}

	// Resolve issue identifier to UUID if needed
	res := resolver.New(client)
	if issueID != "" {
		resolvedID, err := res.ResolveIssue(ctx, issueID)
		if err != nil {
			return fmt.Errorf("failed to resolve issue: %w", err)
		}
		issueID = resolvedID
	}

	// Build input
	input := intgraphql.ReactionCreateInput{
		Emoji: emoji,
	}
	if issueID != "" {
		input.IssueID = &issueID
	}
	if commentID != "" {
		input.CommentID = &commentID
	}

	result, err := client.ReactionCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create reaction: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
}
