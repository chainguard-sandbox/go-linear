package comment

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewCreateCommand creates the comment create command.
func NewCreateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a comment on an issue",
		Long: `Create comment on issue. Supports markdown. Safe operation.

Required: --issue (ID from issue_list), --body

Example: go-linear comment create --issue=ENG-123 --body="Fixed in PR #42" --output=json

Related: comment_list, issue_get`,
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
	_ = cmd.MarkFlagRequired("issue")

	cmd.Flags().String("body", "", "Comment body text (markdown) (required)")
	_ = cmd.MarkFlagRequired("body")

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

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
