package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewStatusListCommand creates the project status-list command.
func NewStatusListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status-list",
		Short: "List project statuses",
		Long: `List organization's project statuses. Returns name and type per status.

Example: go-linear project status-list

Related: project_update, project_get`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runStatusList(cmd, client)
		},
	}

	return cmd
}

func runStatusList(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	statuses, err := client.ProjectStatuses(ctx)
	if err != nil {
		return fmt.Errorf("failed to list project statuses: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), statuses, true)
}
