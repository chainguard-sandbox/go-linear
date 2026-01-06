package label

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	labelfilter "github.com/chainguard-sandbox/go-linear/internal/filter/label"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the label list command.
func NewListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	outputFlags := &cli.OutputFlags{}
	paginationFlags := &cli.PaginationFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all issue labels",
		Long: `List labels with filtering. Returns 4 default fields per label. Use for discovering label names.

Filters: --name, --team, --creator, --is-group
Date filters: --created-after, --created-before, --updated-after, --updated-before

Example: go-linear label list --name=bug --output=json

Related: label_get, label_create, issue_add-label`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client, outputFlags, paginationFlags)
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
	outputFlags.Bind(cmd, "defaults (...) | none | defaults,extra")
	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client, outputFlags *cli.OutputFlags, paginationFlags *cli.PaginationFlags) error {
	ctx := cmd.Context()

	if err := outputFlags.Validate(); err != nil {
		return err
	}
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

		switch outputFlags.Output {
		case "json":
			cfg, _ := config.Load()
			var configOverrides map[string]string
			if cfg != nil {
				configOverrides = cfg.FieldDefaults
			}
			defaults := fieldfilter.GetDefaults("label.list", configOverrides)
			fieldSelector, err := fieldfilter.NewForList(outputFlags.Fields, defaults)
			if err != nil {
				return fmt.Errorf("invalid --fields: %w", err)
			}
			return formatter.FormatJSONFiltered(cmd.OutOrStdout(), labels, true, fieldSelector)
		case "table":
			if len(labels.Nodes) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No labels found")
				return nil
			}
			for _, label := range labels.Nodes {
				fmt.Fprintf(cmd.OutOrStdout(), "%s (%s)\n", label.Name, label.Color)
			}
			return nil
		default:
			return fmt.Errorf("unsupported outputFlags.Output format: %s", outputFlags.Output)
		}
	}

	// No filters: use regular query
	labels, err := client.IssueLabels(ctx, first, nil)
	if err != nil {
		return fmt.Errorf("failed to list labels: %w", err)
	}

	switch outputFlags.Output {
	case "json":
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("label.list", configOverrides)
		fieldSelector, err := fieldfilter.NewForList(outputFlags.Fields, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), labels, true, fieldSelector)
	case "table":
		if len(labels.Nodes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No labels found")
			return nil
		}
		for _, label := range labels.Nodes {
			fmt.Fprintf(cmd.OutOrStdout(), "%s (%s)\n", label.Name, label.Color)
		}
		return nil
	default:
		return fmt.Errorf("unsupported outputFlags.Output format: %s", outputFlags.Output)
	}
}
