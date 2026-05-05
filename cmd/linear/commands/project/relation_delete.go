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
	confirmFlags := &cli.ConfirmationFlags{}

	cmd := &cobra.Command{
		Use:   "relation-delete <relation-id>",
		Short: "Delete a project relation",
		Long: `Delete a project relation. Cannot be undone. Prompts unless --yes.

Example: go-linear project relation-delete <uuid>

Related: project_relation-list, project_relation-create`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runRelationDelete(cmd, client, args[0], confirmFlags)
		},
	}

	confirmFlags.Bind(cmd)

	return cmd
}

func runRelationDelete(cmd *cobra.Command, client *linear.Client, relationID string, confirmFlags *cli.ConfirmationFlags) error {
	ctx := cmd.Context()

	if !confirmFlags.Yes {
		fmt.Fprintf(cmd.OutOrStderr(), "Delete project relation %s? This cannot be undone.\n", relationID)
		fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if !strings.EqualFold(strings.TrimSpace(response), "yes") {
			fmt.Fprintln(cmd.OutOrStderr(), "Canceled.")
			return nil
		}
	}

	err := client.ProjectRelationDelete(ctx, relationID)
	if err != nil {
		return fmt.Errorf("failed to delete project relation: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted project relation\n")
	return nil
}
