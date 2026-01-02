package initiative

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	initfilter "github.com/chainguard-sandbox/go-linear/internal/filter/initiative"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List initiatives with filtering",
		Long: `List initiatives with filtering. Returns 4 default fields per initiative.

Filters: --name, --creator, --owner, --health, --status, --slug-id
Date filters: --created-after, --target-after, etc. (date formats: see issue_list)

Example: go-linear initiative list --status=Active --output=json

Related: initiative_get, project_list, issue_list`,
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
	cmd.Flags().IntP("limit", "l", 50, "Number to return")

	// Date filters
	cmd.Flags().String("created-after", "", "Created after date (ISO8601, 'yesterday', '7d')")
	cmd.Flags().String("created-before", "", "Created before date")
	cmd.Flags().String("target-after", "", "Target date after")
	cmd.Flags().String("target-before", "", "Target date before")

	// Entity filters
	cmd.Flags().String("id", "", "Initiative UUID")
	cmd.Flags().String("creator", "", "Creator name, email, or 'me'")
	cmd.Flags().String("owner", "", "Owner name, email, or 'me'")

	// Text filters
	cmd.Flags().String("name", "", "Name contains (case-insensitive)")
	cmd.Flags().String("slug-id", "", "Initiative slug ID (exact match)")

	// State filters
	cmd.Flags().String("health", "", "Health: onTrack, atRisk, offTrack")
	cmd.Flags().String("status", "", "Status: Planned, Active, Completed")

	// Output
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,name,description,createdAt) | none | defaults,extra")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()
	res := resolver.New(client)

	// Build filter from flags
	filterBuilder := initfilter.NewFilterBuilder(res)
	if err := filterBuilder.FromFlags(ctx, cmd); err != nil {
		return err
	}
	initFilter := filterBuilder.Build()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	// Use filtered or unfiltered query based on whether filters were set
	if initFilter != nil {
		initiatives, err := client.InitiativesFiltered(ctx, &first, nil, initFilter)
		if err != nil {
			return fmt.Errorf("failed to list initiatives: %w", err)
		}

		switch output {
		case "json":
			cfg, _ := config.Load()
			var configOverrides map[string]string
			if cfg != nil {
				configOverrides = cfg.FieldDefaults
			}
			defaults := fieldfilter.GetDefaults("initiative.list", configOverrides)
			fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
			if err != nil {
				return fmt.Errorf("invalid --fields: %w", err)
			}
			return formatter.FormatJSONFiltered(cmd.OutOrStdout(), initiatives, true, fieldSelector)
		case "table":
			if len(initiatives.Nodes) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No initiatives found")
				return nil
			}
			for _, init := range initiatives.Nodes {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", init.Name)
			}
			return nil
		default:
			return fmt.Errorf("unsupported output format: %s", output)
		}
	}

	// No filters: use regular query
	initiatives, err := client.Initiatives(ctx, &first, nil)
	if err != nil {
		return fmt.Errorf("failed to list initiatives: %w", err)
	}

	switch output {
	case "json":
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("initiative.list", configOverrides)
		fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), initiatives, true, fieldSelector)
	case "table":
		if len(initiatives.Nodes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No initiatives found")
			return nil
		}
		for _, init := range initiatives.Nodes {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", init.Name)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
