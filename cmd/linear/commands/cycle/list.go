package cycle

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	cyclefilter "github.com/chainguard-sandbox/go-linear/internal/filter/cycle"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the cycle list command.
func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cycles with filtering",
		Long: `List cycles (sprints) with filtering. Returns 5 default fields per cycle.

Filters: --team (name/key), --active, --past, --future, --name
Date filters: --created-after, --starts-after, --ends-before, etc. (date formats: see issue_list)
State filters: --active, --past, --future, --next, --previous, --in-cooldown

Example: go-linear cycle list --active --team=ENG --output=json

Related: cycle_get, cycle_create, issue_list`,
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
	cmd.Flags().IntP("limit", "l", 50, "Number of cycles to return")

	// Date filters
	cmd.Flags().String("created-after", "", "Created after date (ISO8601, 'yesterday', '7d')")
	cmd.Flags().String("created-before", "", "Created before date")
	cmd.Flags().String("updated-after", "", "Updated after date")
	cmd.Flags().String("updated-before", "", "Updated before date")
	cmd.Flags().String("completed-after", "", "Completed after date")
	cmd.Flags().String("completed-before", "", "Completed before date")
	cmd.Flags().String("starts-after", "", "Starts after date")
	cmd.Flags().String("starts-before", "", "Starts before date")
	cmd.Flags().String("ends-after", "", "Ends after date")
	cmd.Flags().String("ends-before", "", "Ends before date")

	// Boolean state filters
	cmd.Flags().Bool("active", false, "Show only active cycles")
	cmd.Flags().Bool("future", false, "Show only future cycles")
	cmd.Flags().Bool("past", false, "Show only past cycles")
	cmd.Flags().Bool("next", false, "Show only next cycle")
	cmd.Flags().Bool("previous", false, "Show only previous cycle")
	cmd.Flags().Bool("in-cooldown", false, "Show only cycles in cooldown")

	// Entity filters
	cmd.Flags().String("id", "", "Cycle UUID")
	cmd.Flags().String("name", "", "Cycle name (contains, case-insensitive)")
	cmd.Flags().Int("number", -1, "Cycle number")
	cmd.Flags().String("team", "", "Team name or ID")

	// Output
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,name,startsAt,endsAt,createdAt) | none | defaults,extra")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	// Build filter from flags
	filterBuilder := cyclefilter.NewFilterBuilder(res)
	if err := filterBuilder.FromFlags(ctx, cmd); err != nil {
		return err
	}
	cycleFilter := filterBuilder.Build()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	// Use filtered or unfiltered query based on whether filters were set
	if cycleFilter != nil {
		cycles, err := client.CyclesFiltered(ctx, &first, nil, cycleFilter)
		if err != nil {
			return fmt.Errorf("failed to list cycles: %w", err)
		}

		switch output {
		case "json":
			cfg, _ := config.Load()
			var configOverrides map[string]string
			if cfg != nil {
				configOverrides = cfg.FieldDefaults
			}
			defaults := fieldfilter.GetDefaults("cycle.list", configOverrides)
			fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
			if err != nil {
				return fmt.Errorf("invalid --fields: %w", err)
			}
			return formatter.FormatJSONFiltered(cmd.OutOrStdout(), cycles, true, fieldSelector)
		case "table":
			if len(cycles.Nodes) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No cycles found")
				return nil
			}
			for _, cycle := range cycles.Nodes {
				name := "Unnamed"
				if cycle.Name != nil {
					name = *cycle.Name
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s (%.0f)\n", name, cycle.Number)
			}
			return nil
		default:
			return fmt.Errorf("unsupported output format: %s", output)
		}
	}

	// No filters: use regular query
	cycles, err := client.Cycles(ctx, &first, nil)
	if err != nil {
		return fmt.Errorf("failed to list cycles: %w", err)
	}

	switch output {
	case "json":
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("cycle.list", configOverrides)
		fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), cycles, true, fieldSelector)
	case "table":
		if len(cycles.Nodes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No cycles found")
			return nil
		}
		for _, cycle := range cycles.Nodes {
			name := "Unnamed"
			if cycle.Name != nil {
				name = *cycle.Name
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s (%.0f)\n", name, cycle.Number)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
