package state

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/config"
	"github.com/chainguard-sandbox/go-linear/v2/internal/fieldfilter"
	statefilter "github.com/chainguard-sandbox/go-linear/v2/internal/filter/state"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewListCommand creates the state list command.
func NewListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	fieldFlags := &cli.FieldFlags{}
	paginationFlags := &cli.PaginationFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all workflow states",
		Long: `List workflow states with filtering. Returns 5 default fields per state. Use for discovering state names.

States define issue lifecycle: triage, backlog, unstarted, started, completed, canceled.

Filters: --name, --type, --team, --description
Date filters: --created-after, --created-before, --updated-after, --updated-before

Example: go-linear state list --type=started

Related: state_get, issue_update, issue_list`,
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

	// Entity filters
	cmd.Flags().String("id", "", "Workflow state UUID")
	cmd.Flags().String("team", "", "Team name, key, or UUID")

	// Text filters
	cmd.Flags().String("name", "", "Name contains (case-insensitive)")
	cmd.Flags().String("description", "", "Description contains (case-insensitive)")
	cmd.Flags().String("type", "", "State type: triage, backlog, unstarted, started, completed, canceled")

	// Output
	paginationFlags.Bind(cmd, 100)
	fieldFlags.Bind(cmd, "defaults (...) | none | defaults,extra")
	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client, fieldFlags *cli.FieldFlags, paginationFlags *cli.PaginationFlags) error {
	ctx := cmd.Context()

	res := resolver.New(client)

	// Build filter from flags
	filterBuilder := statefilter.NewFilterBuilder(res)
	if err := filterBuilder.FromFlags(ctx, cmd); err != nil {
		return err
	}
	stateFilter := filterBuilder.Build()

	first := paginationFlags.LimitPtr()

	// Use filtered or unfiltered query based on whether filters were set
	if stateFilter != nil {
		states, err := client.WorkflowStatesFiltered(ctx, first, nil, stateFilter)
		if err != nil {
			return fmt.Errorf("failed to list workflow states: %w", err)
		}

		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("state.list", configOverrides)
		fieldSelector, err := fieldfilter.NewForList(fieldFlags.Fields, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), states, true, fieldSelector)
	}

	// No filters: use regular query
	states, err := client.WorkflowStates(ctx, first, nil)
	if err != nil {
		return fmt.Errorf("failed to list workflow states: %w", err)
	}

	cfg, _ := config.Load()
	var configOverrides map[string]string
	if cfg != nil {
		configOverrides = cfg.FieldDefaults
	}
	defaults := fieldfilter.GetDefaults("state.list", configOverrides)
	fieldSelector, err := fieldfilter.NewForList(fieldFlags.Fields, defaults)
	if err != nil {
		return fmt.Errorf("invalid --fields: %w", err)
	}
	return formatter.FormatJSONFiltered(cmd.OutOrStdout(), states, true, fieldSelector)
}
