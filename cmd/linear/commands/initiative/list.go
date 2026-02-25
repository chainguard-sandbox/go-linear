package initiative

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	initfilter "github.com/chainguard-sandbox/go-linear/internal/filter/initiative"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

func NewListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	fieldFlags := &cli.FieldFlags{}
	paginationFlags := &cli.PaginationFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List initiatives with filtering",
		Long: `List initiatives with filtering. Returns 4 default fields per initiative.

Filters: --name, --creator, --owner, --parent, --health, --status, --slug-id
Date filters: --created-after, --target-after, etc. (date formats: see issue_list)

Example: go-linear initiative list --status=Active

Related: initiative_get, project_list, issue_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client, fieldFlags, paginationFlags)
		},
	}

	// Pagination

	// Date filters
	cmd.Flags().String("created-after", "", "Created after date (ISO8601, 'yesterday', '7d')")
	cmd.Flags().String("created-before", "", "Created before date")
	cmd.Flags().String("target-after", "", "Target date after")
	cmd.Flags().String("target-before", "", "Target date before")

	// Entity filters
	cmd.Flags().String("id", "", "Initiative UUID")
	cmd.Flags().String("creator", "", "Creator name, email, or 'me'")
	cmd.Flags().String("owner", "", "Owner name, email, or 'me'")
	cmd.Flags().String("parent", "", "Parent initiative UUID or name")

	// Text filters
	cmd.Flags().String("name", "", "Name contains (case-insensitive)")
	cmd.Flags().String("slug-id", "", "Initiative slug ID (exact match)")

	// State filters
	cmd.Flags().String("health", "", "Health: onTrack, atRisk, offTrack")
	cmd.Flags().String("status", "", "Status: Planned, Active, Completed")

	// Output
	paginationFlags.Bind(cmd, 50)
	fieldFlags.Bind(cmd, "defaults (...) | none | defaults,extra")
	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client, fieldFlags *cli.FieldFlags, paginationFlags *cli.PaginationFlags) error {
	ctx := cmd.Context()

	res := resolver.New(client)

	// Build filter from flags
	filterBuilder := initfilter.NewFilterBuilder(res)
	if err := filterBuilder.FromFlags(ctx, cmd); err != nil {
		return err
	}
	initFilter := filterBuilder.Build()

	first := paginationFlags.LimitPtr()

	// Use filtered or unfiltered query based on whether filters were set
	if initFilter != nil {
		initiatives, err := client.InitiativesFiltered(ctx, first, nil, initFilter)
		if err != nil {
			return fmt.Errorf("failed to list initiatives: %w", err)
		}

		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("initiative.list", configOverrides)
		fieldSelector, err := fieldfilter.NewForList(fieldFlags.Fields, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), initiatives, true, fieldSelector)
	}

	// No filters: use regular query
	initiatives, err := client.Initiatives(ctx, first, nil)
	if err != nil {
		return fmt.Errorf("failed to list initiatives: %w", err)
	}

	cfg, _ := config.Load()
	var configOverrides map[string]string
	if cfg != nil {
		configOverrides = cfg.FieldDefaults
	}
	defaults := fieldfilter.GetDefaults("initiative.list", configOverrides)
	fieldSelector, err := fieldfilter.NewForList(fieldFlags.Fields, defaults)
	if err != nil {
		return fmt.Errorf("invalid --fields: %w", err)
	}
	return formatter.FormatJSONFiltered(cmd.OutOrStdout(), initiatives, true, fieldSelector)
}
