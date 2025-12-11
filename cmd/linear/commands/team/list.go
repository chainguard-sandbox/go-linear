package team

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the team list command.
func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all teams",
		Long: `List all teams in the Linear workspace.

Use this to discover team names and keys for filtering issues, creating projects, or assigning work.
Essential for parameter discovery when creating or filtering issues.

Output (--output=json):
  Returns JSON with:
  - nodes: Array of teams
  - pageInfo: {hasNextPage: bool, endCursor: string}

  Each team contains:
  - id: Team UUID
  - name: Team name (e.g., "Engineering")
  - key: Team key (e.g., "ENG" - used in issue identifiers like ENG-123)
  - description: Team description
  - color: Team color hex code

Examples:
  # List all teams
  linear team list

  # JSON output for parameter discovery
  linear team list --output=json

TIP: Use team names or keys (not UUIDs) when creating or filtering issues

Related Commands:
  - linear team get - Get single team details
  - linear team members - List team members
  - linear issue list --team=<name> - Filter issues by team`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client)
		},
	}

	cmd.Flags().IntP("limit", "l", 100, "Number of teams to return")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	teams, err := client.Teams(ctx, &first, nil)
	if err != nil {
		return fmt.Errorf("failed to list teams: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), teams, true)
	case "table":
		return formatter.FormatTeamsTable(cmd.OutOrStdout(), teams.Nodes)
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
