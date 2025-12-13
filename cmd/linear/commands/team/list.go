package team

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the team list command.
func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all teams",
		Long: `List teams. Returns 6 default fields per team. Use for discovering team names/keys.

Team keys appear in issue identifiers (e.g., ENG-123 where ENG is the team key).

Example: go-linear team list --output=json

Returns: {nodes: [{6 team fields}...], pageInfo: {hasNextPage, endCursor}}
Related: team_get, team_members, issue_list`,
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
	cmd.Flags().String("fields", "", "defaults (id,name,key,description,icon,createdAt) | none | defaults,extra | id,name,...")

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
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	switch output {
	case "json":
		// Load config for field defaults
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}

		// Get command defaults
		defaults := fieldfilter.GetDefaults("team.list", configOverrides)

		// Parse field selector with defaults (list command preserves nodes/pageInfo)
		fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}

		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), teams, true, fieldSelector)
	case "table":
		return formatter.FormatTeamsTable(cmd.OutOrStdout(), teams.Nodes)
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
