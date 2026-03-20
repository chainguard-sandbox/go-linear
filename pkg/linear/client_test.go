package linear

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := NewClient("lin_api_test",
		WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	return client
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		opts    []Option
		wantErr bool
	}{
		{
			name:    "valid api key",
			apiKey:  "lin_api_test123",
			wantErr: false,
		},
		{
			name:    "empty api key",
			apiKey:  "",
			wantErr: true,
		},
		{
			name:   "with custom timeout",
			apiKey: "lin_api_test123",
			opts: []Option{
				WithTimeout(60 * time.Second),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.apiKey, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client without error")
			}
		})
	}
}

func TestClient_Viewer(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		if auth := r.Header.Get("Authorization"); auth != "lin_api_test" {
			t.Errorf("Authorization = %q, want %q", auth, "lin_api_test")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"viewer":{"id":"u1","name":"Test","email":"test@example.com","displayName":"test","createdAt":"2024-01-01T00:00:00.000Z","admin":false,"active":true,"avatarUrl":""}}}`))
	})

	viewer, err := client.Viewer(context.Background())
	if err != nil {
		t.Fatalf("Viewer() error = %v", err)
	}

	if viewer.ID != "u1" {
		t.Errorf("ID = %q, want %q", viewer.ID, "u1")
	}
	if viewer.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", viewer.Email, "test@example.com")
	}
}

func TestClient_Close(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Should not error on first call
	if err := client.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}

	// Should be idempotent (safe to call multiple times)
	if err := client.Close(); err != nil {
		t.Errorf("Close() second call error = %v, want nil", err)
	}

	// Should be safe to call even after close
	if err := client.Close(); err != nil {
		t.Errorf("Close() third call error = %v, want nil", err)
	}
}

func TestClientOptions(t *testing.T) {
	// Test WithHTTPClient
	t.Run("WithHTTPClient", func(t *testing.T) {
		customClient := &http.Client{Timeout: 5 * time.Second}
		client, err := NewClient("lin_api_test", WithHTTPClient(customClient))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
	})

	// Test WithUserAgent
	t.Run("WithUserAgent", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ua := r.Header.Get("User-Agent")
			if ua != "custom-agent/1.0" {
				t.Errorf("User-Agent = %q, want %q", ua, "custom-agent/1.0")
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"viewer":{"id":"u1","name":"Test","email":"test@example.com","displayName":"test","createdAt":"2024-01-01T00:00:00.000Z","admin":false,"active":true,"avatarUrl":""}}}`))
		}))
		defer server.Close()

		client, err := NewClient("lin_api_test",
			WithBaseURL(server.URL),
			WithUserAgent("custom-agent/1.0"),
		)
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()

		_, err = client.Viewer(context.Background())
		if err != nil {
			t.Errorf("Viewer() error = %v", err)
		}
	})

	// Test WithLogger
	t.Run("WithLogger", func(t *testing.T) {
		logger := NewLogger()
		client, err := NewClient("lin_api_test", WithLogger(logger))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
	})

	// Test WithRetry
	t.Run("WithRetry", func(t *testing.T) {
		client, err := NewClient("lin_api_test", WithRetry(5, 100*time.Millisecond, 1*time.Second))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
	})

	// Test WithRateLimitCallback
	t.Run("WithRateLimitCallback", func(t *testing.T) {
		callback := func(info *RateLimitInfo) {}
		client, err := NewClient("lin_api_test", WithRateLimitCallback(callback))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
	})

	// Test WithMaxRetryDuration
	t.Run("WithMaxRetryDuration", func(t *testing.T) {
		client, err := NewClient("lin_api_test", WithMaxRetryDuration(5*time.Minute))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
	})

	// Test WithCircuitBreaker
	t.Run("WithCircuitBreaker", func(t *testing.T) {
		cb := &CircuitBreaker{
			MaxFailures:  5,
			ResetTimeout: 30 * time.Second,
		}
		client, err := NewClient("lin_api_test", WithCircuitBreaker(cb))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
	})

	// Test WithTransport
	t.Run("WithTransport", func(t *testing.T) {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		client, err := NewClient("lin_api_test", WithTransport(transport))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
	})

	// Test WithMetrics
	t.Run("WithMetrics", func(t *testing.T) {
		client, err := NewClient("lin_api_test", WithMetrics())
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
	})

	// Test WithTracing
	t.Run("WithTracing", func(t *testing.T) {
		client, err := NewClient("lin_api_test", WithTracing())
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer client.Close()
	})
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		substr   string
		expected bool
	}{
		{"contains substring", "hello world", "world", true},
		{"does not contain", "hello world", "foo", false},
		{"empty substring", "hello", "", true},
		{"case sensitive", "Hello", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.str, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.str, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"401 error", errors.New("HTTP 401 Unauthorized"), true},
		{"authentication error", errors.New("authentication failed"), true},
		{"unauthorized error", errors.New("unauthorized access"), true},
		{"random error", errors.New("Something went wrong"), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAuthError(tt.err)
			if result != tt.expected {
				t.Errorf("isAuthError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestClient_AttachmentMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"attachment":{"id":"att-1","title":"Test Attachment","url":"https://example.com/file.pdf","createdAt":"2024-01-01T00:00:00.000Z"}}}`))
	})

	attachment, err := client.Attachment(context.Background(), "att-1")
	if err != nil {
		t.Fatalf("Attachment() error = %v", err)
	}
	if attachment.ID != "att-1" {
		t.Errorf("Attachment().ID = %q, want att-1", attachment.ID)
	}
}

func TestClient_AttachmentsMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"attachments":{"nodes":[{"id":"att-1","title":"Test","url":"https://example.com","createdAt":"2024-01-01T00:00:00.000Z"}],"pageInfo":{"hasNextPage":false}}}}`))
	})

	result, err := client.Attachments(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Attachments() error = %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Errorf("len(Attachments().Nodes) = %d, want 1", len(result.Nodes))
	}
}

func TestClient_CycleMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"cycle":{"id":"cyc-1","name":"Sprint 1","startsAt":"2024-01-01T00:00:00.000Z","endsAt":"2024-01-14T00:00:00.000Z","createdAt":"2024-01-01T00:00:00.000Z"}}}`))
	})

	cycle, err := client.Cycle(context.Background(), "cyc-1")
	if err != nil {
		t.Fatalf("Cycle() error = %v", err)
	}
	if cycle.ID != "cyc-1" {
		t.Errorf("Cycle().ID = %q, want cyc-1", cycle.ID)
	}
}

