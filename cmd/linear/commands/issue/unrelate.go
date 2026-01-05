package issue

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUnrelateCommand creates the issue unrelate command.
func NewUnrelateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unrelate <relation-id>",
		Short: "Delete an issue relationship",
		Long: `⚠️ Delete issue relationship. Cannot be undone. Prompts unless --yes.

Example: go-linear issue unrelate <relation-uuid>

Related: issue_relate, issue_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runUnrelate(cmd, client, args[0])
		},
	}

	cmd.Flags().Bool("yes", false, "Skip confirmation prompt")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runUnrelate(cmd *cobra.Command, client *linear.Client, relationID string) error {
	ctx := cmd.Context()

	// Confirmation prompt
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		fmt.Fprintf(cmd.OutOrStderr(), "⚠️  Are you sure you want to delete this issue relation? This cannot be undone.\n")
		fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)

		if !strings.EqualFold(response, "yes") {
			fmt.Fprintln(cmd.OutOrStderr(), "Canceled")
			return nil
		}
	}

	// Delete relation
	err := client.IssueRelationDelete(ctx, relationID)
	if err != nil {
		return fmt.Errorf("failed to delete issue relation: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Deleted issue relation\n")
	return nil
}
