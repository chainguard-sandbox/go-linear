package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// Webhook retrieves a single webhook by ID.
//
// Returns the webhook with its full configuration including URL, enabled state,
// resource types, team association, and creator.
//
// Permissions Required: Admin
//
// Related: [Webhooks], [WebhookCreate]
func (c *Client) Webhook(ctx context.Context, id string) (*intgraphql.GetWebhook_Webhook, error) {
	resp, err := c.gqlClient.GetWebhook(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("webhook query", err)
	}
	return &resp.Webhook, nil
}

// Webhooks retrieves a paginated list of webhooks.
//
// Parameters:
//   - first: Number of webhooks to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Webhooks.Nodes: Array of webhooks (may be empty)
//   - Webhooks.PageInfo.HasNextPage: true if more results available
//   - Webhooks.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Admin
//
// Related: [Webhook], [WebhookCreate]
func (c *Client) Webhooks(ctx context.Context, first *int64, after *string) (*intgraphql.ListWebhooks_Webhooks, error) {
	resp, err := c.gqlClient.ListWebhooks(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("webhooks query", err)
	}
	return &resp.Webhooks, nil
}

// WebhookCreate creates a new webhook.
//
// Required Input Fields:
//   - URL: Webhook endpoint URL (string)
//   - ResourceTypes: Resource types to subscribe to ([]string, e.g. "Issue", "Comment")
//
// Optional Input Fields:
//   - TeamID: Team to scope webhook to (*string)
//   - Label: Display label (*string)
//   - Enabled: Enable/disable (*bool, default true)
//   - Secret: Signing secret (*string)
//   - AllPublicTeams: Subscribe to all public teams (*bool)
//
// Permissions Required: Admin
//
// Related: [WebhookUpdate], [WebhookDelete], [Webhooks]
func (c *Client) WebhookCreate(ctx context.Context, input intgraphql.WebhookCreateInput) (*intgraphql.WebhookCreate_WebhookCreate_Webhook, error) {
	resp, err := c.gqlClient.WebhookCreate(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("WebhookCreate", err)
	}
	if !resp.WebhookCreate.Success {
		return nil, errMutationFailed("WebhookCreate")
	}
	return &resp.WebhookCreate.Webhook, nil
}

// WebhookUpdate updates an existing webhook.
//
// Parameters:
//   - id: Webhook UUID (required)
//   - input: Fields to update (all optional, nil = unchanged)
//
// Permissions Required: Admin
//
// Related: [WebhookCreate], [WebhookDelete], [Webhook]
func (c *Client) WebhookUpdate(ctx context.Context, id string, input intgraphql.WebhookUpdateInput) (*intgraphql.WebhookUpdate_WebhookUpdate_Webhook, error) {
	resp, err := c.gqlClient.WebhookUpdate(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("WebhookUpdate", err)
	}
	if !resp.WebhookUpdate.Success {
		return nil, errMutationFailed("WebhookUpdate")
	}
	return &resp.WebhookUpdate.Webhook, nil
}

// WebhookDelete deletes a webhook.
//
// Parameters:
//   - id: Webhook UUID to delete (required)
//
// Returns:
//   - nil: Webhook successfully deleted
//   - error: Non-nil if webhook not found or deletion fails
//
// Permissions Required: Admin
//
// Related: [WebhookCreate], [Webhook]
func (c *Client) WebhookDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.WebhookDelete(ctx, id)
	if err != nil {
		return wrapGraphQLError("WebhookDelete", err)
	}
	if !resp.WebhookDelete.Success {
		return errMutationFailed("WebhookDelete")
	}
	return nil
}

// WebhookRotateSecret rotates the signing secret for a webhook.
//
// Parameters:
//   - id: Webhook UUID (required)
//
// Returns:
//   - The new signing secret string
//   - error: Non-nil if rotation fails
//
// Permissions Required: Admin
//
// Related: [WebhookCreate], [WebhookUpdate]
func (c *Client) WebhookRotateSecret(ctx context.Context, id string) (string, error) {
	resp, err := c.gqlClient.WebhookRotateSecret(ctx, id)
	if err != nil {
		return "", wrapGraphQLError("WebhookRotateSecret", err)
	}
	if !resp.WebhookRotateSecret.Success {
		return "", errMutationFailed("WebhookRotateSecret")
	}
	return resp.WebhookRotateSecret.Secret, nil
}
