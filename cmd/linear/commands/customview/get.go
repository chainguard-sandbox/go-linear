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

// NewGetCommand creates the custom-view get command.
func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a custom view by ID",
		Long: `Get custom view by UUID. Returns detailed fields including filterData.

Example: go-linear custom-view get <uuid>

Related: custom-view_list, custom-view_update`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runGet(cmd, client, args[0])
		},
	}

	cmd.Flags().String("fields", "", "defaults (id,name,description,filterData,shared) | none | defaults,extra")

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, id string) error {
	ctx := cmd.Context()

	view, err := client.CustomView(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get custom view: %w", err)
	}

	fieldsSpec, _ := cmd.Flags().GetString("fields")

	cfg, _ := config.Load()
	var configOverrides map[string]string
	if cfg != nil {
		configOverrides = cfg.FieldDefaults
	}
	defaults := fieldfilter.GetDefaults("custom-view.get", configOverrides)
	fieldSelector, err := fieldfilter.New(fieldsSpec, defaults)
	if err != nil {
		return fmt.Errorf("invalid --fields: %w", err)
	}
	return formatter.FormatJSONFiltered(cmd.OutOrStdout(), view, true, fieldSelector)
}
