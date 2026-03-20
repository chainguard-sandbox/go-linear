package template

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
	// Templates returns an array directly, not paginated
	mockTemplatesResponse = `{"data": {"templates": [{"id": "tpl-123", "name": "Bug Report", "description": "Bug template", "templateData": "{\"priority\":1,\"description\":\"## Steps to reproduce\\n\"}", "createdAt": "2024-01-01T00:00:00.000Z", "type": "issue"}]}}`
	mockTemplateResponse  = `{"data": {"template": {"id": "tpl-123", "name": "Bug Report", "description": "Bug template", "templateData": "{\"priority\":1,\"description\":\"## Steps to reproduce\\n\"}", "createdAt": "2024-01-01T00:00:00.000Z", "updatedAt": "2024-01-01T00:00:00.000Z", "type": "issue"}}}`
)

func defaultHandlers() map[string]string {
	return map[string]string{
		"ListTemplates": mockTemplatesResponse,
		"GetTemplate":   mockTemplateResponse,
	}
}