func TestClient_CyclesMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"cycles":{"nodes":[{"id":"cyc-1","name":"Sprint 1","startsAt":"2024-01-01T00:00:00.000Z","endsAt":"2024-01-14T00:00:00.000Z","createdAt":"2024-01-01T00:00:00.000Z"}],"pageInfo":{"hasNextPage":false}}}}`))
	})

	result, err := client.Cycles(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Cycles() error = %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Errorf("len(Cycles().Nodes) = %d, want 1", len(result.Nodes))
	}
}

func TestClient_InitiativeMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"initiative":{"id":"init-1","name":"Q1 Goals","description":"Q1 2024 initiatives","createdAt":"2024-01-01T00:00:00.000Z"}}}`))
	})

	init, err := client.Initiative(context.Background(), "init-1")
	if err != nil {
		t.Fatalf("Initiative() error = %v", err)
	}
	if init.ID != "init-1" {
		t.Errorf("Initiative().ID = %q, want init-1", init.ID)
	}
}

func TestClient_InitiativesMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"initiatives":{"nodes":[{"id":"init-1","name":"Q1 Goals","description":"","createdAt":"2024-01-01T00:00:00.000Z"}],"pageInfo":{"hasNextPage":false}}}}`))
	})

	result, err := client.Initiatives(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Initiatives() error = %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Errorf("len(Initiatives().Nodes) = %d, want 1", len(result.Nodes))
	}
}

func TestClient_ProjectMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"project":{"id":"proj-1","name":"Platform Redesign","description":"Redesigning platform","createdAt":"2024-01-01T00:00:00.000Z"}}}`))
	})

	project, err := client.Project(context.Background(), "proj-1")
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}
	if project.ID != "proj-1" {
		t.Errorf("Project().ID = %q, want proj-1", project.ID)
	}
}

