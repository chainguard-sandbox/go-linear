package comment

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewDeleteCommand creates the comment delete command.
func NewDeleteCommand(clientFactory cli.ClientFactory) *cobra.Command {
	confirmFlags := &cli.ConfirmationFlags{}
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a comment permanently",
		Long: `Delete comment. Cannot be undone. Prompts unless --yes.

Example: go-linear comment delete <uuid>

Related: comment_list, comment_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runDelete(cmd, client, args[0], confirmFlags)
		},
	}

	confirmFlags.Bind(cmd)
	return cmd
}

func runDelete(cmd *cobra.Command, client *linear.Client, commentID string, confirmFlags *cli.ConfirmationFlags) error {
	ctx := cmd.Context()

	// Confirmation prompt unless --yes
	if !confirmFlags.Yes {
		fmt.Fprintf(cmd.OutOrStderr(), "Are you sure you want to delete comment %s? This cannot be undone.\n", commentID)
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

	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
		"success":   true,
		"commentId": commentID,
	}, true)
}
