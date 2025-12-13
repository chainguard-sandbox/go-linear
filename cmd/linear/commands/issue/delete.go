package issue

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewDeleteCommand creates the issue delete command.
func NewDeleteCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an issue permanently",
		Long: `⚠️ Delete issue. Cannot be undone. Prompts for confirmation unless --yes.

Example: go-linear-cli issue delete ENG-123

Related: issue_list, issue_get`,
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

func runDelete(cmd *cobra.Command, client *linear.Client, issueID string) error {
	ctx := context.Background()

	// Confirmation prompt unless --yes
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		fmt.Fprintf(cmd.OutOrStderr(), "⚠️  Are you sure you want to delete issue %s? This cannot be undone.\n", issueID)
		fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "yes" {
			fmt.Fprintln(cmd.OutOrStderr(), "Canceled.")
			return nil
		}
	}

	// Delete issue
	err := client.IssueDelete(ctx, issueID)
	if err != nil {
		return fmt.Errorf("failed to delete issue: %w", err)
	}

	// Format output
	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
			"success": true,
			"issueId": issueID,
		}, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Issue %s deleted successfully\n", issueID)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
