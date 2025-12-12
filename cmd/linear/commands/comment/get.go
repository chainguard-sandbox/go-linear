package comment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the comment get command.
func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <comment-id>",
		Short: "Get a single comment by ID",
		Long: `Get detailed information about a specific comment.

Retrieve full comment details including body text, author, creation timestamp,
and associated issue information.

Parameters:
  <comment-id>: Comment UUID (required)

Output (--output=json):
  Returns JSON with: id, body, user (author), issue, createdAt, updatedAt

Examples:
  # Get comment by UUID
  linear comment get <comment-uuid>

  # Get comment with JSON output
  linear comment get <comment-uuid> --output=json

TIP: Use 'linear comment list' to discover comment IDs from issues`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runGet(cmd, client, args[0])
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "Comma-separated fields for JSON output (e.g., 'id,body')")

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, commentID string) error {
	ctx := context.Background()

	comment, err := client.Comment(ctx, commentID)
	if err != nil {
		return fmt.Errorf("failed to get comment: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	switch output {
	case "json":
		fieldSelector, err := fieldfilter.New(fieldsSpec)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), comment, true, fieldSelector)
	case "table":
		// Simple table output for single comment
		fmt.Fprintf(cmd.OutOrStdout(), "ID:          %s\n", comment.ID)
		if comment.User != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Author:      %s\n", comment.User.Name)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Body:        %s\n", comment.Body)
		fmt.Fprintf(cmd.OutOrStdout(), "Created:     %s\n", comment.CreatedAt)
		fmt.Fprintf(cmd.OutOrStdout(), "Updated:     %s\n", comment.UpdatedAt)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
