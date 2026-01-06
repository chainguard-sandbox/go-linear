package roadmap

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the roadmap list command.
func NewListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	outputFlags := &cli.OutputFlags{}
	paginationFlags := &cli.PaginationFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all roadmaps",
		Long: `List roadmaps. Returns 4 default fields per roadmap.

Example: go-linear roadmap list --output=json

Related: roadmap_get, project_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client, outputFlags, paginationFlags)
		},
	}

	paginationFlags.Bind(cmd, 50)
	outputFlags.Bind(cmd, "defaults (...) | none | defaults,extra")
	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client, outputFlags *cli.OutputFlags, paginationFlags *cli.PaginationFlags) error {
	ctx := cmd.Context()

	if err := outputFlags.Validate(); err != nil {
		return err
	}

	first := paginationFlags.LimitPtr()

	roadmaps, err := client.Roadmaps(ctx, first, nil)
	if err != nil {
		return fmt.Errorf("failed to list roadmaps: %w", err)
	}

	switch outputFlags.Output {
	case "json":
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("roadmap.list", configOverrides)
		fieldSelector, err := fieldfilter.NewForList(outputFlags.Fields, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), roadmaps, true, fieldSelector)
	case "table":
		if len(roadmaps.Nodes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No roadmaps found")
			return nil
		}
		for _, rm := range roadmaps.Nodes {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", rm.Name)
		}
		return nil
	default:
		return fmt.Errorf("unsupported outputFlags.Output format: %s", outputFlags.Output)
	}
}
