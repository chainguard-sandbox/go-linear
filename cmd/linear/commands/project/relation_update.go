package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewRelationUpdateCommand creates the project relation-update command.
func NewRelationUpdateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relation-update <relation-id>",
		Short: "Update a project relation",
		Long: `Update a project relation. Modifies existing data.

Fields: --type, --anchor-type, --related-anchor-type

Example: go-linear project relation-update <uuid> --type=dependsOn

Related: project_relation-create, project_relation-delete, project_relation-list`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runRelationUpdate(cmd, client, args[0])
		},
	}

	cmd.Flags().String("type", "", "New relation type: blocks, dependsOn, related")
	cmd.Flags().String("anchor-type", "", "New anchor type for project end")
	cmd.Flags().String("related-anchor-type", "", "New anchor type for related project end")

	return cmd
}

func runRelationUpdate(cmd *cobra.Command, client *linear.Client, relationID string) error {
	ctx := cmd.Context()

	input := intgraphql.ProjectRelationUpdateInput{}

	if relationType, _ := cmd.Flags().GetString("type"); relationType != "" {
		input.Type = &relationType
	}

	if anchorType, _ := cmd.Flags().GetString("anchor-type"); anchorType != "" {
		input.AnchorType = &anchorType
	}

	if relatedAnchorType, _ := cmd.Flags().GetString("related-anchor-type"); relatedAnchorType != "" {
		input.RelatedAnchorType = &relatedAnchorType
	}

	result, err := client.ProjectRelationUpdate(ctx, relationID, input)
	if err != nil {
		return fmt.Errorf("failed to update project relation: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
}
