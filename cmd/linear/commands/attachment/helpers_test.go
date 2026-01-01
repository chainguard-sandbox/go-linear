package attachment

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

func mockServer(t *testing.T, handlers map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var reqBody struct {
			Query         string `json:"query"`
			OperationName string `json:"operationName"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		query := strings.ToLower(reqBody.Query)
		opName := strings.ToLower(reqBody.OperationName)
		for key, response := range handlers {
			if strings.ToLower(key) == opName {
				_, _ = w.Write([]byte(response))
				return
			}
		}
		for key, response := range handlers {
			if strings.Contains(query, strings.ToLower(key)) {
				_, _ = w.Write([]byte(response))
				return
			}
		}
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
}

func testFactory(t *testing.T, serverURL string) ClientFactory {
	t.Helper()
	return func() (*linear.Client, error) {
		return linear.NewClient("lin_api_test", linear.WithBaseURL(serverURL))
	}
}

const (
	mockAttachmentsResponse = `{
		"data": {
			"attachments": {
				"nodes": [
					{"id": "att-123", "title": "Design Doc", "url": "https://example.com/doc", "createdAt": "2024-01-01T00:00:00.000Z", "source": "url", "sourceType": "custom"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockAttachmentResponse = `{
		"data": {
			"attachment": {
				"id": "att-123",
				"title": "Design Doc",
				"subtitle": "Design document",
				"url": "https://example.com/doc",
				"createdAt": "2024-01-01T00:00:00.000Z",
				"updatedAt": "2024-01-01T00:00:00.000Z",
				"source": "url",
				"sourceType": "custom"
			}
		}
	}`

	mockAttachmentCreateResponse = `{
		"data": {
			"attachmentCreate": {
				"success": true,
				"attachment": {
					"id": "att-new",
					"title": "New Attachment",
					"subtitle": "Subtitle",
					"url": "https://example.com/new"
				}
			}
		}
	}`

	mockAttachmentLinkURLResponse = `{
		"data": {
			"attachmentLinkURL": {
				"success": true,
				"attachment": {
					"id": "att-link",
					"title": "Linked URL",
					"url": "https://example.com/linked"
				}
			}
		}
	}`

	mockAttachmentLinkGitHubResponse = `{
		"data": {
			"attachmentLinkGitHubPR": {
				"success": true,
				"attachment": {
					"id": "att-github",
					"title": "PR #123",
					"url": "https://github.com/owner/repo/pull/123"
				}
			}
		}
	}`

	mockAttachmentLinkSlackResponse = `{
		"data": {
			"attachmentLinkSlack": {
				"success": true,
				"attachment": {
					"id": "att-slack",
					"title": "Slack Message",
					"url": "https://slack.com/archives/C123/p456"
				}
			}
		}
	}`

	mockAttachmentDeleteResponse = `{
		"data": {
			"attachmentDelete": {
				"success": true
			}
		}
	}`

	mockIssueResponse = `{
		"data": {
			"issue": {
				"id": "issue-123",
				"identifier": "ENG-123"
			}
		}
	}`
)

func defaultHandlers() map[string]string {
	return map[string]string{
		"ListAttachments":       mockAttachmentsResponse,
		"GetAttachment":         mockAttachmentResponse,
		"AttachmentCreate":      mockAttachmentCreateResponse,
		"AttachmentLinkURL":     mockAttachmentLinkURLResponse,
		"AttachmentLinkGitHubPR": mockAttachmentLinkGitHubResponse,
		"AttachmentLinkSlack":   mockAttachmentLinkSlackResponse,
		"AttachmentDelete":      mockAttachmentDeleteResponse,
		"issue":                 mockIssueResponse,
	}
}