func TestClient_ProjectsMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"projects":{"nodes":[{"id":"proj-1","name":"Platform","description":"","createdAt":"2024-01-01T00:00:00.000Z"}],"pageInfo":{"hasNextPage":false}}}}`))
	})

	result, err := client.Projects(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Projects() error = %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Errorf("len(Projects().Nodes) = %d, want 1", len(result.Nodes))
	}
}

func TestClient_RoadmapMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"roadmap":{"id":"road-1","name":"2024 Roadmap","description":"","createdAt":"2024-01-01T00:00:00.000Z"}}}`))
	})

	roadmap, err := client.Roadmap(context.Background(), "road-1")
	if err != nil {
		t.Fatalf("Roadmap() error = %v", err)
	}
	if roadmap.ID != "road-1" {
		t.Errorf("Roadmap().ID = %q, want road-1", roadmap.ID)
	}
}

func TestClient_RoadmapsMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"roadmaps":{"nodes":[{"id":"road-1","name":"2024 Roadmap","description":"","createdAt":"2024-01-01T00:00:00.000Z"}],"pageInfo":{"hasNextPage":false}}}}`))
	})

	result, err := client.Roadmaps(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Roadmaps() error = %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Errorf("len(Roadmaps().Nodes) = %d, want 1", len(result.Nodes))
	}
}

func TestClient_TemplateMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"template":{"id":"tmpl-1","name":"Bug Report","description":"","createdAt":"2024-01-01T00:00:00.000Z"}}}`))
	})

	template, err := client.Template(context.Background(), "tmpl-1")
	if err != nil {
		t.Fatalf("Template() error = %v", err)
	}
	if template.ID != "tmpl-1" {
		t.Errorf("Template().ID = %q, want tmpl-1", template.ID)
	}
}

func TestClient_TemplatesMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"templates":[{"id":"tmpl-1","name":"Bug Report","description":"","createdAt":"2024-01-01T00:00:00.000Z"}]}}`))
	})

	result, err := client.Templates(context.Background())
	if err != nil {
		t.Fatalf("Templates() error = %v", err)
	}
	if len(result) != 1 {
		t.Errorf("len(Templates()) = %d, want 1", len(result))
	}
}

func TestClient_IssuesFilteredMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issues":{"nodes":[{"id":"iss-1","title":"Bug fix","identifier":"ENG-123","url":"https://linear.app","createdAt":"2024-01-01T00:00:00.000Z"}],"pageInfo":{"hasNextPage":false}}}}`))
	})

	result, err := client.IssuesFiltered(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("IssuesFiltered() error = %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Errorf("len(IssuesFiltered().Nodes) = %d, want 1", len(result.Nodes))
	}
}

func TestClient_AttachmentsFilteredMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"attachments":{"nodes":[{"id":"att-1","title":"Doc","url":"https://example.com","createdAt":"2024-01-01T00:00:00.000Z"}],"pageInfo":{"hasNextPage":false}}}}`))
	})

	result, err := client.AttachmentsFiltered(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("AttachmentsFiltered() error = %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Errorf("len(AttachmentsFiltered().Nodes) = %d, want 1", len(result.Nodes))
	}
}

func TestClient_CommentsFilteredMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"comments":{"nodes":[{"id":"com-1","body":"Test comment","createdAt":"2024-01-01T00:00:00.000Z","url":"https://linear.app"}],"pageInfo":{"hasNextPage":false}}}}`))
	})

	result, err := client.CommentsFiltered(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("CommentsFiltered() error = %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Errorf("len(CommentsFiltered().Nodes) = %d, want 1", len(result.Nodes))
	}
}

