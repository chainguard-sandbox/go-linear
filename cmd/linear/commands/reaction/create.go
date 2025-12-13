package reaction

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewCreateCommand creates the reaction create command.
func NewCreateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Add an emoji reaction to an issue or comment",
		Long: `Add emoji reaction. Safe operation. Must specify exactly one of: --issue, --comment.

Required: --emoji (single emoji like 👍)

Example: go-linear-cli reaction create --issue=ENG-123 --emoji=👍 --output=json

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
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

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

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		resourceType := "issue"
		if commentID != "" {
			resourceType = "comment"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Added %s reaction to %s\n", emoji, resourceType)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
