package issue

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUnrelateCommand creates the issue unrelate command.
func NewUnrelateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unrelate <relation-id>",
		Short: "Delete an issue relationship",
		Long: `Delete a relationship between two issues.

⚠️ Warning: Destructive operation - cannot be undone

This permanently removes the relationship between two issues.
The issues themselves are not deleted, only the relationship link.

Confirmation prompt appears unless --yes flag is used.

Parameters:
  <relation-id>: IssueRelation UUID to delete (required)

Examples:
  # Delete relation with confirmation
  linear issue unrelate <relation-uuid>

  # Delete without confirmation (use with caution)
  linear issue unrelate <relation-uuid> --yes

TIP: Use 'linear issue get ENG-123 --output=json' to see relations and their IDs

Common Errors:
  - "relation not found": Relation ID may be invalid or already deleted
  - "permission denied": Check API key has write permissions

Related Commands:
  - linear issue relate - Create a new relationship
  - linear issue update-relation - Change relationship type
  - linear issue get - View issue's current relationships`,
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
	ctx := context.Background()

	// Confirmation prompt
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		fmt.Fprintf(cmd.OutOrStderr(), "⚠️  Are you sure you want to delete this issue relation? This cannot be undone.\n")
		fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "yes" {
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
