package user

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
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
			if strings.EqualFold(key, opName) {
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

func testFactory(t *testing.T, serverURL string) cli.ClientFactory {
	t.Helper()
	return func() (*linear.Client, error) {
		return linear.NewClient("lin_api_test", linear.WithBaseURL(serverURL))
	}
}

const (
	mockUsersResponse = `{
		"data": {
			"users": {
				"nodes": [
					{"id": "user-123", "name": "Test User", "email": "test@example.com", "displayName": "Test", "createdAt": "2024-01-01T00:00:00.000Z", "active": true, "avatarUrl": "https://example.com/avatar.png"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockUserResponse = `{
		"data": {
			"user": {
				"id": "user-123",
				"name": "Test User",
				"email": "test@example.com",
				"displayName": "Test",
				"createdAt": "2024-01-01T00:00:00.000Z",
				"updatedAt": "2024-01-01T00:00:00.000Z",
				"admin": false,
				"active": true,
				"avatarUrl": "https://example.com/avatar.png"
			}
		}
	}`

	mockViewerResponse = `{
		"data": {
			"viewer": {
				"id": "viewer-123",
				"name": "Current User",
				"email": "current@example.com",
				"displayName": "Current",
				"active": true
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

	mockIssuesResponse = `{
		"data": {
			"issues": {
				"nodes": [
					{"id": "issue-123", "identifier": "ENG-123", "title": "Test Issue", "completedAt": "2024-01-01T00:00:00.000Z"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`
)

func defaultHandlers() map[string]string {
	return map[string]string{
		"ListUsers":  mockUsersResponse,
		"GetUser":    mockUserResponse,
		"GetViewer":  mockViewerResponse,
		"ListTeams":  mockTeamsResponse,
		"ListIssues": mockIssuesResponse,
	}
}
