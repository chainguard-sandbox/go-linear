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

// NewGetCommand creates the webhook get command.
func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
	fields := &cli.FieldFlags{}

	cmd := &cobra.Command{
		Use:   "get <webhook-id>",
		Short: "Get a single webhook by ID",
		Long: `Get webhook by UUID. Returns full webhook configuration.

Example: go-linear webhook get <uuid>

Related: webhook_list, webhook_update`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runGet(cmd, client, args[0], fields)
		},
	}

	fields.Bind(cmd, "defaults (id,label,url,enabled,resourceTypes,createdAt) | none | defaults,extra")

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, webhookID string, fields *cli.FieldFlags) error {
	ctx := cmd.Context()

	webhook, err := client.Webhook(ctx, webhookID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	cfg, _ := config.Load()
	var configOverrides map[string]string
	if cfg != nil {
		configOverrides = cfg.FieldDefaults
	}
	defaults := fieldfilter.GetDefaults("webhook.get", configOverrides)
	fieldSelector, err := fieldfilter.New(fields.Fields, defaults)
	if err != nil {
		return fmt.Errorf("invalid --fields: %w", err)
	}
	return formatter.FormatJSONFiltered(cmd.OutOrStdout(), webhook, true, fieldSelector)
}
