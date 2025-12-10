package linear

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExtractOperationName(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "with operationName",
			body: `{"operationName":"GetIssue","query":"query GetIssue { issue { id } }","variables":{}}`,
			want: "GetIssue",
		},
		{
			name: "without operationName",
			body: `{"query":"query { issues { id } }","variables":{}}`,
			want: "graphql",
		},
		{
			name: "empty operationName",
			body: `{"operationName":"","query":"query { issues { id } }"}`,
			want: "graphql",
		},
		{
			name: "invalid JSON",
			body: `not json`,
			want: "graphql",
		},
		{
			name: "nil body",
			body: "",
			want: "graphql",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest("POST", "http://example.com/graphql",
					bytes.NewReader([]byte(tt.body)))
			} else {
				req = httptest.NewRequest("POST", "http://example.com/graphql", http.NoBody)
			}

			got := extractOperationName(req)
			if got != tt.want {
				t.Errorf("extractOperationName() = %q, want %q", got, tt.want)
			}

			// Verify body is restored
			if tt.body != "" && req.Body != nil {
				restoredBody, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("Failed to read restored body: %v", err)
				}
				if string(restoredBody) != tt.body {
					t.Errorf("Body not restored correctly. got %q, want %q",
						string(restoredBody), tt.body)
				}
			}
		})
	}
}
