package comment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewCreateCommand creates the comment create command.
func NewCreateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a comment on an issue",
		Long: `Create a new comment on a Linear issue.

This operation creates new data and is safe to execute.

Examples:
  linear comment create --issue=ENG-123 --body="This is a comment"
  linear comment create --issue=<uuid> --body="Added fix" --output=json

TIP: Comments support markdown formatting`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runCreate(cmd, client)
		},
	}

	cmd.Flags().String("issue", "", "Issue ID to comment on (required)")
	cmd.MarkFlagRequired("issue")

	cmd.Flags().String("body", "", "Comment body text (markdown) (required)")
	cmd.MarkFlagRequired("body")

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	issueID, _ := cmd.Flags().GetString("issue")
	body, _ := cmd.Flags().GetString("body")

	input := intgraphql.CommentCreateInput{
		IssueID: &issueID,
		Body:    &body,
	}

	result, err := client.CommentCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Comment created on issue %s\n", issueID)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
