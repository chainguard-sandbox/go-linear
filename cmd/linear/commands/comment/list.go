package comment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the comment list command.
func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List comments",
		Long: `List comments from Linear.

Use this to browse comments across all issues or find specific discussion threads.

Output (--output=json):
  Returns JSON with:
  - nodes: Array of comments
  - pageInfo: {hasNextPage: bool, endCursor: string}

  Each comment contains:
  - id: Comment UUID
  - body: Comment text (markdown)
  - user: Author {id, name, email}
  - issue: Associated issue reference
  - createdAt: Creation timestamp

Examples:
  # List comments
  linear comment list

  # List with limit
  linear comment list --limit=100

  # JSON output for parsing
  linear comment list --output=json

Related Commands:
  - linear comment get - Get single comment details
  - linear comment create - Create new comment
  - linear comment update - Modify comment text`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client)
		},
	}

	cmd.Flags().IntP("limit", "l", 50, "Number of comments to return")
	cmd.Flags().String("after", "", "Cursor for pagination")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	after, _ := cmd.Flags().GetString("after")
	var afterPtr *string
	if after != "" {
		afterPtr = &after
	}

	comments, err := client.Comments(ctx, &first, afterPtr)
	if err != nil {
		return fmt.Errorf("failed to list comments: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), comments, true)
	case "table":
		if len(comments.Nodes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No comments found")
			return nil
		}
		for _, comment := range comments.Nodes {
			fmt.Fprintf(cmd.OutOrStdout(), "Comment by %s:\n", comment.User.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n\n", comment.Body)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
