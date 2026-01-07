package initiative

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
)

func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
	flags := &cli.OutputFlags{}

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single initiative by ID",
		Long: `Get initiative by UUID. Returns 4 default fields.

Example: go-linear initiative get <uuid> --output=json

Related: initiative_list, project_list`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()
			initiative, err := client.Initiative(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get initiative: %w", err)
			}

			switch flags.Output {
			case "json":
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
			case "table":
				fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", initiative.Name)
				return nil
			default:
				return fmt.Errorf("unsupported output format: %s", flags.Output)
			}
		},
	}

	flags.Bind(cmd, "defaults (id,name,description,createdAt) | none | defaults,extra")

	return cmd
}
