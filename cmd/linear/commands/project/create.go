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

// NewCreateCommand creates the project create command.
func NewCreateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		Long: `Create project. Safe operation.

Required: --name, --team (from team_list)
Optional: --description

Example: go-linear project create --name="Q1 Platform" --team=ENG --description="Platform improvements" --output=json

Related: project_list, project_get, team_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runCreate(cmd, client)
		},
	}

	cmd.Flags().String("name", "", "Project name (required)")
	_ = cmd.MarkFlagRequired("name")

	cmd.Flags().String("team", "", "Team name or ID (required)")
	_ = cmd.MarkFlagRequired("team")

	cmd.Flags().String("description", "", "Project description")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	// Resolve team to UUID
	res := resolver.New(client)
	team, _ := cmd.Flags().GetString("team")
	teamID, err := res.ResolveTeam(ctx, team)
	if err != nil {
		return fmt.Errorf("failed to resolve team: %w", err)
	}

	name, _ := cmd.Flags().GetString("name")
	input := intgraphql.ProjectCreateInput{
		Name:    name,
		TeamIds: []string{teamID},
	}

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		input.Description = &desc
	}

	result, err := client.ProjectCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Created project: %s\n", result.Name)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
