package webhook

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/config"
	"github.com/chainguard-sandbox/go-linear/v2/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewListCommand creates the webhook list command.
func NewListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	pagination := &cli.PaginationFlags{}
	fields := &cli.FieldFlags{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all webhooks",
		Long: `List webhooks in the workspace. Returns key fields per webhook.

Example: go-linear webhook list --limit=20

Related: webhook_get, webhook_create`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client, pagination, fields)
		},
	}

	pagination.Bind(cmd, 50)
	fields.Bind(cmd, "defaults (id,label,url,enabled,resourceTypes,createdAt) | none | defaults,extra")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client, pagination *cli.PaginationFlags, fields *cli.FieldFlags) error {
	ctx := cmd.Context()

	webhooks, err := client.Webhooks(ctx, pagination.LimitPtr(), pagination.AfterPtr())
	if err != nil {
		return fmt.Errorf("failed to list webhooks: %w", err)
	}

	cfg, _ := config.Load()
	var configOverrides map[string]string
	if cfg != nil {
		configOverrides = cfg.FieldDefaults
	}
	defaults := fieldfilter.GetDefaults("webhook.list", configOverrides)
	fieldSelector, err := fieldfilter.NewForList(fields.Fields, defaults)
	if err != nil {
		return fmt.Errorf("invalid --fields: %w", err)
	}
	return formatter.FormatJSONFiltered(cmd.OutOrStdout(), webhooks, true, fieldSelector)
}
