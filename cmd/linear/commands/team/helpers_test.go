package team

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// mockServer creates a test server that handles various Linear API queries.
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
			if strings.ToLower(key) == opName {
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

// Standard mock responses for team queries
const (
	mockTeamsResponse = `{
		"data": {
			"teams": {
				"nodes": [
					{"id": "team-123", "key": "ENG", "name": "Engineering", "description": "Engineering team", "createdAt": "2024-01-01T00:00:00.000Z"},
					{"id": "team-456", "key": "PLT", "name": "Platform", "description": "Platform team", "createdAt": "2024-01-02T00:00:00.000Z"}
				],
				"pageInfo": {"hasNextPage": false, "endCursor": null}
			}
		}
	}`

	mockTeamResponse = `{
		"data": {
			"team": {
				"id": "team-123",
				"key": "ENG",
				"name": "Engineering",
				"description": "Engineering team",
				"icon": "🛠",
				"createdAt": "2024-01-01T00:00:00.000Z",
				"color": "#3b82f6",
				"private": false
			}
		}
	}`

	mockTeamMembersResponse = `{
		"data": {
			"team": {
				"id": "team-123",
				"members": {
					"nodes": [
						{"id": "user-123", "name": "Test User", "email": "test@example.com"},
						{"id": "user-456", "name": "Another User", "email": "another@example.com"}
					],
					"pageInfo": {"hasNextPage": false}
				}
			}
		}
	}`

	mockTeamCreateResponse = `{
		"data": {
			"teamCreate": {
				"success": true,
				"team": {
					"id": "team-new",
					"key": "NEW",
					"name": "New Team",
					"description": "A new team",
					"createdAt": "2024-01-01T00:00:00.000Z"
				}
			}
		}
	}`

	mockTeamUpdateResponse = `{
		"data": {
			"teamUpdate": {
				"success": true,
				"team": {
					"id": "team-123",
					"key": "ENG",
					"name": "Updated Engineering",
					"description": "Updated description",
					"updatedAt": "2024-01-02T00:00:00.000Z"
				}
			}
		}
	}`

	mockTeamDeleteResponse = `{
		"data": {
			"teamDelete": {
				"success": true
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
)

// defaultHandlers returns a map of handlers for common queries.
func defaultHandlers() map[string]string {
	return map[string]string{
		// Queries
		"teams":       mockTeamsResponse,
		"team":        mockTeamResponse,
		"users":       mockUsersResponse,
		"viewer":      mockViewerResponse,
		"teamMembers": mockTeamMembersResponse,

		// Mutations
		"CreateTeam": mockTeamCreateResponse,
		"UpdateTeam": mockTeamUpdateResponse,
		"DeleteTeam": mockTeamDeleteResponse,
	}
}
