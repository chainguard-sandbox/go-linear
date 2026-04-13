package webhook

const (
	mockWebhooksResponse = `{
		"data": {
			"webhooks": {
				"nodes": [
					{"id": "wh-123", "label": "My Hook", "url": "https://example.com/hook", "enabled": true, "resourceTypes": ["Issue"], "createdAt": "2024-01-01T00:00:00.000Z"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockWebhookResponse = `{
		"data": {
			"webhook": {
				"id": "wh-123",
				"label": "My Hook",
				"url": "https://example.com/hook",
				"enabled": true,
				"resourceTypes": ["Issue", "Comment"],
				"createdAt": "2024-01-01T00:00:00.000Z"
			}
		}
	}`

	mockWebhookCreateResponse = `{
		"data": {
			"webhookCreate": {
				"success": true,
				"webhook": {
					"id": "wh-999",
					"label": "New Hook",
					"url": "https://example.com/new",
					"enabled": true
				}
			}
		}
	}`

	mockWebhookUpdateResponse = `{
		"data": {
			"webhookUpdate": {
				"success": true,
				"webhook": {
					"id": "wh-123",
					"label": "Updated Hook"
				}
			}
		}
	}`

	mockWebhookDeleteResponse = `{
		"data": {
			"webhookDelete": {
				"success": true
			}
		}
	}`

	mockWebhookRotateSecretResponse = `{
		"data": {
			"webhookRotateSecret": {
				"success": true,
				"secret": "whsec_new_secret_value"
			}
		}
	}`

	mockTeamsResponse = `{
		"data": {
			"teams": {
				"nodes": [{"id": "team-123", "key": "ENG", "name": "Engineering"}],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`
)

func defaultHandlers() map[string]string {
	return map[string]string{
		"ListWebhooks":       mockWebhooksResponse,
		"GetWebhook":         mockWebhookResponse,
		"WebhookCreate":      mockWebhookCreateResponse,
		"WebhookUpdate":      mockWebhookUpdateResponse,
		"WebhookDelete":      mockWebhookDeleteResponse,
		"WebhookRotateSecret": mockWebhookRotateSecretResponse,
		"ListTeams":           mockTeamsResponse,
	}
}