func TestClient_CyclesFilteredMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"cycles":{"nodes":[{"id":"cyc-1","name":"Sprint 1","startsAt":"2024-01-01T00:00:00.000Z","endsAt":"2024-01-14T00:00:00.000Z","createdAt":"2024-01-01T00:00:00.000Z"}],"pageInfo":{"hasNextPage":false}}}}`))
	})

	result, err := client.CyclesFiltered(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("CyclesFiltered() error = %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Errorf("len(CyclesFiltered().Nodes) = %d, want 1", len(result.Nodes))
	}
}

func TestClient_DocumentsFilteredMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"documents":{"nodes":[{"id":"doc-1","title":"Design Doc","content":"Content","createdAt":"2024-01-01T00:00:00.000Z"}],"pageInfo":{"hasNextPage":false}}}}`))
	})

	result, err := client.DocumentsFiltered(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("DocumentsFiltered() error = %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Errorf("len(DocumentsFiltered().Nodes) = %d, want 1", len(result.Nodes))
	}
}

func TestClient_InitiativesFilteredMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"initiatives":{"nodes":[{"id":"init-1","name":"Q1 Goals","description":"","createdAt":"2024-01-01T00:00:00.000Z"}],"pageInfo":{"hasNextPage":false}}}}`))
	})

	result, err := client.InitiativesFiltered(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("InitiativesFiltered() error = %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Errorf("len(InitiativesFiltered().Nodes) = %d, want 1", len(result.Nodes))
	}
}

func TestClient_IssueCreateMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueCreate":{"success":true,"issue":{"id":"iss-new","title":"New Issue","identifier":"ENG-999","url":"https://linear.app","createdAt":"2024-01-01T00:00:00.000Z"}}}}`))
	})

	title := "New Issue"
	result, err := client.IssueCreate(context.Background(), intgraphql.IssueCreateInput{
		TeamID: "team-1",
		Title:  &title,
	})
	if err != nil {
		t.Fatalf("IssueCreate() error = %v", err)
	}
	if result.ID != "iss-new" {
		t.Errorf("IssueCreate().ID = %q, want iss-new", result.ID)
	}
}

func TestClient_IssueDeleteMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueDelete":{"success":true}}}`))
	})

	err := client.IssueDelete(context.Background(), "iss-1", nil)
	if err != nil {
		t.Fatalf("IssueDelete() error = %v", err)
	}
}

func TestClient_IssueBatchUpdateMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueBatchUpdate":{"success":true,"issues":[{"id":"iss-1"},{"id":"iss-2"}]}}}`))
	})

	result, err := client.IssueBatchUpdate(context.Background(), []string{"iss-1", "iss-2"}, intgraphql.IssueUpdateInput{})
	if err != nil {
		t.Fatalf("IssueBatchUpdate() error = %v", err)
	}
	if result == nil {
		t.Error("IssueBatchUpdate() returned nil")
	}
}

func TestClient_CommentCreateMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"commentCreate":{"success":true,"comment":{"id":"com-1","body":"Test","createdAt":"2024-01-01T00:00:00.000Z","url":"https://linear.app"}}}}`))
	})

	body := "Test"
	issueID := "iss-1"
	result, err := client.CommentCreate(context.Background(), intgraphql.CommentCreateInput{
		IssueID: &issueID,
		Body:    &body,
	})
	if err != nil {
		t.Fatalf("CommentCreate() error = %v", err)
	}
	if result.ID != "com-1" {
		t.Errorf("CommentCreate().ID = %q, want com-1", result.ID)
	}
}

func TestClient_CommentDeleteMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"commentDelete":{"success":true}}}`))
	})

	err := client.CommentDelete(context.Background(), "com-1")
	if err != nil {
		t.Fatalf("CommentDelete() error = %v", err)
	}
}

func TestClient_CycleArchiveMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"cycleArchive":{"success":true}}}`))
	})

	err := client.CycleArchive(context.Background(), "cyc-1")
	if err != nil {
		t.Fatalf("CycleArchive() error = %v", err)
	}
}

func TestClient_AttachmentDeleteMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"attachmentDelete":{"success":true}}}`))
	})

	err := client.AttachmentDelete(context.Background(), "att-1")
	if err != nil {
		t.Fatalf("AttachmentDelete() error = %v", err)
	}
}

