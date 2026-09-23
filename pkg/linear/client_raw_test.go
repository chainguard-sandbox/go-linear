package linear

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestClient_Execute_RawData(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Auth header must be set by the inherited interceptor stack.
		if auth := r.Header.Get("Authorization"); auth != "lin_api_test" {
			t.Errorf("Authorization = %q, want %q", auth, "lin_api_test")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"viewer":{"id":"u1","email":"test@example.com"}}}`))
	})

	var data json.RawMessage
	err := client.Execute(context.Background(), `query { viewer { id email } }`, nil, &data)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var got struct {
		Viewer struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"viewer"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal raw data: %v", err)
	}
	if got.Viewer.ID != "u1" || got.Viewer.Email != "test@example.com" {
		t.Errorf("unexpected data: %+v", got)
	}
}

func TestClient_Execute_IntoMap(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Nested field absent from any baked selection set (e.g. project.lead.email).
		_, _ = w.Write([]byte(`{"data":{"project":{"lead":{"email":"lead@example.com"}}}}`))
	})

	out := map[string]any{}
	if err := client.Execute(context.Background(), `query { project(id:"p1") { lead { email } } }`, nil, &out); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	project, _ := out["project"].(map[string]any)
	lead, _ := project["lead"].(map[string]any)
	if lead["email"] != "lead@example.com" {
		t.Errorf("project.lead.email = %v, want lead@example.com", lead["email"])
	}
}

func TestClient_Execute_GraphQLError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[{"message":"Field 'bogus' doesn't exist","path":["viewer","bogus"],"extensions":{"code":"GRAPHQL_VALIDATION_FAILED"}}]}`))
	})

	var data json.RawMessage
	err := client.Execute(context.Background(), `query { viewer { bogus } }`, nil, &data)
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}

	var gqlErr *GraphQLResponseError
	if !errors.As(err, &gqlErr) {
		t.Fatalf("error is not *GraphQLResponseError: %T (%v)", err, err)
	}
	if len(gqlErr.Errors) != 1 {
		t.Fatalf("len(Errors) = %d, want 1", len(gqlErr.Errors))
	}
	e := gqlErr.Errors[0]
	if !strings.Contains(e.Message, "doesn't exist") {
		t.Errorf("Message = %q, want it to mention the field", e.Message)
	}
	if len(e.Path) != 2 {
		t.Errorf("Path = %v, want 2 elements", e.Path)
	}
	if e.Extensions["code"] != "GRAPHQL_VALIDATION_FAILED" {
		t.Errorf("Extensions[code] = %v, want GRAPHQL_VALIDATION_FAILED", e.Extensions["code"])
	}
	// The rendered error must mention the underlying message.
	if !strings.Contains(err.Error(), "doesn't exist") {
		t.Errorf("Error() = %q, want it to include the message", err.Error())
	}
}

func TestClient_Execute_VarsReachRequestBody(t *testing.T) {
	var gotBody atomic.Value
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody.Store(string(b))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issue":{"id":"iss-1"}}}`))
	})

	vars := map[string]any{"id": "iss-1", "first": 10, "active": true}
	var data json.RawMessage
	if err := client.Execute(context.Background(), `query($id:String!){ issue(id:$id){ id } }`, vars, &data); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	body, _ := gotBody.Load().(string)
	var sent struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables"`
	}
	if err := json.Unmarshal([]byte(body), &sent); err != nil {
		t.Fatalf("request body not valid JSON: %v (%s)", err, body)
	}
	if sent.Variables["id"] != "iss-1" {
		t.Errorf("variables.id = %v, want iss-1", sent.Variables["id"])
	}
	if sent.Variables["active"] != true {
		t.Errorf("variables.active = %v, want true", sent.Variables["active"])
	}
	// json numbers decode as float64.
	if sent.Variables["first"] != float64(10) {
		t.Errorf("variables.first = %v, want 10", sent.Variables["first"])
	}
	if !strings.Contains(sent.Query, "issue(id:$id)") {
		t.Errorf("query not sent verbatim: %q", sent.Query)
	}
}

