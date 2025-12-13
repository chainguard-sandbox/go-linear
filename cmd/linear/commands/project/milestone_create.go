package project

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/dateparser"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewMilestoneCreateCommand creates the project milestone-create command.
func NewMilestoneCreateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "milestone-create",
		Short: "Create a milestone within a project",
		Long: `Create project milestone. Safe operation.

Required: --project (UUID from project_list), --name
Optional: --description, --target-date (date formats: see issue_list)

Example: go-linear-cli project milestone-create --project=<uuid> --name="Q1 2025" --target-date=2025-03-31 --output=json

Related: project_milestone-update, project_milestone-delete, project_get`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runMilestoneCreate(cmd, client)
		},
	}

	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().String("project", "", "Project name or ID (required)")
	cmd.Flags().String("name", "", "Milestone name (required)")
	cmd.Flags().String("description", "", "Milestone description (markdown)")
	cmd.Flags().String("target-date", "", "Target completion date (ISO8601 or relative)")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runMilestoneCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()
	res := resolver.New(client)

	projectName, _ := cmd.Flags().GetString("project")
	projectID, err := res.ResolveProject(ctx, projectName)
	if err != nil {
		return fmt.Errorf("failed to resolve project: %w", err)
	}

	name, _ := cmd.Flags().GetString("name")

	input := intgraphql.ProjectMilestoneCreateInput{
		ProjectID: projectID,
		Name:      name,
	}

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		input.Description = &desc
	}

	if targetDateStr, _ := cmd.Flags().GetString("target-date"); targetDateStr != "" {
		parser := dateparser.New()
		targetDate, err := parser.Parse(targetDateStr)
		if err != nil {
			return fmt.Errorf("invalid target-date: %w", err)
		}
		dateStr := targetDate.Format("2006-01-02")
		input.TargetDate = &dateStr
	}

	result, err := client.ProjectMilestoneCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create milestone: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Created milestone '%s' in project %s\n", name, projectName)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