func TestClient_FavoriteCreateMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"favoriteCreate":{"success":true,"favorite":{"id":"fav-1"}}}}`))
	})

	issueID := "iss-1"
	result, err := client.FavoriteCreate(context.Background(), intgraphql.FavoriteCreateInput{
		IssueID: &issueID,
	})
	if err != nil {
		t.Fatalf("FavoriteCreate() error = %v", err)
	}
	if result.ID != "fav-1" {
		t.Errorf("FavoriteCreate().ID = %q, want fav-1", result.ID)
	}
}

func TestClient_FavoriteDeleteMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"favoriteDelete":{"success":true}}}`))
	})

	err := client.FavoriteDelete(context.Background(), "fav-1")
	if err != nil {
		t.Fatalf("FavoriteDelete() error = %v", err)
	}
}

func TestClient_InitiativeCreateMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"initiativeCreate":{"success":true,"initiative":{"id":"init-1","name":"New Initiative","description":"","createdAt":"2024-01-01T00:00:00.000Z"}}}}`))
	})

	result, err := client.InitiativeCreate(context.Background(), intgraphql.InitiativeCreateInput{
		Name: "New Initiative",
	})
	if err != nil {
		t.Fatalf("InitiativeCreate() error = %v", err)
	}
	if result.ID != "init-1" {
		t.Errorf("InitiativeCreate().ID = %q, want init-1", result.ID)
	}
}

func TestClient_ProjectsFilteredMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"projects":{"nodes":[{"id":"proj-1","name":"Project A","description":"","createdAt":"2024-01-01T00:00:00.000Z","color":"#ff0000","state":"started"}],"pageInfo":{"hasNextPage":false}}}}`))
	})

	result, err := client.ProjectsFiltered(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("ProjectsFiltered() error = %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Errorf("len(ProjectsFiltered().Nodes) = %d, want 1", len(result.Nodes))
	}
}

func TestClient_TeamsFilteredMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"teams":{"nodes":[{"id":"team-1","name":"Engineering","key":"ENG","description":"","icon":"","createdAt":"2024-01-01T00:00:00.000Z","color":"#0000ff","private":false}],"pageInfo":{"hasNextPage":false}}}}`))
	})

	result, err := client.TeamsFiltered(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("TeamsFiltered() error = %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Errorf("len(TeamsFiltered().Nodes) = %d, want 1", len(result.Nodes))
	}
}

func TestClient_UsersFilteredMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"users":{"nodes":[{"id":"user-1","name":"Alice","displayName":"alice","email":"alice@example.com","active":true,"avatarUrl":"","admin":false}],"pageInfo":{"hasNextPage":false}}}}`))
	})

	result, err := client.UsersFiltered(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("UsersFiltered() error = %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Errorf("len(UsersFiltered().Nodes) = %d, want 1", len(result.Nodes))
	}
}

func TestClient_WorkflowStatesFilteredMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"workflowStates":{"nodes":[{"id":"state-1","name":"In Progress","type":"started","color":"#ffcc00","position":1.0}],"pageInfo":{"hasNextPage":false}}}}`))
	})

	result, err := client.WorkflowStatesFiltered(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("WorkflowStatesFiltered() error = %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Errorf("len(WorkflowStatesFiltered().Nodes) = %d, want 1", len(result.Nodes))
	}
}

func TestClient_IssueLabelsFilteredMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueLabels":{"nodes":[{"id":"label-1","name":"bug","color":"#ff0000","createdAt":"2024-01-01T00:00:00.000Z","description":"Bug reports"}],"pageInfo":{"hasNextPage":false}}}}`))
	})

	result, err := client.IssueLabelsFiltered(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("IssueLabelsFiltered() error = %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Errorf("len(IssueLabelsFiltered().Nodes) = %d, want 1", len(result.Nodes))
	}
}

