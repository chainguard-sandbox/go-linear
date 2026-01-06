package issue

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewDeleteCommand creates the issue delete command.
func NewDeleteCommand(clientFactory cli.ClientFactory) *cobra.Command {
	confirmFlags := &cli.ConfirmationFlags{}
	outputFlags := &cli.OutputOnlyFlags{}
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an issue permanently",
		Long: `⚠️ Delete issue. Cannot be undone. Prompts for confirmation unless --yes.

Example: go-linear issue delete ENG-123

Related: issue_list, issue_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runDelete(cmd, client, args[0], confirmFlags, outputFlags)
		},
	}

	outputFlags.Bind(cmd)
	confirmFlags.Bind(cmd)
	return cmd
}

func runDelete(cmd *cobra.Command, client *linear.Client, issueID string, confirmFlags *cli.ConfirmationFlags, outputFlags *cli.OutputOnlyFlags) error {
	ctx := cmd.Context()

	if err := outputFlags.Validate(); err != nil {
		return err
	}

	// Confirmation prompt unless --yes
	if !confirmFlags.Yes {
		fmt.Fprintf(cmd.OutOrStderr(), "⚠️  Are you sure you want to delete issue %s? This cannot be undone.\n", issueID)
		fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)

		if !strings.EqualFold(response, "yes") {
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
	switch outputFlags.Output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
			"success": true,
			"issueId": issueID,
		}, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Issue %s deleted successfully\n", issueID)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", outputFlags.Output)
	}
}
