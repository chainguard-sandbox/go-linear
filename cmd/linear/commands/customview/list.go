package customview

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/config"
	"github.com/chainguard-sandbox/go-linear/v2/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewListCommand creates the custom-view list command.
func NewListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List custom views",
		Long: `List custom views. Returns default fields per view.

Example: go-linear custom-view list
         go-linear custom-view list --limit=10

Related: custom-view_get, custom-view_create`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client)
		},
	}

	cmd.Flags().IntP("limit", "l", 50, "Number to return")
	cmd.Flags().String("fields", "", "defaults (id,name,shared,modelName) | none | defaults,extra")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	views, err := client.CustomViews(ctx, &first, nil)
	if err != nil {
		return fmt.Errorf("failed to list custom views: %w", err)
	}

	fieldsSpec, _ := cmd.Flags().GetString("fields")

	cfg, _ := config.Load()
	var configOverrides map[string]string
	if cfg != nil {
		configOverrides = cfg.FieldDefaults
	}
	defaults := fieldfilter.GetDefaults("custom-view.list", configOverrides)
	fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
	if err != nil {
		return fmt.Errorf("invalid --fields: %w", err)
	}
	return formatter.FormatJSONFiltered(cmd.OutOrStdout(), views, true, fieldSelector)
}
