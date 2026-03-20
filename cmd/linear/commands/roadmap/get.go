package roadmap

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/config"
	"github.com/chainguard-sandbox/go-linear/v2/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewGetCommand creates the roadmap get command.
func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
	flags := &cli.FieldFlags{}

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single roadmap by ID",
		Long: `Get roadmap by UUID. Returns 4 default fields.

Example: go-linear roadmap get <uuid>

Related: roadmap_list, project_list`,
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

	flags.Bind(cmd, "defaults (id,name,description,createdAt) | none | defaults,extra")

	return cmd
}
func runGet(cmd *cobra.Command, client *linear.Client, roadmapID string, flags *cli.FieldFlags) error {
	ctx := cmd.Context()

	roadmap, err := client.Roadmap(ctx, roadmapID)
	if err != nil {
		return fmt.Errorf("failed to get roadmap: %w", err)
	}

	cfg, _ := config.Load()
	var configOverrides map[string]string
	if cfg != nil {
		configOverrides = cfg.FieldDefaults
	}
	defaults := fieldfilter.GetDefaults("roadmap.get", configOverrides)
	fieldSelector, err := fieldfilter.New(flags.Fields, defaults)
	if err != nil {
		return fmt.Errorf("invalid --fields: %w", err)
	}
	return formatter.FormatJSONFiltered(cmd.OutOrStdout(), roadmap, true, fieldSelector)
}
