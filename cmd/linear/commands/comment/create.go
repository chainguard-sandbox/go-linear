package comment

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewCreateCommand creates the comment create command.
func NewCreateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a comment on an issue",
		Long: `Create comment on issue. Supports markdown and threading. Safe operation.

Required: --issue (ID from issue_list), --body
Optional: --parent (comment ID to create a threaded reply)

Example: go-linear comment create --issue=ENG-123 --body="Fixed in PR #42"
Example: go-linear comment create --issue=ENG-123 --body="Good point!" --parent=<comment-id>

Related: comment_list, comment_get, issue_get`,
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

	cmd.Flags().String("parent", "", "Parent comment ID to create a threaded reply")

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

	// Add parent comment for threading
	if parent, _ := cmd.Flags().GetString("parent"); parent != "" {
		input.ParentID = &parent
	}

	result, err := client.CommentCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
}
