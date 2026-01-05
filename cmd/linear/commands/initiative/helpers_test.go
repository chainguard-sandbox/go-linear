package initiative

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
	mockInitiativesResponse      = `{"data": {"initiatives": {"nodes": [{"id": "init-123", "name": "Security Initiative", "description": "Improve security", "createdAt": "2024-01-01T00:00:00.000Z"}], "pageInfo": {"hasNextPage": false}}}}`
	mockInitiativeResponse       = `{"data": {"initiative": {"id": "init-123", "name": "Security Initiative", "description": "Improve security", "createdAt": "2024-01-01T00:00:00.000Z"}}}`
	mockInitiativeCreateResponse = `{"data": {"initiativeCreate": {"success": true, "initiative": {"id": "init-new", "name": "New Initiative"}}}}`
	mockInitiativeUpdateResponse = `{"data": {"initiativeUpdate": {"success": true, "initiative": {"id": "init-123", "name": "Updated Initiative"}}}}`
)

func defaultHandlers() map[string]string {
	return map[string]string{
		"ListInitiatives":  mockInitiativesResponse,
		"GetInitiative":    mockInitiativeResponse,
		"CreateInitiative": mockInitiativeCreateResponse,
		"UpdateInitiative": mockInitiativeUpdateResponse,
	}
}
