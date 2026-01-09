package initiative

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
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

Example: go-linear initiative add-project --initiative=<uuid> --project=<uuid> --output=json

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
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

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

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Linked project %s to initiative %s\n",
			result.Project.Name, result.Initiative.Name)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
