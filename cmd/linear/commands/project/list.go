package project

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the project list command.
func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		Long: `List all projects in the Linear workspace.

Use this to browse active projects, discover project names for milestones, or track project progress.

Output (--output=json):
  Returns JSON with:
  - nodes: Array of projects
  - pageInfo: {hasNextPage: bool, endCursor: string}

  Each project contains:
  - id: Project UUID
  - name: Project name
  - description: Project description
  - state: Project state (planned, started, completed, etc.)
  - progress: Completion percentage (0-100)
  - lead: Project lead user reference

Examples:
  # List all projects
  linear project list

  # List with limit
  linear project list --limit=20

  # JSON output for parsing
  linear project list --output=json

TIP: Use project names when creating milestones

Related Commands:
  - linear project get - Get single project details with milestones
  - linear project create - Create new project
  - linear project milestone-create - Add milestone to project`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client)
		},
	}

	cmd.Flags().IntP("limit", "l", 50, "Number of projects to return")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	projects, err := client.Projects(ctx, &first, nil)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), projects, true)
	case "table":
		if len(projects.Nodes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No projects found")
			return nil
		}
		for _, proj := range projects.Nodes {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", proj.Name)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
