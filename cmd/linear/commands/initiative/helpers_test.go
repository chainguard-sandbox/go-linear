package initiative

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/testutil"
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
		opName := strings.ToLower(reqBody.OperationName)
		for key, response := range handlers {
			if strings.EqualFold(key, opName) {
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
	mockInitiativesResponse = `{"data": {"initiatives": {"nodes": [{"id": "init-123", "name": "Security Initiative", "description": "Improve security", "createdAt": "2024-01-01T00:00:00.000Z"}], "pageInfo": {"hasNextPage": false}}}}`
	mockInitiativeResponse  = `{
		"data": {
			"initiative": {
				"id": "init-123",
				"name": "Security Initiative",
				"description": "Improve security",
				"createdAt": "2024-01-01T00:00:00.000Z",
				"updatedAt": "2024-01-02T00:00:00.000Z",
				"targetDate": "2024-12-31",
				"sortOrder": 1,
				"color": "#ff0000",
				"icon": "🔒",
				"status": "Active",
				"health": "atRisk",
				"healthUpdatedAt": "2024-01-02T00:00:00.000Z",
				"url": "https://linear.app/test/initiative/init-123",
				"parentInitiative": {
					"id": "init-parent",
					"name": "Parent Initiative"
				},
				"owner": {
					"id": "user-123",
					"name": "Test Owner",
					"email": "owner@example.com"
				},
				"projects": {
					"nodes": [
						{"id": "proj-1", "name": "Test Project 1", "state": "started", "progress": 0.75},
						{"id": "proj-2", "name": "Test Project 2", "state": "planned", "progress": 0.0}
					]
				}
			}
		}
	}`
	mockInitiativeCreateResponse         = `{"data": {"initiativeCreate": {"success": true, "initiative": {"id": "init-new", "name": "New Initiative"}}}}`
	mockInitiativeMutationUpdateResponse = `{"data": {"initiativeUpdate": {"success": true, "initiative": {"id": "init-123", "name": "Updated Initiative"}}}}`
	mockInitiativeDeleteResponse         = `{"data": {"initiativeDelete": {"success": true}}}`

	mockInitiativeUpdateCreateResponse = `{
		"data": {
			"initiativeUpdateCreate": {
				"success": true,
				"initiativeUpdate": {
					"id": "update-123",
					"body": "Test update body",
					"health": "onTrack",
					"createdAt": "2024-01-08T00:00:00.000Z",
					"url": "https://linear.app/test/initiative/init-123/update-123"
				}
			}
		}
	}`

	mockInitiativeUpdatesResponse = `{
		"data": {
			"initiative": {
				"id": "init-123",
				"name": "Security Initiative",
				"initiativeUpdates": {
					"nodes": [
						{
							"id": "update-123",
							"body": "Test update body",
							"health": "onTrack",
							"createdAt": "2024-01-08T00:00:00.000Z",
							"url": "https://linear.app/test/initiative/init-123/update-123",
							"user": {
								"id": "user-123",
								"name": "Test User"
							}
						}
					],
					"pageInfo": {"hasNextPage": false}
				}
			}
		}
	}`

	mockInitiativeUpdateResponse = `{
		"data": {
			"initiativeUpdate": {
				"id": "update-123",
				"body": "Test update body",
				"health": "onTrack",
				"createdAt": "2024-01-08T00:00:00.000Z",
				"url": "https://linear.app/test/initiative/init-123/update-123",
				"user": {
					"id": "user-123",
					"name": "Test User",
					"email": "test@example.com"
				}
			}
		}
	}`

	mockInitiativeUpdateArchiveResponse = `{
		"data": {
			"initiativeUpdateArchive": {
				"success": true
			}
		}
	}`

	mockInitiativeToProjectCreateResponse = `{
		"data": {
			"initiativeToProjectCreate": {
				"success": true,
				"initiativeToProject": {
					"id": "link-123",
					"initiative": {
						"id": "init-123",
						"name": "Security Initiative"
					},
					"project": {
						"id": "project-123",
						"name": "Test Project"
					},
					"createdAt": "2024-01-08T00:00:00.000Z"
				}
			}
		}
	}`

	mockInitiativeToProjectsResponse = `{
		"data": {
			"initiativeToProjects": {
				"nodes": [
					{
						"id": "link-123",
						"initiative": {
							"id": "init-123",
							"name": "Security Initiative"
						},
						"project": {
							"id": "project-123",
							"name": "Test Project"
						},
						"createdAt": "2024-01-08T00:00:00.000Z"
					}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockInitiativeToProjectDeleteResponse = `{
		"data": {
			"initiativeToProjectDelete": {
				"success": true
			}
		}
	}`

	mockInitiativeArchiveResponse = `{
		"data": {
			"initiativeArchive": {
				"success": true
			}
		}
	}`

	mockInitiativeUnarchiveResponse = `{
		"data": {
			"initiativeUnarchive": {
				"success": true
			}
		}
	}`
)

func defaultHandlers() map[string]string {
	handlers := testutil.DefaultHandlers()
	handlers["ListInitiatives"] = mockInitiativesResponse
	handlers["GetInitiative"] = mockInitiativeResponse
	handlers["CreateInitiative"] = mockInitiativeCreateResponse
	handlers["UpdateInitiative"] = mockInitiativeMutationUpdateResponse
	handlers["DeleteInitiative"] = mockInitiativeDeleteResponse
	handlers["CreateInitiativeUpdate"] = mockInitiativeUpdateCreateResponse
	handlers["ListInitiativeUpdates"] = mockInitiativeUpdatesResponse
	handlers["GetInitiativeUpdate"] = mockInitiativeUpdateResponse
	handlers["ArchiveInitiativeUpdate"] = mockInitiativeUpdateArchiveResponse
	handlers["CreateInitiativeToProject"] = mockInitiativeToProjectCreateResponse
	handlers["ListInitiativeToProjects"] = mockInitiativeToProjectsResponse
	handlers["DeleteInitiativeToProject"] = mockInitiativeToProjectDeleteResponse
	handlers["ArchiveInitiative"] = mockInitiativeArchiveResponse
	handlers["UnarchiveInitiative"] = mockInitiativeUnarchiveResponse
	return handlers
}
