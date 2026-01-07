package team

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	teamfilter "github.com/chainguard-sandbox/go-linear/internal/filter/team"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the team list command.
func NewListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	outputFlags := &cli.OutputFlags{}
	paginationFlags := &cli.PaginationFlags{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all teams",
		Long: `List teams with filtering. Returns 7 default fields per team (includes issueCount). Use for discovering team names/keys.

Team keys appear in issue identifiers (e.g., ENG-123 where ENG is the team key).

Filters: --name, --key, --description, --private
Date filters: --created-after, --created-before, --updated-after, --updated-before

Example: go-linear team list --private=true --output=json

Returns: {nodes: [{7 team fields}...], pageInfo: {hasNextPage, endCursor}}
Related: team_get, team_members, issue_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client, outputFlags, paginationFlags)
		},
	}

	// Bind common flags
	paginationFlags.Bind(cmd, 100)

	// Date filters
	cmd.Flags().String("created-after", "", "Created after date (ISO8601, 'yesterday', '7d')")
	cmd.Flags().String("created-before", "", "Created before date")
	cmd.Flags().String("updated-after", "", "Updated after date")
	cmd.Flags().String("updated-before", "", "Updated before date")

	// Entity filters
	cmd.Flags().String("id", "", "Team UUID")

	// Text filters
	cmd.Flags().String("name", "", "Name contains (case-insensitive)")
	cmd.Flags().String("key", "", "Team key (exact match, case-insensitive)")
	cmd.Flags().String("description", "", "Description contains (case-insensitive)")

	// Boolean filters
	cmd.Flags().Bool("private", false, "Filter by private status")

	// Output flags
	outputFlags.Bind(cmd, "defaults (id,name,key,description,icon,createdAt) | none | defaults,extra | id,name,...")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client, outputFlags *cli.OutputFlags, paginationFlags *cli.PaginationFlags) error {
	ctx := cmd.Context()

	// Validate flags
	if err := outputFlags.Validate(); err != nil {
		return err
	}

	// Build filter from flags
	filterBuilder := teamfilter.NewFilterBuilder(nil)
	if err := filterBuilder.FromFlags(ctx, cmd); err != nil {
		return err
	}
	tmFilter := filterBuilder.Build()

	first := paginationFlags.LimitPtr()

	// Use filtered or unfiltered query based on whether filters were set
	if tmFilter != nil {
		teams, err := client.TeamsFiltered(ctx, first, nil, tmFilter)
		if err != nil {
			return fmt.Errorf("failed to list teams: %w", err)
		}

		switch outputFlags.Output {
		case "json":
			cfg, _ := config.Load()
			var configOverrides map[string]string
			if cfg != nil {
				configOverrides = cfg.FieldDefaults
			}
			defaults := fieldfilter.GetDefaults("team.list", configOverrides)
			fieldSelector, err := fieldfilter.NewForList(outputFlags.Fields, defaults)
			if err != nil {
				return fmt.Errorf("invalid --fields: %w", err)
			}
			return formatter.FormatJSONFiltered(cmd.OutOrStdout(), teams, true, fieldSelector)
		case "table":
			return formatter.FormatTeamsTableFiltered(cmd.OutOrStdout(), teams.Nodes)
		default:
			return fmt.Errorf("unsupported output format: %s", outputFlags.Output)
		}
	}

	// No filters: use regular query
	teams, err := client.Teams(ctx, first, nil)
	if err != nil {
		return fmt.Errorf("failed to list teams: %w", err)
	}

	switch outputFlags.Output {
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
		fieldSelector, err := fieldfilter.NewForList(outputFlags.Fields, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}

		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), teams, true, fieldSelector)
	case "table":
		return formatter.FormatTeamsTable(cmd.OutOrStdout(), teams.Nodes)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFlags.Output)
	}
}
