package webhook

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// NewCreateCommand creates the webhook create command.
func NewCreateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new webhook",
		Long: `Create a webhook to receive notifications on data changes.

Required: --url, --resource-types
Optional: --team, --label, --enabled, --secret, --all-public-teams

Resource types: Issue, Comment, Project, Cycle, IssueLabel, Reaction, etc.

Example: go-linear webhook create --url=https://example.com/hook --resource-types=Issue,Comment --label="My Hook"

Related: webhook_list, webhook_get, webhook_update`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()

			url, _ := cmd.Flags().GetString("url")
			resourceTypesStr, _ := cmd.Flags().GetString("resource-types")

			// Parse resource types from comma-separated string
			var resourceTypes []string
			for rt := range strings.SplitSeq(resourceTypesStr, ",") {
				rt = strings.TrimSpace(rt)
				if rt != "" {
					resourceTypes = append(resourceTypes, rt)
				}
			}
			if len(resourceTypes) == 0 {
				return fmt.Errorf("--resource-types must contain at least one resource type")
			}

			input := intgraphql.WebhookCreateInput{
				URL:           url,
				ResourceTypes: resourceTypes,
			}

			if label, _ := cmd.Flags().GetString("label"); label != "" {
				input.Label = &label
			}
			if teamID, _ := cmd.Flags().GetString("team"); teamID != "" {
				input.TeamID = &teamID
			}
			if secret, _ := cmd.Flags().GetString("secret"); secret != "" {
				input.Secret = &secret
			}
			if cmd.Flags().Changed("enabled") {
				enabled, _ := cmd.Flags().GetBool("enabled")
				input.Enabled = &enabled
			}
			if cmd.Flags().Changed("all-public-teams") {
				allPublic, _ := cmd.Flags().GetBool("all-public-teams")
				input.AllPublicTeams = &allPublic
			}

			result, err := client.WebhookCreate(ctx, input)
			if err != nil {
				return fmt.Errorf("failed to create webhook: %w", err)
			}

			return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
		},
	}

	cmd.Flags().String("url", "", "Webhook endpoint URL (required)")
	_ = cmd.MarkFlagRequired("url")
	cmd.Flags().String("resource-types", "", "Comma-separated resource types, e.g. Issue,Comment (required)")
	_ = cmd.MarkFlagRequired("resource-types")
	cmd.Flags().String("label", "", "Webhook label")
	cmd.Flags().String("team", "", "Team ID to scope webhook to")
	cmd.Flags().String("secret", "", "Signing secret for payload verification")
	cmd.Flags().Bool("enabled", true, "Whether the webhook is enabled")
	cmd.Flags().Bool("all-public-teams", false, "Enable for all public teams")

	return cmd
}
