package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewLabelListCommand creates the project label-list command.
func NewLabelListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "label-list",
		Short: "List project labels",
		Long: `List all project labels in the organization.

Example: go-linear project label-list

Related: project_label-create, project_label-update, project_label-delete`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runLabelList(cmd, client)
		},
	}

	return cmd
}

func runLabelList(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	labels, err := client.ProjectLabels(ctx, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to list project labels: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), labels.Nodes, true)
}
