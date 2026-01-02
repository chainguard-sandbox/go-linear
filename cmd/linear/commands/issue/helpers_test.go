package issue

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// mockServer creates a test server that handles various Linear API queries.
// It returns mock responses based on the operation name in the GraphQL request.
func mockServer(t *testing.T, handlers map[string]string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Parse request body to determine operation
		var reqBody struct {
			Query         string         `json:"query"`
			OperationName string         `json:"operationName"`
			Variables     map[string]any `json:"variables"`
		}

		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Logf("Failed to decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		query := strings.ToLower(reqBody.Query)
		opName := strings.ToLower(reqBody.OperationName)

		// Try to match by operation name first (case insensitive)
		for key, response := range handlers {
			if strings.EqualFold(key, opName) {
				_, _ = w.Write([]byte(response))
				return
			}
		}

		// Fall back to query content matching (case insensitive)
		for key, response := range handlers {
			if strings.Contains(query, strings.ToLower(key)) {
				_, _ = w.Write([]byte(response))
				return
			}
		}

		// Default: empty response
		t.Logf("No handler matched for query: %s (operation: %s)", reqBody.Query[:min(100, len(reqBody.Query))], reqBody.OperationName)
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
}

// testFactory creates a ClientFactory that connects to the given test server.
func testFactory(t *testing.T, serverURL string) ClientFactory {
	t.Helper()
	return func() (*linear.Client, error) {
		return linear.NewClient("lin_api_test", linear.WithBaseURL(serverURL))
	}
}

// Standard mock responses for common queries
const (
	mockTeamsResponse = `{
		"data": {
			"teams": {
				"nodes": [
					{"id": "team-123", "key": "ENG", "name": "Engineering"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockUsersResponse = `{
		"data": {
			"users": {
				"nodes": [
					{"id": "user-123", "name": "Test User", "email": "test@example.com", "displayName": "test"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockViewerResponse = `{
		"data": {
			"viewer": {
				"id": "user-123",
				"name": "Test User",
				"email": "test@example.com",
				"displayName": "test",
				"active": true,
				"admin": false,
				"createdAt": "2024-01-01T00:00:00.000Z"
			}
		}
	}`

	mockStatesResponse = `{
		"data": {
			"workflowStates": {
				"nodes": [
					{"id": "state-123", "name": "Todo", "type": "unstarted"},
					{"id": "state-456", "name": "In Progress", "type": "started"},
					{"id": "state-789", "name": "Done", "type": "completed"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockLabelsResponse = `{
		"data": {
			"issueLabels": {
				"nodes": [
					{"id": "label-123", "name": "bug"},
					{"id": "label-456", "name": "feature"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockIssueResponse = `{
		"data": {
			"issue": {
				"id": "issue-123",
				"identifier": "ENG-123",
				"title": "Test Issue",
				"description": "Test description",
				"priority": 2,
				"state": {"id": "state-456", "name": "In Progress"},
				"team": {"id": "team-123", "key": "ENG", "name": "Engineering"},
				"assignee": {"id": "user-123", "name": "Test User"},
				"createdAt": "2024-01-01T00:00:00.000Z",
				"updatedAt": "2024-01-02T00:00:00.000Z"
			}
		}
	}`

	mockIssuesResponse = `{
		"data": {
			"issues": {
				"nodes": [
					{
						"id": "issue-123",
						"identifier": "ENG-123",
						"title": "Test Issue",
						"priority": 2,
						"state": {"id": "state-456", "name": "In Progress"},
						"team": {"id": "team-123", "key": "ENG"},
						"createdAt": "2024-01-01T00:00:00.000Z",
						"updatedAt": "2024-01-02T00:00:00.000Z"
					}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockIssueCreateResponse = `{
		"data": {
			"issueCreate": {
				"success": true,
				"issue": {
					"id": "issue-new",
					"identifier": "ENG-999",
					"title": "New Issue",
					"priority": 1,
					"number": 999,
					"url": "https://linear.app/test/issue/ENG-999",
					"state": {"id": "state-123", "name": "Todo"},
					"team": {"id": "team-123", "key": "ENG", "name": "Engineering"},
					"createdAt": "2024-01-01T00:00:00.000Z"
				}
			}
		}
	}`

	mockIssueUpdateResponse = `{
		"data": {
			"issueUpdate": {
				"success": true,
				"issue": {
					"id": "issue-123",
					"identifier": "ENG-123",
					"title": "Updated Issue",
					"description": "Updated description",
					"priority": 1,
					"updatedAt": "2024-01-02T00:00:00.000Z"
				}
			}
		}
	}`

	mockIssueDeleteResponse = `{
		"data": {
			"issueDelete": {
				"success": true
			}
		}
	}`

	mockSearchResponse = `{
		"data": {
			"searchIssues": {
				"nodes": [
					{
						"id": "issue-123",
						"identifier": "ENG-123",
						"title": "Test Issue with search term",
						"priority": 2,
						"state": {"id": "state-456", "name": "In Progress"},
						"team": {"id": "team-123", "key": "ENG"},
						"createdAt": "2024-01-01T00:00:00.000Z",
						"updatedAt": "2024-01-02T00:00:00.000Z"
					}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockIssueLabelResponse = `{
		"data": {
			"issueAddLabel": {
				"success": true,
				"issue": {
					"id": "issue-123",
					"title": "Test Issue",
					"labels": {
						"nodes": [{"id": "label-123", "name": "bug"}]
					}
				}
			}
		}
	}`

	mockIssueRemoveLabelResponse = `{
		"data": {
			"issueRemoveLabel": {
				"success": true,
				"issue": {
					"id": "issue-123",
					"title": "Test Issue",
					"labels": {"nodes": []}
				}
			}
		}
	}`

	mockIssueRelationCreateResponse = `{
		"data": {
			"issueRelationCreate": {
				"success": true,
				"issueRelation": {
					"id": "relation-123",
					"type": "blocks",
					"issue": {"id": "issue-123", "title": "Test Issue"},
					"relatedIssue": {"id": "issue-456", "title": "Related Issue"}
				}
			}
		}
	}`

	mockIssueRelationDeleteResponse = `{
		"data": {
			"issueRelationDelete": {
				"success": true
			}
		}
	}`

	mockIssueRelationUpdateResponse = `{
		"data": {
			"issueRelationUpdate": {
				"success": true,
				"issueRelation": {
					"id": "relation-123",
					"type": "related",
					"issue": {"id": "issue-123", "title": "Test Issue"},
					"relatedIssue": {"id": "issue-456", "title": "Related Issue"}
				}
			}
		}
	}`
)

// defaultHandlers returns a map of handlers for common queries.
// Keys should match either the GraphQL operation name or a substring of the query.
func defaultHandlers() map[string]string {
	return map[string]string{
		// Queries - match by query name in document
		"teams":          mockTeamsResponse,
		"users":          mockUsersResponse,
		"viewer":         mockViewerResponse,
		"workflowstates": mockStatesResponse,
		"issuelabels":    mockLabelsResponse,
		"issue":          mockIssueResponse,
		"issues":         mockIssuesResponse,
		"searchissues":   mockSearchResponse,

		// Mutations - match by operation name (gqlgenc uses these)
		"CreateIssue":         mockIssueCreateResponse,
		"UpdateIssue":         mockIssueUpdateResponse,
		"DeleteIssue":         mockIssueDeleteResponse,
		"IssueAddLabel":       mockIssueLabelResponse,
		"IssueRemoveLabel":    mockIssueRemoveLabelResponse,
		"IssueRelationCreate": mockIssueRelationCreateResponse,
		"IssueRelationDelete": mockIssueRelationDeleteResponse,
		"IssueRelationUpdate": mockIssueRelationUpdateResponse,
	}
}
