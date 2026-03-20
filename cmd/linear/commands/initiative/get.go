package initiative

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/config"
	"github.com/chainguard-sandbox/go-linear/v2/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
)

func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
	flags := &cli.FieldFlags{}

	cmd := &cobra.Command{
		Use:   "get <name-or-id>",
		Short: "Get a single initiative by name or ID",
		Long: `Get initiative by name or UUID. Returns 4 default fields. Shows parent initiative and sub-initiative count in hierarchy.

Example: go-linear initiative get "Shrink Wolfi"
Example: go-linear initiative get <uuid>

Related: initiative_list, project_list`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()

			// Resolve initiative name to UUID
			res := resolver.New(client)
			initiativeID, err := res.ResolveInitiative(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve initiative: %w", err)
			}

			initiative, err := client.Initiative(ctx, initiativeID)
			if err != nil {
				return fmt.Errorf("failed to get initiative: %w", err)
			}

			cfg, _ := config.Load()
			var configOverrides map[string]string
			if cfg != nil {
				configOverrides = cfg.FieldDefaults
			}
			defaults := fieldfilter.GetDefaults("initiative.get", configOverrides)
			fieldSelector, err := fieldfilter.New(flags.Fields, defaults)
			if err != nil {
				return fmt.Errorf("invalid --fields: %w", err)
			}
			return formatter.FormatJSONFiltered(cmd.OutOrStdout(), initiative, true, fieldSelector)
		},
	}

	flags.Bind(cmd, "defaults (id,name,description,createdAt) | none | defaults,extra")

	return cmd
}
