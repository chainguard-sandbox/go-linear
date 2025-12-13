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
		Long: `⚠️ Delete project milestone. Cannot be undone. Prompts unless --yes.

Example: go-linear project milestone-delete <uuid>

Related: project_get, project_milestone-create`,
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
