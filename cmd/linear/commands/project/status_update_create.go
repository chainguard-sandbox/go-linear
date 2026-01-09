package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewStatusUpdateCreateCommand creates the project status-update-create command.
func NewStatusUpdateCreateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status-update-create",
		Short: "Create a status update for a project",
		Long: `Create project status update. Safe operation.

Required: --project (UUID or name), --body
Optional: --health (onTrack, atRisk, offTrack)

Example: go-linear project status-update-create --project=<uuid> --body="On track for Q1 release" --health=onTrack --output=json

Related: project_status-update-list, project_status-update-get, project_status-update-delete`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runStatusUpdateCreate(cmd, client)
		},
	}

	cmd.Flags().String("project", "", "Project name or UUID (required)")
	_ = cmd.MarkFlagRequired("project")
	cmd.Flags().String("body", "", "Status update body (markdown) (required)")
	_ = cmd.MarkFlagRequired("body")
	cmd.Flags().String("health", "", "Project health: onTrack, atRisk, offTrack")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runStatusUpdateCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	// Resolve project
	projectInput, _ := cmd.Flags().GetString("project")
	projectID, err := res.ResolveProject(ctx, projectInput)
	if err != nil {
		return fmt.Errorf("failed to resolve project: %w", err)
	}

	body, _ := cmd.Flags().GetString("body")
	input := intgraphql.ProjectUpdateCreateInput{
		ProjectID: projectID,
		Body:      &body,
	}

	if health, _ := cmd.Flags().GetString("health"); health != "" {
		healthType := intgraphql.ProjectUpdateHealthType(health)
		input.Health = &healthType
	}

	result, err := client.ProjectUpdateCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create project status update: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Created project status update\n")
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
