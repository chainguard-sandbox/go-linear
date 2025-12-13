package cycle

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the cycle list command.
func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all cycles",
		Long: `List cycles (sprints). Returns 5 default fields per cycle.

Example: go-linear-cli cycle list --output=json

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

	cmd.Flags().IntP("limit", "l", 50, "Number of cycles to return")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,name,startsAt,endsAt,createdAt) | none | defaults,extra")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	cycles, err := client.Cycles(ctx, &first, nil)
	if err != nil {
		return fmt.Errorf("failed to list cycles: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

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
			fmt.Fprintf(cmd.OutOrStdout(), "%s (%.0f - %.0f)\n", name, cycle.Number, cycle.Number)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
