package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewLabelUpdateCommand creates the project label-update command.
func NewLabelUpdateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "label-update <label-id>",
		Short: "Update a project label",
		Long: `Update a project label. Modifies existing data.

Fields: --name, --color, --description, --parent-id

Example: go-linear project label-update <uuid> --name="Infrastructure" --color="#00ff00"

Related: project_label-create, project_label-list, project_label-delete`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runLabelUpdate(cmd, client, args[0])
		},
	}

	cmd.Flags().String("name", "", "New label name")
	cmd.Flags().String("color", "", "New label color as hex string")
	cmd.Flags().String("description", "", "New label description")
	cmd.Flags().String("parent-id", "", "New parent label ID")

	return cmd
}

func runLabelUpdate(cmd *cobra.Command, client *linear.Client, labelID string) error {
	ctx := cmd.Context()

	input := intgraphql.ProjectLabelUpdateInput{}

	if name, _ := cmd.Flags().GetString("name"); name != "" {
		input.Name = &name
	}

	if color, _ := cmd.Flags().GetString("color"); color != "" {
		input.Color = &color
	}

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		input.Description = &desc
	}

	if parentID, _ := cmd.Flags().GetString("parent-id"); parentID != "" {
		input.ParentID = &parentID
	}

	result, err := client.ProjectLabelUpdate(ctx, labelID, input)
	if err != nil {
		return fmt.Errorf("failed to update project label: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
}
