package cycle

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
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
	mockCyclesResponse = `{
		"data": {
			"cycles": {
				"nodes": [
					{"id": "cycle-123", "name": "Sprint 1", "startsAt": "2024-01-01T00:00:00.000Z", "endsAt": "2024-01-14T00:00:00.000Z", "createdAt": "2024-01-01T00:00:00.000Z"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockCycleResponse = `{
		"data": {
			"cycle": {
				"id": "cycle-123",
				"name": "Sprint 1",
				"startsAt": "2024-01-01T00:00:00.000Z",
				"endsAt": "2024-01-14T00:00:00.000Z",
				"createdAt": "2024-01-01T00:00:00.000Z",
				"description": "Sprint description"
			}
		}
	}`

	mockCycleCreateResponse = `{
		"data": {
			"cycleCreate": {
				"success": true,
				"cycle": {
					"id": "cycle-new",
					"number": 1,
					"name": "New Sprint",
					"startsAt": "2024-01-15T00:00:00.000Z",
					"endsAt": "2024-01-28T00:00:00.000Z",
					"team": {"id": "team-123", "name": "Engineering", "key": "ENG"}
				}
			}
		}
	}`

	mockCycleUpdateResponse = `{
		"data": {
			"cycleUpdate": {
				"success": true,
				"cycle": {
					"id": "cycle-123",
					"number": 1,
					"name": "Updated Sprint",
					"startsAt": "2024-01-01T00:00:00.000Z",
					"endsAt": "2024-01-14T00:00:00.000Z"
				}
			}
		}
	}`

	mockCycleArchiveResponse = `{
		"data": {
			"cycleArchive": {
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
)

func defaultHandlers() map[string]string {
	return map[string]string{
		"ListCycles":   mockCyclesResponse,
		"GetCycle":     mockCycleResponse,
		"ListTeams":    mockTeamsResponse,
		"CreateCycle":  mockCycleCreateResponse,
		"UpdateCycle":  mockCycleUpdateResponse,
		"ArchiveCycle": mockCycleArchiveResponse,
	}
}
