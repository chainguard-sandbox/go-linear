package project

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewMilestoneDeleteCommand creates the project milestone-delete command.
func NewMilestoneDeleteCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "milestone-delete <milestone-id>",
		Short: "Delete a project milestone",
		Long: `Delete a project milestone permanently.

⚠️ Warning: Destructive operation - cannot be undone

This permanently removes the milestone from the project.
Issues associated with the milestone are not deleted, only the milestone itself.

Confirmation prompt appears unless --yes flag is used.

Parameters:
  <milestone-id>: Milestone UUID to delete (required)

Examples:
  # Delete milestone with confirmation
  linear project milestone-delete <milestone-uuid>

  # Delete without confirmation (use with caution)
  linear project milestone-delete <milestone-uuid> --yes

TIP: Use 'linear project get <project-id> --output=json' to see milestones and their IDs

Common Errors:
  - "milestone not found": Milestone ID may be invalid or already deleted
  - "permission denied": Check API key has write permissions

Related Commands:
  - linear project milestone-create - Create a new milestone
  - linear project milestone-update - Update milestone details
  - linear project get - View project's milestones`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runMilestoneDelete(cmd, client, args[0])
		},
	}

	cmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	return cmd
}

func runMilestoneDelete(cmd *cobra.Command, client *linear.Client, milestoneID string) error {
	ctx := context.Background()

	// Confirmation prompt
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		fmt.Fprintf(cmd.OutOrStderr(), "⚠️  Are you sure you want to delete this milestone? This cannot be undone.\n")
		fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "yes" {
			fmt.Fprintln(cmd.OutOrStderr(), "Canceled")
			return nil
		}
	}

	// Delete milestone
	err := client.ProjectMilestoneDelete(ctx, milestoneID)
	if err != nil {
		return fmt.Errorf("failed to delete milestone: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Deleted milestone\n")
	return nil
}
