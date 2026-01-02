package project

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	projectfilter "github.com/chainguard-sandbox/go-linear/internal/filter/project"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the project list command.
func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects with filtering",
		Long: `List projects with filtering. Returns 4 default fields per project.

Filters: --name, --creator, --lead, --health, --slug-id, --priority
Date filters: --created-after, --completed-after, --started-after, --target-after, etc. (date formats: see issue_list)
Relation filters: --has-blocked-by, --has-blocking, --has-related

Example: go-linear project list --health=onTrack --output=json

Related: project_get, project_create, project_milestone-create`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client)
		},
	}

	// Pagination
	cmd.Flags().IntP("limit", "l", 50, "Number of projects to return")

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
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,name,description,createdAt) | none | defaults,extra")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()
	res := resolver.New(client)

	// Build filter from flags
	filterBuilder := projectfilter.NewFilterBuilder(res)
	if err := filterBuilder.FromFlags(ctx, cmd); err != nil {
		return err
	}
	projFilter := filterBuilder.Build()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	// Use filtered or unfiltered query based on whether filters were set
	if projFilter != nil {
		projects, err := client.ProjectsFiltered(ctx, &first, nil, projFilter)
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}

		switch output {
		case "json":
			cfg, _ := config.Load()
			var configOverrides map[string]string
			if cfg != nil {
				configOverrides = cfg.FieldDefaults
			}
			defaults := fieldfilter.GetDefaults("project.list", configOverrides)
			fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
			if err != nil {
				return fmt.Errorf("invalid --fields: %w", err)
			}
			return formatter.FormatJSONFiltered(cmd.OutOrStdout(), projects, true, fieldSelector)
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

	// No filters: use regular query
	projects, err := client.Projects(ctx, &first, nil)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	switch output {
	case "json":
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("project.list", configOverrides)
		fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), projects, true, fieldSelector)
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
