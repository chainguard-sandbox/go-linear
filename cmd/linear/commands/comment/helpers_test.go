package comment

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
	mockCommentsResponse = `{
		"data": {
			"comments": {
				"nodes": [
					{"id": "comment-123", "body": "Test comment", "createdAt": "2024-01-01T00:00:00.000Z", "user": {"name": "Test"}, "url": "https://linear.app/test"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockCommentResponse = `{
		"data": {
			"comment": {
				"id": "comment-123",
				"body": "Test comment",
				"createdAt": "2024-01-01T00:00:00.000Z",
				"updatedAt": "2024-01-01T00:00:00.000Z",
				"url": "https://linear.app/test",
				"user": {"id": "user-123", "name": "Test User"},
				"issue": {"id": "issue-123", "title": "Test Issue"}
			}
		}
	}`

	mockCommentCreateResponse = `{
		"data": {
			"commentCreate": {
				"success": true,
				"comment": {
					"id": "comment-new",
					"body": "New comment",
					"createdAt": "2024-01-01T00:00:00.000Z"
				}
			}
		}
	}`

	mockCommentUpdateResponse = `{
		"data": {
			"commentUpdate": {
				"success": true,
				"comment": {
					"id": "comment-123",
					"body": "Updated comment",
					"updatedAt": "2024-01-02T00:00:00.000Z"
				}
			}
		}
	}`

	mockCommentDeleteResponse = `{
		"data": {
			"commentDelete": {
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
		"ListComments":  mockCommentsResponse,
		"GetComment":    mockCommentResponse,
		"issue":         mockIssueResponse,
		"CreateComment": mockCommentCreateResponse,
		"UpdateComment": mockCommentUpdateResponse,
		"DeleteComment": mockCommentDeleteResponse,
	}
}
