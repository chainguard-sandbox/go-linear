package project

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewRelationDeleteCommand creates the project relation-delete command.
func NewRelationDeleteCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relation-delete <relation-id>",
		Short: "Delete a project relation",
		Long: `⚠️ Delete a project relation. Cannot be undone. Prompts unless --yes.

Example: go-linear project relation-delete <uuid>

Related: project_relation-list, project_relation-create`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runRelationDelete(cmd, client, args[0])
		},
	}

	cmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	return cmd
}

func runRelationDelete(cmd *cobra.Command, client *linear.Client, relationID string) error {
	ctx := cmd.Context()

	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		fmt.Fprintf(cmd.OutOrStderr(), "⚠️  Are you sure you want to delete this project relation? This cannot be undone.\n")
		fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)

		if !strings.EqualFold(response, "yes") {
			fmt.Fprintln(cmd.OutOrStderr(), "Canceled")
			return nil
		}
	}

	err := client.ProjectRelationDelete(ctx, relationID)
	if err != nil {
		return fmt.Errorf("failed to delete project relation: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Deleted project relation\n")
	return nil
}
