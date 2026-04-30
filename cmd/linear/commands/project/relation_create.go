package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewRelationCreateCommand creates the project relation-create command.
func NewRelationCreateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relation-create",
		Short: "Create a project relation",
		Long: `Create a relationship between two projects. Safe operation.

Required: --project, --related-project, --type, --anchor-type, --related-anchor-type
Optional: --project-milestone, --related-project-milestone

Relation types: blocks, dependsOn, related
Anchor types: milestone, project

Note: Linear automatically creates the inverse relation (e.g. blocks A→B creates a blockedBy view from B).

Example: go-linear project relation-create --project=<uuid> --related-project=<uuid> --type=blocks --anchor-type=project --related-anchor-type=project

Related: project_relation-update, project_relation-delete, project_relation-list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runRelationCreate(cmd, client)
		},
	}

	cmd.Flags().String("project", "", "Project name or ID (required)")
	cmd.Flags().String("related-project", "", "Related project name or ID (required)")
	cmd.Flags().String("type", "", "Relation type: blocks, dependsOn, related (required)")
	cmd.Flags().String("anchor-type", "", "Anchor type for project end: project, milestone (required)")
	cmd.Flags().String("related-anchor-type", "", "Anchor type for related project end: project, milestone (required)")
	cmd.Flags().String("project-milestone", "", "Project milestone ID (optional)")
	cmd.Flags().String("related-project-milestone", "", "Related project milestone ID (optional)")

	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("related-project")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("anchor-type")
	_ = cmd.MarkFlagRequired("related-anchor-type")

	return cmd
}

func runRelationCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	projectName, _ := cmd.Flags().GetString("project")
	projectID, err := res.ResolveProject(ctx, projectName)
	if err != nil {
		return fmt.Errorf("failed to resolve project: %w", err)
	}

	relatedProjectName, _ := cmd.Flags().GetString("related-project")
	relatedProjectID, err := res.ResolveProject(ctx, relatedProjectName)
	if err != nil {
		return fmt.Errorf("failed to resolve related project: %w", err)
	}

	relationType, _ := cmd.Flags().GetString("type")
	if !validRelationTypes[relationType] {
		return fmt.Errorf("invalid --type %q: must be one of blocks, dependsOn, related", relationType)
	}

	anchorType, _ := cmd.Flags().GetString("anchor-type")
	if !validAnchorTypes[anchorType] {
		return fmt.Errorf("invalid --anchor-type %q: must be one of project, milestone", anchorType)
	}

	relatedAnchorType, _ := cmd.Flags().GetString("related-anchor-type")
	if !validAnchorTypes[relatedAnchorType] {
		return fmt.Errorf("invalid --related-anchor-type %q: must be one of project, milestone", relatedAnchorType)
	}

	input := intgraphql.ProjectRelationCreateInput{
		ProjectID:         projectID,
		RelatedProjectID:  relatedProjectID,
		Type:              relationType,
		AnchorType:        anchorType,
		RelatedAnchorType: relatedAnchorType,
	}

	if milestoneID, _ := cmd.Flags().GetString("project-milestone"); milestoneID != "" {
		input.ProjectMilestoneID = &milestoneID
	}

	if relMilestoneID, _ := cmd.Flags().GetString("related-project-milestone"); relMilestoneID != "" {
		input.RelatedProjectMilestoneID = &relMilestoneID
	}

	result, err := client.ProjectRelationCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create project relation: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
}
