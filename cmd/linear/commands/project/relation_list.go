package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewRelationListCommand creates the project relation-list command.
func NewRelationListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relation-list",
		Short: "List project relations",
		Long: `List all project relations in the organization.

Example: go-linear project relation-list

Related: project_relation-create, project_relation-update, project_relation-delete`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runRelationList(cmd, client)
		},
	}

	return cmd
}

func runRelationList(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	relations, err := client.ProjectRelations(ctx, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to list project relations: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), relations.Nodes, true)
}
