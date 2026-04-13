package webhook

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// NewUpdateCommand creates the webhook update command.
func NewUpdateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <webhook-id>",
		Short: "Update an existing webhook",
		Long: `Update webhook properties. All fields are optional.

Example: go-linear webhook update <uuid> --url=https://new-endpoint.com --enabled=false
Example: go-linear webhook update <uuid> --resource-types=Issue,Comment,Project

Related: webhook_get, webhook_create, webhook_delete`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()
			id := args[0]

			input := intgraphql.WebhookUpdateInput{}
			changed := false

			if cmd.Flags().Changed("url") {
				url, _ := cmd.Flags().GetString("url")
				input.URL = &url
				changed = true
			}
			if cmd.Flags().Changed("label") {
				label, _ := cmd.Flags().GetString("label")
				input.Label = &label
				changed = true
			}
			if cmd.Flags().Changed("enabled") {
				enabled, _ := cmd.Flags().GetBool("enabled")
				input.Enabled = &enabled
				changed = true
			}
			if cmd.Flags().Changed("secret") {
				secret, _ := cmd.Flags().GetString("secret")
				input.Secret = &secret
				changed = true
			}
			if cmd.Flags().Changed("resource-types") {
				resourceTypesStr, _ := cmd.Flags().GetString("resource-types")
				var resourceTypes []string
				for rt := range strings.SplitSeq(resourceTypesStr, ",") {
					rt = strings.TrimSpace(rt)
					if rt != "" {
						resourceTypes = append(resourceTypes, rt)
					}
				}
				input.ResourceTypes = resourceTypes
				changed = true
			}

			if !changed {
				return fmt.Errorf("no update flags provided; use --url, --label, --enabled, --secret, or --resource-types")
			}

			result, err := client.WebhookUpdate(ctx, id, input)
			if err != nil {
				return fmt.Errorf("failed to update webhook: %w", err)
			}

			return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
		},
	}

	cmd.Flags().String("url", "", "New webhook endpoint URL")
	cmd.Flags().String("label", "", "New webhook label")
	cmd.Flags().Bool("enabled", true, "Enable or disable the webhook")
	cmd.Flags().String("secret", "", "New signing secret")
	cmd.Flags().String("resource-types", "", "Comma-separated resource types, e.g. Issue,Comment")

	return cmd
}