func TestClient_ReactionCreateMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"reactionCreate":{"success":true,"reaction":{"id":"reaction-1","emoji":"👍"}}}}`))
	})

	commentID := "comment-1"
	result, err := client.ReactionCreate(context.Background(), intgraphql.ReactionCreateInput{
		CommentID: &commentID,
		Emoji:     "👍",
	})
	if err != nil {
		t.Fatalf("ReactionCreate() error = %v", err)
	}
	if result.ID != "reaction-1" {
		t.Errorf("ReactionCreate().ID = %q, want reaction-1", result.ID)
	}
}

func TestClient_ReactionDeleteMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"reactionDelete":{"success":true}}}`))
	})

	err := client.ReactionDelete(context.Background(), "reaction-1")
	if err != nil {
		t.Fatalf("ReactionDelete() error = %v", err)
	}
}

func TestClient_NotificationArchiveMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"notificationArchive":{"success":true}}}`))
	})

	err := client.NotificationArchive(context.Background(), "notif-1")
	if err != nil {
		t.Fatalf("NotificationArchive() error = %v", err)
	}
}

func TestClient_NotificationSubscriptionDeleteMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"notificationSubscriptionDelete":{"success":true}}}`))
	})

	err := client.NotificationSubscriptionDelete(context.Background(), "sub-1")
	if err != nil {
		t.Fatalf("NotificationSubscriptionDelete() error = %v", err)
	}
}

func TestClient_ProjectMilestoneDeleteMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"projectMilestoneDelete":{"success":true}}}`))
	})

	err := client.ProjectMilestoneDelete(context.Background(), "milestone-1")
	if err != nil {
		t.Fatalf("ProjectMilestoneDelete() error = %v", err)
	}
}

func TestClient_IssueRelationDeleteMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueRelationDelete":{"success":true}}}`))
	})

	err := client.IssueRelationDelete(context.Background(), "rel-1")
	if err != nil {
		t.Fatalf("IssueRelationDelete() error = %v", err)
	}
}

func TestClient_IssueUpdateNullableMock(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueUpdate":{"success":true,"issue":{"id":"iss-1","identifier":"ENG-123","title":"Updated","description":"","priority":1,"parent":null,"cycle":null,"project":null,"state":{"id":"state-1","name":"In Progress"},"updatedAt":"2024-01-01T00:00:00.000Z"}}}}`))
	})

	title := "Updated"
	result, err := client.IssueUpdateNullable(context.Background(), "iss-1", IssueUpdateNullableInput{
		Title:    &title,
		ParentID: NewNull[string](), // Explicitly remove parent
	})
	if err != nil {
		t.Fatalf("IssueUpdateNullable() error = %v", err)
	}
	if result.ID != "iss-1" {
		t.Errorf("IssueUpdateNullable().ID = %q, want iss-1", result.ID)
	}
}

func TestClient_IssueUpdateNullable_GraphQLError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[{"message":"Issue not found"}]}`))
	})

	title := "Updated"
	_, err := client.IssueUpdateNullable(context.Background(), "invalid", IssueUpdateNullableInput{
		Title: &title,
	})
	if err == nil {
		t.Fatal("Expected error for GraphQL error response")
	}
}

func TestClient_IssueUpdateNullable_Failure(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"issueUpdate":{"success":false,"issue":null}}}`))
	})

	title := "Updated"
	_, err := client.IssueUpdateNullable(context.Background(), "iss-1", IssueUpdateNullableInput{
		Title: &title,
	})
	if err == nil {
		t.Fatal("Expected error for failed mutation")
	}
}

// testCredentialProvider implements CredentialProvider for testing.
type testCredentialProvider struct {
	apiKey string
}

func (p *testCredentialProvider) GetCredential(_ context.Context) (string, error) {
	return p.apiKey, nil
}

func TestClient_WithCredentialProvider(t *testing.T) {
	provider := &testCredentialProvider{apiKey: "lin_api_dynamic"}

	client, err := NewClient("lin_api_initial", WithCredentialProvider(provider))
	if err != nil {
		t.Fatalf("NewClient with WithCredentialProvider() error = %v", err)
	}

	// Verify credential provider is set
	if client.config.CredentialProvider == nil {
		t.Error("CredentialProvider should be set")
	}
}
