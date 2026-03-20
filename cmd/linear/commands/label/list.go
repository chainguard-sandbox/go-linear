package label

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/config"
	"github.com/chainguard-sandbox/go-linear/v2/internal/fieldfilter"
	labelfilter "github.com/chainguard-sandbox/go-linear/v2/internal/filter/label"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewListCommand creates the label list command.
func NewListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	fieldFlags := &cli.FieldFlags{}
	paginationFlags := &cli.PaginationFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all issue labels",
		Long: `List labels with filtering. Returns 4 default fields per label. Use for discovering label names.

Filters: --name, --team, --creator, --is-group
Date filters: --created-after, --created-before, --updated-after, --updated-before

Example: go-linear label list --name=bug

Related: label_get, label_create, issue_add-label`,
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
	cmd.Flags().String("id", "", "Label UUID")
	cmd.Flags().String("creator", "", "Creator name, email, or 'me'")
	cmd.Flags().String("team", "", "Team name, key, or UUID")

	// Text filters
	cmd.Flags().String("name", "", "Name contains (case-insensitive)")

	// Boolean filters
	cmd.Flags().Bool("is-group", false, "Filter by group labels")

	// Output
	paginationFlags.Bind(cmd, 250)
	fieldFlags.Bind(cmd, "defaults (...) | none | defaults,extra")
	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client, fieldFlags *cli.FieldFlags, paginationFlags *cli.PaginationFlags) error {
	ctx := cmd.Context()

	res := resolver.New(client)

	// Build filter from flags
	filterBuilder := labelfilter.NewFilterBuilder(res)
	if err := filterBuilder.FromFlags(ctx, cmd); err != nil {
		return err
	}
	lblFilter := filterBuilder.Build()

	first := paginationFlags.LimitPtr()

	// Use filtered or unfiltered query based on whether filters were set
	if lblFilter != nil {
		labels, err := client.IssueLabelsFiltered(ctx, first, nil, lblFilter)
		if err != nil {
			return fmt.Errorf("failed to list labels: %w", err)
		}

		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("label.list", configOverrides)
		fieldSelector, err := fieldfilter.NewForList(fieldFlags.Fields, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), labels, true, fieldSelector)
	}

	// No filters: use regular query
	labels, err := client.IssueLabels(ctx, first, nil)
	if err != nil {
		return fmt.Errorf("failed to list labels: %w", err)
	}

	cfg, _ := config.Load()
	var configOverrides map[string]string
	if cfg != nil {
		configOverrides = cfg.FieldDefaults
	}
	defaults := fieldfilter.GetDefaults("label.list", configOverrides)
	fieldSelector, err := fieldfilter.NewForList(fieldFlags.Fields, defaults)
	if err != nil {
		return fmt.Errorf("invalid --fields: %w", err)
	}
	return formatter.FormatJSONFiltered(cmd.OutOrStdout(), labels, true, fieldSelector)
}
