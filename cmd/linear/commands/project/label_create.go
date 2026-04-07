package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewLabelCreateCommand creates the project label-create command.
func NewLabelCreateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "label-create",
		Short: "Create a project label",
		Long: `Create a new project label. Safe operation.

Required: --name
Optional: --color (hex color), --description, --parent-id (parent label UUID)

Example: go-linear project label-create --name="Backend" --color="#ff0000"

Related: project_label-list, project_label-update, project_label-delete`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runLabelCreate(cmd, client)
		},
	}

	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().String("name", "", "Label name (required)")
	cmd.Flags().String("color", "", "Label color as hex string (e.g. #ff0000)")
	cmd.Flags().String("description", "", "Label description")
	cmd.Flags().String("parent-id", "", "Parent label ID for grouping")

	return cmd
}

func runLabelCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	name, _ := cmd.Flags().GetString("name")

	input := intgraphql.ProjectLabelCreateInput{
		Name: name,
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

	result, err := client.ProjectLabelCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create project label: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
}
