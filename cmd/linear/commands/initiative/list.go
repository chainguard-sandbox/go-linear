package initiative

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
)

func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all initiatives",
		Long: `List initiatives. Returns 4 default fields per initiative.

Example: go-linear-cli initiative list --output=json

Related: initiative_get, project_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()
			limit, _ := cmd.Flags().GetInt("limit")
			first := int64(limit)

			initiatives, err := client.Initiatives(ctx, &first, nil)
			if err != nil {
				return fmt.Errorf("failed to list initiatives: %w", err)
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
				defaults := fieldfilter.GetDefaults("initiative.list", configOverrides)
				fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
				if err != nil {
					return fmt.Errorf("invalid --fields: %w", err)
				}
				return formatter.FormatJSONFiltered(cmd.OutOrStdout(), initiatives, true, fieldSelector)
			case "table":
				for _, init := range initiatives.Nodes {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", init.Name)
				}
				return nil
			default:
				return fmt.Errorf("unsupported output format: %s", output)
			}
		},
	}

	cmd.Flags().IntP("limit", "l", 50, "Number to return")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,name,description,createdAt) | none | defaults,extra")

	return cmd
}
