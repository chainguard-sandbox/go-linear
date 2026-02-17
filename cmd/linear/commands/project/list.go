package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	projectfilter "github.com/chainguard-sandbox/go-linear/internal/filter/project"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the project list command.
func NewListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	fieldFlags := &cli.FieldFlags{}
	paginationFlags := &cli.PaginationFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects with filtering",
		Long: `List projects with filtering. Returns 4 default fields per project.

Filters: --name, --creator, --lead, --health, --slug-id, --priority
Date filters: --created-after, --completed-after, --started-after, --target-after, etc. (date formats: see issue_list)
Relation filters: --has-blocked-by, --has-blocking, --has-related

Example: go-linear project list --health=onTrack

Related: project_get, project_create, project_milestone-create`,
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
	cmd.Flags().String("updated-after", "", "Updated after date")
	cmd.Flags().String("updated-before", "", "Updated before date")
	cmd.Flags().String("completed-after", "", "Completed after date")
	cmd.Flags().String("completed-before", "", "Completed before date")
	cmd.Flags().String("canceled-after", "", "Canceled after date")
	cmd.Flags().String("canceled-before", "", "Canceled before date")
	cmd.Flags().String("started-after", "", "Started after date")
	cmd.Flags().String("started-before", "", "Started before date")
	cmd.Flags().String("target-after", "", "Target date after")
	cmd.Flags().String("target-before", "", "Target date before")

	// Entity filters
	cmd.Flags().String("id", "", "Project UUID")
	cmd.Flags().String("creator", "", "Creator name, email, or 'me'")
	cmd.Flags().String("lead", "", "Lead name, email, or 'me'")

	// Text filters
	cmd.Flags().String("name", "", "Name contains (case-insensitive)")
	cmd.Flags().String("slug-id", "", "Project slug ID (exact match)")

	// State filters
	cmd.Flags().String("health", "", "Health: onTrack, atRisk, offTrack")
	cmd.Flags().Int("priority", -1, "Priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low")

	// Relation filters
	cmd.Flags().Bool("has-blocked-by", false, "Has blocked-by relations")
	cmd.Flags().Bool("has-blocking", false, "Has blocking relations")
	cmd.Flags().Bool("has-related", false, "Has related relations")

	// Output
	paginationFlags.Bind(cmd, 50)
	fieldFlags.Bind(cmd, "defaults (...) | none | defaults,extra")
	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client, fieldFlags *cli.FieldFlags, paginationFlags *cli.PaginationFlags) error {
	ctx := cmd.Context()

	res := resolver.New(client)

	// Build filter from flags
	filterBuilder := projectfilter.NewFilterBuilder(res)
	if err := filterBuilder.FromFlags(ctx, cmd); err != nil {
		return err
	}
	projFilter := filterBuilder.Build()

	first := paginationFlags.LimitPtr()

	// Use filtered or unfiltered query based on whether filters were set
	if projFilter != nil {
		projects, err := client.ProjectsFiltered(ctx, first, nil, projFilter)
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("project.list", configOverrides)
		fieldSelector, err := fieldfilter.NewForList(fieldFlags.Fields, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), projects, true, fieldSelector)
	}

	// No filters: use regular query
	projects, err := client.Projects(ctx, first, nil)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	cfg, _ := config.Load()
	var configOverrides map[string]string
	if cfg != nil {
		configOverrides = cfg.FieldDefaults
	}
	defaults := fieldfilter.GetDefaults("project.list", configOverrides)
	fieldSelector, err := fieldfilter.NewForList(fieldFlags.Fields, defaults)
	if err != nil {
		return fmt.Errorf("invalid --fields: %w", err)
	}
	return formatter.FormatJSONFiltered(cmd.OutOrStdout(), projects, true, fieldSelector)
}
