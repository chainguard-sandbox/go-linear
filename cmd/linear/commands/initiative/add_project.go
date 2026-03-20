package initiative

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewAddProjectCommand creates the initiative add-project command.
func NewAddProjectCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-project",
		Short: "Link a project to an initiative",
		Long: `Link project to initiative. Safe operation.

Creates an InitiativeToProject association. Projects can be linked to multiple
initiatives, and initiatives can have multiple projects.

Required: --initiative (UUID or name), --project (UUID or name)

Example: go-linear initiative add-project --initiative=<uuid> --project=<uuid>

Related: initiative_remove-project, project_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runAddProject(cmd, client)
		},
	}

	cmd.Flags().String("initiative", "", "Initiative name or UUID (required)")
	_ = cmd.MarkFlagRequired("initiative")
	cmd.Flags().String("project", "", "Project name or UUID (required)")
	_ = cmd.MarkFlagRequired("project")

	return cmd
}

func runAddProject(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	// Resolve initiative
	initiativeInput, _ := cmd.Flags().GetString("initiative")
	initiativeID, err := res.ResolveInitiative(ctx, initiativeInput)
	if err != nil {
		return fmt.Errorf("failed to resolve initiative: %w", err)
	}

	// Resolve project
	projectInput, _ := cmd.Flags().GetString("project")
	projectID, err := res.ResolveProject(ctx, projectInput)
	if err != nil {
		return fmt.Errorf("failed to resolve project: %w", err)
	}

	input := intgraphql.InitiativeToProjectCreateInput{
		InitiativeID: initiativeID,
		ProjectID:    projectID,
	}

	result, err := client.InitiativeToProjectCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to link project to initiative: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
}
