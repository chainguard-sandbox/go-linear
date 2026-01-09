package cycle

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the cycle get command.
func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
	flags := &cli.OutputFlags{}

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

			return runGet(cmd, client, args[0], flags)
		},
	}

	flags.Bind(cmd, "defaults (id,name,startsAt,endsAt,createdAt,description) | none | defaults,extra")

	return cmd
}
func runGet(cmd *cobra.Command, client *linear.Client, cycleID string, flags *cli.OutputFlags) error {
	ctx := cmd.Context()

	if err := flags.Validate(); err != nil {
		return err
	}

	cycle, err := client.Cycle(ctx, cycleID)
	if err != nil {
		return fmt.Errorf("failed to get cycle: %w", err)
	}

	switch flags.Output {
	case "json":
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("cycle.get", configOverrides)
		fieldSelector, err := fieldfilter.New(flags.Fields, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), cycle, true, fieldSelector)
	case "table":
		if cycle.Name != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Name:     %s\n", *cycle.Name)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Number:   %.0f\n", cycle.Number)
		fmt.Fprintf(cmd.OutOrStdout(), "Progress: %.1f%%\n", cycle.Progress*100)

		// Show scope metrics if available
		if len(cycle.CompletedScopeHistory) > 0 {
			completedScope := cycle.CompletedScopeHistory[len(cycle.CompletedScopeHistory)-1]
			totalScope := cycle.ScopeHistory[len(cycle.ScopeHistory)-1]
			fmt.Fprintf(cmd.OutOrStdout(), "Scope:    %.0f / %.0f points completed\n", completedScope, totalScope)
		}

		// Show issue count metrics if available
		if len(cycle.CompletedIssueCountHistory) > 0 {
			completedIssues := cycle.CompletedIssueCountHistory[len(cycle.CompletedIssueCountHistory)-1]
			totalIssues := cycle.IssueCountHistory[len(cycle.IssueCountHistory)-1]
			fmt.Fprintf(cmd.OutOrStdout(), "Issues:   %.0f / %.0f completed\n", completedIssues, totalIssues)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", flags.Output)
	}
}