func TestClient_Execute_OperationNameOnWire(t *testing.T) {
	var gotBody atomic.Value
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody.Store(string(b))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"viewer":{"id":"u1"}}}`))
	})

	readOpName := func() (string, bool) {
		var sent map[string]json.RawMessage
		_ = json.Unmarshal([]byte(gotBody.Load().(string)), &sent)
		raw, ok := sent["operationName"]
		if !ok {
			return "", false
		}
		var s string
		_ = json.Unmarshal(raw, &s)
		return s, true
	}

	// With WithOperationName, the selected name is sent verbatim.
	var data json.RawMessage
	err := client.Execute(context.Background(),
		`query A { viewer { id } } query B { viewer { name } }`, nil, &data,
		WithOperationName("B"))
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if name, ok := readOpName(); !ok || name != "B" {
		t.Errorf("operationName on wire = %q (present=%v), want %q", name, ok, "B")
	}

	// Without the option, operationName must be ABSENT from the wire body (not
	// present-but-empty), so the server runs the sole operation.
	err = client.Execute(context.Background(), `query { viewer { id } }`, nil, &data)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if name, ok := readOpName(); ok {
		t.Errorf("operationName present on wire = %q, want the key absent", name)
	}
}

func TestClient_Execute_GraphQLErrorWithStatusToken(t *testing.T) {
	// A well-formed 200 GraphQL error whose extensions.code contains a status
	// token ("FORBIDDEN") must still surface as *GraphQLResponseError with its
	// structured detail — not be misclassified as *ForbiddenError.
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":null,"errors":[{"message":"Access denied to entity","extensions":{"code":"FORBIDDEN","type":"authorization"}}]}`))
	})

	var data json.RawMessage
	err := client.Execute(context.Background(), `query { issue(id:"x") { id } }`, nil, &data)
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}

	var forbidden *ForbiddenError
	if errors.As(err, &forbidden) {
		t.Fatalf("200 GraphQL error misclassified as *ForbiddenError: %v", err)
	}
	var gqlErr *GraphQLResponseError
	if !errors.As(err, &gqlErr) {
		t.Fatalf("error is not *GraphQLResponseError: %T (%v)", err, err)
	}
	if len(gqlErr.Errors) != 1 || gqlErr.Errors[0].Extensions["code"] != "FORBIDDEN" {
		t.Errorf("structured detail lost: %+v", gqlErr.Errors)
	}
}

func TestClient_Execute_WithMaxResponseBytes(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"x":"` + strings.Repeat("a", 4096) + `"}}`))
	})

	var data json.RawMessage
	err := client.Execute(context.Background(), `query { x }`, nil, &data, WithMaxResponseBytes(1024))
	if err == nil {
		t.Fatal("Execute() expected an oversize error with a 1KB cap, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("error = %q, want it to mention the size limit", err.Error())
	}
}

func TestClient_Execute_5xxWithForbiddenTokenNotMisclassified(t *testing.T) {
	// A 500 whose body happens to contain "FORBIDDEN" must NOT be classified as
	// *ForbiddenError (a 403). Classification is by HTTP status code, not by
	// string-matching the embedded body. The retry transport exhausts on 500,
	// so this reaches Execute as a network error.
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errors":[{"message":"upstream boom","extensions":{"code":"FORBIDDEN"}}]}`))
	})

	var data json.RawMessage
	err := client.Execute(context.Background(), `query { viewer { id } }`, nil, &data)
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	var forbidden *ForbiddenError
	if errors.As(err, &forbidden) {
		t.Fatalf("500 with FORBIDDEN in body misclassified as *ForbiddenError: %v", err)
	}
	// It should surface the structured GraphQL errors (non-auth non-2xx with errors[]).
	var gqlErr *GraphQLResponseError
	if !errors.As(err, &gqlErr) {
		t.Fatalf("want *GraphQLResponseError, got %T (%v)", err, err)
	}
}

func TestClient_Execute_RetryOn5xx(t *testing.T) {
	var mu sync.Mutex
	var bodies []string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, string(b))
		n := len(bodies)
		mu.Unlock()
		if n == 1 {
			// First attempt fails with a retryable status.
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"errors":[{"message":"internal"}]}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"viewer":{"id":"u1"}}}`))
	})

	var data json.RawMessage
	if err := client.Execute(context.Background(), `query { viewer { id } }`, nil, &data); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(bodies) < 2 {
		t.Fatalf("server calls = %d, want >= 2 (retry should have fired)", len(bodies))
	}
	// The retried request must resend the full body, not an empty one (guards
	// against request-body-reuse regressions across retries).
	if bodies[1] == "" || bodies[1] != bodies[0] {
		t.Errorf("retry body = %q, want it to equal the first body %q", bodies[1], bodies[0])
	}
}

func TestClient_Execute_AuthErrorClassified(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":[{"message":"unauthorized","extensions":{"code":401}}]}`))
	})

	var data json.RawMessage
	err := client.Execute(context.Background(), `query { viewer { id } }`, nil, &data)
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	var authErr *AuthenticationError
	if !errors.As(err, &authErr) {
		t.Fatalf("error is not *AuthenticationError: %T (%v)", err, err)
	}
}
