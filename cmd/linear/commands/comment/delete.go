package comment

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewDeleteCommand creates the comment delete command.
func NewDeleteCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a comment permanently",
		Long: `⚠️ Delete comment. Cannot be undone. Prompts unless --yes.

Example: go-linear comment delete <uuid>

Related: comment_list, comment_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runDelete(cmd, client, args[0])
		},
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runDelete(cmd *cobra.Command, client *linear.Client, commentID string) error {
	ctx := cmd.Context()

	// Confirmation prompt unless --yes
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		fmt.Fprintf(cmd.OutOrStderr(), "⚠️  Are you sure you want to delete comment %s? This cannot be undone.\n", commentID)
		fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)

		if !strings.EqualFold(response, "yes") {
			fmt.Fprintln(cmd.OutOrStderr(), "Canceled.")
			return nil
		}
	}

	// Delete comment
	err := client.CommentDelete(ctx, commentID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
			"success":   true,
			"commentId": commentID,
		}, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Comment deleted successfully\n")
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
