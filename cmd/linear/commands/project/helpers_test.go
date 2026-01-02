package project

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

func testFactory(t *testing.T, serverURL string) ClientFactory {
	t.Helper()
	return func() (*linear.Client, error) {
		return linear.NewClient("lin_api_test", linear.WithBaseURL(serverURL))
	}
}

const (
	mockProjectsResponse = `{
		"data": {
			"projects": {
				"nodes": [
					{"id": "proj-123", "name": "Test Project", "description": "Test desc", "createdAt": "2024-01-01T00:00:00.000Z"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockProjectResponse = `{
		"data": {
			"project": {
				"id": "proj-123",
				"name": "Test Project",
				"description": "Test description",
				"createdAt": "2024-01-01T00:00:00.000Z",
				"color": "#3b82f6",
				"state": "started"
			}
		}
	}`

	mockProjectCreateResponse = `{
		"data": {
			"projectCreate": {
				"success": true,
				"project": {
					"id": "proj-new",
					"name": "New Project",
					"createdAt": "2024-01-01T00:00:00.000Z"
				}
			}
		}
	}`

	mockProjectUpdateResponse = `{
		"data": {
			"projectUpdate": {
				"success": true,
				"project": {
					"id": "proj-123",
					"name": "Updated Project",
					"updatedAt": "2024-01-02T00:00:00.000Z"
				}
			}
		}
	}`

	mockProjectDeleteResponse = `{
		"data": {
			"projectDelete": {
				"success": true
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

	mockMilestoneCreateResponse = `{
		"data": {
			"projectMilestoneCreate": {
				"success": true,
				"projectMilestone": {
					"id": "milestone-new",
					"name": "Q1 2025",
					"description": "Q1 milestone",
					"targetDate": "2025-03-31",
					"sortOrder": 0,
					"project": {"id": "proj-123", "name": "Test Project"}
				}
			}
		}
	}`

	mockMilestoneUpdateResponse = `{
		"data": {
			"projectMilestoneUpdate": {
				"success": true,
				"projectMilestone": {
					"id": "milestone-123",
					"name": "Q2 2025",
					"description": "Updated milestone",
					"targetDate": "2025-06-30",
					"sortOrder": 1
				}
			}
		}
	}`

	mockMilestoneDeleteResponse = `{
		"data": {
			"projectMilestoneDelete": {
				"success": true
			}
		}
	}`
)

func defaultHandlers() map[string]string {
	return map[string]string{
		"ListProjects":           mockProjectsResponse,
		"GetProject":             mockProjectResponse,
		"ListTeams":              mockTeamsResponse,
		"CreateProject":          mockProjectCreateResponse,
		"UpdateProject":          mockProjectUpdateResponse,
		"DeleteProject":          mockProjectDeleteResponse,
		"ProjectMilestoneCreate": mockMilestoneCreateResponse,
		"ProjectMilestoneUpdate": mockMilestoneUpdateResponse,
		"ProjectMilestoneDelete": mockMilestoneDeleteResponse,
	}
}
