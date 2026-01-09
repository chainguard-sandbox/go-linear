package document

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
	mockDocumentsResponse = `{"data": {"documents": {"nodes": [{"id": "doc-123", "title": "Design Doc", "content": "Content here", "createdAt": "2024-01-01T00:00:00.000Z"}], "pageInfo": {"hasNextPage": false}}}}`
	mockDocumentResponse  = `{"data": {"document": {"id": "doc-123", "title": "Design Doc", "content": "Content here", "createdAt": "2024-01-01T00:00:00.000Z"}}}`

	mockDocumentCreateResponse = `{
		"data": {
			"documentCreate": {
				"success": true,
				"document": {
					"id": "doc-new",
					"title": "New Document",
					"content": "New content",
					"url": "https://linear.app/test/doc/doc-new",
					"createdAt": "2024-01-08T00:00:00.000Z",
					"creator": {
						"id": "user-123",
						"name": "Test User"
					}
				}
			}
		}
	}`

	mockDocumentUpdateResponse = `{
		"data": {
			"documentUpdate": {
				"success": true,
				"document": {
					"id": "doc-123",
					"title": "Updated Document",
					"content": "Updated content",
					"url": "https://linear.app/test/doc/doc-123",
					"updatedAt": "2024-01-08T00:00:00.000Z"
				}
			}
		}
	}`

	mockDocumentDeleteResponse = `{
		"data": {
			"documentDelete": {
				"success": true
			}
		}
	}`
)

func defaultHandlers() map[string]string {
	handlers := testutil.DefaultHandlers()
	handlers["ListDocuments"] = mockDocumentsResponse
	handlers["GetDocument"] = mockDocumentResponse
	handlers["CreateDocument"] = mockDocumentCreateResponse
	handlers["UpdateDocument"] = mockDocumentUpdateResponse
	handlers["DeleteDocument"] = mockDocumentDeleteResponse
	return handlers
}
