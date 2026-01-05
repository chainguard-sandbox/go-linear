package cycle

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the cycle get command.
func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single cycle by ID",
		Long: `Get cycle by UUID. Returns 6 default fields.

Example: go-linear cycle get <uuid> --output=json

Related: cycle_list, cycle_update`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runGet(cmd, client, args[0])
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,name,startsAt,endsAt,createdAt,description) | none | defaults,extra")

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, cycleID string) error {
	ctx := cmd.Context()

	cycle, err := client.Cycle(ctx, cycleID)
	if err != nil {
		return fmt.Errorf("failed to get cycle: %w", err)
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
		defaults := fieldfilter.GetDefaults("cycle.get", configOverrides)
		fieldSelector, err := fieldfilter.New(fieldsSpec, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), cycle, true, fieldSelector)
	case "table":
		if cycle.Name != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Name:   %s\n", *cycle.Name)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Number: %.0f\n", cycle.Number)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
