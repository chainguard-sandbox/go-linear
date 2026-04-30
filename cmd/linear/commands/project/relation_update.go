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

Fields: --type, --anchor-type, --related-anchor-type, --project-id, --related-project-id, --project-milestone-id, --related-project-milestone-id

Relation types: blocks, dependsOn, related
Anchor types: project, milestone

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
	cmd.Flags().String("anchor-type", "", "New anchor type for project end: project, milestone")
	cmd.Flags().String("related-anchor-type", "", "New anchor type for related project end: project, milestone")
	cmd.Flags().String("project-id", "", "New project ID")
	cmd.Flags().String("related-project-id", "", "New related project ID")
	cmd.Flags().String("project-milestone-id", "", "New project milestone ID")
	cmd.Flags().String("related-project-milestone-id", "", "New related project milestone ID")

	return cmd
}

var validRelationTypes = map[string]bool{"blocks": true, "dependsOn": true, "related": true}
var validAnchorTypes = map[string]bool{"project": true, "milestone": true}

func runRelationUpdate(cmd *cobra.Command, client *linear.Client, relationID string) error {
	ctx := cmd.Context()

	input := intgraphql.ProjectRelationUpdateInput{}

	if cmd.Flags().Changed("type") {
		relationType, _ := cmd.Flags().GetString("type")
		if !validRelationTypes[relationType] {
			return fmt.Errorf("invalid --type %q: must be one of blocks, dependsOn, related", relationType)
		}
		input.Type = &relationType
	}

	if cmd.Flags().Changed("anchor-type") {
		anchorType, _ := cmd.Flags().GetString("anchor-type")
		if !validAnchorTypes[anchorType] {
			return fmt.Errorf("invalid --anchor-type %q: must be one of project, milestone", anchorType)
		}
		input.AnchorType = &anchorType
	}

	if cmd.Flags().Changed("related-anchor-type") {
		relatedAnchorType, _ := cmd.Flags().GetString("related-anchor-type")
		if !validAnchorTypes[relatedAnchorType] {
			return fmt.Errorf("invalid --related-anchor-type %q: must be one of project, milestone", relatedAnchorType)
		}
		input.RelatedAnchorType = &relatedAnchorType
	}

	if cmd.Flags().Changed("project-id") {
		projectID, _ := cmd.Flags().GetString("project-id")
		input.ProjectID = &projectID
	}

	if cmd.Flags().Changed("related-project-id") {
		relatedProjectID, _ := cmd.Flags().GetString("related-project-id")
		input.RelatedProjectID = &relatedProjectID
	}

	if cmd.Flags().Changed("project-milestone-id") {
		projectMilestoneID, _ := cmd.Flags().GetString("project-milestone-id")
		input.ProjectMilestoneID = &projectMilestoneID
	}

	if cmd.Flags().Changed("related-project-milestone-id") {
		relatedProjectMilestoneID, _ := cmd.Flags().GetString("related-project-milestone-id")
		input.RelatedProjectMilestoneID = &relatedProjectMilestoneID
	}

	result, err := client.ProjectRelationUpdate(ctx, relationID, input)
	if err != nil {
		return fmt.Errorf("failed to update project relation: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
}
