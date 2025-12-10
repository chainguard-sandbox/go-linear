# go-linear

**Type-safe Go client for the Linear API.** Production-ready with automatic retry, rate limiting, per-operation Prometheus metrics, and multi-tenancy support.

[![Go Reference](https://pkg.go.dev/badge/github.com/eslerm/go-linear.svg)](https://pkg.go.dev/github.com/eslerm/go-linear)
[![Go Report Card](https://goreportcard.com/badge/github.com/eslerm/go-linear)](https://goreportcard.com/report/github.com/eslerm/go-linear)

```bash
go get github.com/eslerm/go-linear
```

---

## Quick Start

**Prerequisites:** Linear API key from https://linear.app/settings/account/security

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/eslerm/go-linear/pkg/linear"
)

func main() {
    client, err := linear.NewClient("lin_api_xxx")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Get authenticated user
    viewer, err := client.Viewer(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Authenticated as: %s\n", viewer.Email)
}
```

**Next:** See [Common Operations](#common-operations) or [Production Setup](#production-setup)

---

## Common Operations

### List Issues with Pagination

**Manual pagination:**
```go
first := int64(50)
issues, err := client.Issues(ctx, &first, nil)
if err != nil {
    return err
}

for _, issue := range issues.Nodes {
    fmt.Printf("%s: %s\n", issue.ID, issue.Title)
}

// Next page
if issues.PageInfo.HasNextPage {
    issues, err = client.Issues(ctx, &first, issues.PageInfo.EndCursor)
}
```

**Automatic pagination with iterator:**
```go
iter := linear.NewIssueIterator(client, 100)
for {
    issue, err := iter.Next(ctx)
    if errors.Is(err, io.EOF) {
        break
    }
    if err != nil {
        return err
    }
    fmt.Printf("%s: %s\n", issue.ID, issue.Title)
}
```

**Thread Safety:** Iterators NOT safe for concurrent use. Create separate iterators per goroutine.

### Create Issue

**Required:** `TeamID` (string)
**Optional:** All fields are pointers (`*string`, `*int64`), `nil` = omit

```go
// Get team ID first
teams, err := client.Teams(ctx, nil, nil)
teamID := teams.Nodes[0].ID

// Create issue
title := "Fix authentication bug"
description := "Users cannot log in on Safari"
priority := int64(1) // 0=none, 1=urgent, 2=high, 3=normal, 4=low

issue, err := client.IssueCreate(ctx, linear.IssueCreateInput{
    TeamID:      teamID,
    Title:       &title,
    Description: &description,
    Priority:    &priority,
})
if err != nil {
    return err
}
fmt.Printf("Created: %s\n", issue.URL)
```

### Update Issue

**All fields optional.** `nil` = unchanged, provide value = update

```go
newTitle := "Fix critical authentication bug"
newPriority := int64(1) // Urgent

updated, err := client.IssueUpdate(ctx, issueID, linear.IssueUpdateInput{
    Title:    &newTitle,
    Priority: &newPriority,
})
```

### Get Single Resource

**Pattern:** Singular method name + ID parameter

```go
issue, err := client.Issue(ctx, "issue-uuid")
team, err := client.Team(ctx, "team-uuid")
user, err := client.User(ctx, "user-uuid")
comment, err := client.Comment(ctx, "comment-uuid")
```

### List Resources

**Pattern:** Plural method name + pagination parameters (`first`, `after`)

```go
first := int64(50)
issues, err := client.Issues(ctx, &first, nil)
teams, err := client.Teams(ctx, &first, nil)
users, err := client.Users(ctx, &first, nil)
comments, err := client.Comments(ctx, &first, nil)
```

---

## Production Setup

**Install:** Copy this configuration for production deployments

```go
package main

import (
    "context"
    "crypto/tls"
    "log/slog"
    "os"
    "time"

    "github.com/eslerm/go-linear/pkg/linear"
)

func main() {
    // Structured logging (JSON for production)
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))

    client, err := linear.NewClient(os.Getenv("LINEAR_API_KEY"),
        // Automatic retry: 5 attempts, 500ms-60s backoff
        linear.WithRetry(5, 500*time.Millisecond, 60*time.Second),

        // Request timeout
        linear.WithTimeout(30*time.Second),

        // Structured logging (all requests/errors)
        linear.WithLogger(logger),

        // Rate limit monitoring (for metrics)
        linear.WithRateLimitCallback(func(info *linear.RateLimitInfo) {
            slog.Info("rate limit status",
                "requests_remaining", info.RequestsRemaining,
                "requests_limit", info.RequestsLimit,
            )
        }),

        // TLS security (enforce TLS 1.2+)
        linear.WithTLSConfig(&tls.Config{
            MinVersion: tls.VersionTLS12,
        }),
    )
    if err != nil {
        logger.Error("client creation failed", "error", err)
        os.Exit(1)
    }
    defer client.Close()

    // Your application code here
}
```

**Production Examples:**
- [examples/production/main.go](examples/production/main.go) - Complete production setup with graceful shutdown
- [examples/prometheus/main.go](examples/prometheus/main.go) - Prometheus metrics integration

**Operational Docs:**
- [docs/RUNBOOK.md](docs/RUNBOOK.md) - Incident response procedures
- [docs/MONITORING.md](docs/MONITORING.md) - Prometheus queries and alerts

---

## API Coverage

### Query Methods (Read Permission)

| Resource | Get Single | List Multiple | Search |
|----------|-----------|---------------|--------|
| **Issues** | `Issue(ctx, id)` | `Issues(ctx, first, after)` | `IssueSearch(ctx, query, first, after)` |
| **Teams** | `Team(ctx, id)` | `Teams(ctx, first, after)` | - |
| **Projects** | `Project(ctx, id)` | `Projects(ctx, first, after)` | - |
| **Users** | `User(ctx, id)` | `Users(ctx, first, after)` | - |
| **Comments** | `Comment(ctx, id)` | `Comments(ctx, first, after)` | - |
| **Labels** | `IssueLabel(ctx, id)` | `IssueLabels(ctx, first, after)` | - |
| **Workflow States** | `WorkflowState(ctx, id)` | `WorkflowStates(ctx, first, after)` | - |
| **Cycles** | `Cycle(ctx, id)` | `Cycles(ctx, first, after)` | - |
| **Roadmaps** | `Roadmap(ctx, id)` | `Roadmaps(ctx, first, after)` | - |
| **Documents** | `Document(ctx, id)` | `Documents(ctx, first, after)` | - |
| **Attachments** | `Attachment(ctx, id)` | `Attachments(ctx, first, after)` | - |
| **Initiatives** | `Initiative(ctx, id)` | `Initiatives(ctx, first, after)` | - |
| **Templates** | `Template(ctx, id)` | `Templates(ctx)` | - |

**Special:** `Viewer(ctx)` - Get authenticated user, `Organization(ctx)` - Get workspace info

### Mutation Methods (Write Permission)

| Resource | Create | Update | Delete |
|----------|--------|--------|--------|
| **Issues** | `IssueCreate(ctx, input)` | `IssueUpdate(ctx, id, input)` | `IssueDelete(ctx, id)` |
| **Comments** | `CommentCreate(ctx, input)` | `CommentUpdate(ctx, id, input)` | `CommentDelete(ctx, id)` |
| **Labels** | `IssueLabelCreate(ctx, input)` | `IssueLabelUpdate(ctx, id, input)` | `IssueLabelDelete(ctx, id)` |
| **Teams** | `TeamCreate(ctx, input)` | `TeamUpdate(ctx, id, input)` | `TeamDelete(ctx, id)` |
| **Projects** | `ProjectCreate(ctx, input)` | `ProjectUpdate(ctx, id, input)` | `ProjectDelete(ctx, id)` |

### Pagination Iterators

Automatic pagination for large result sets:

```go
iter := linear.NewIssueIterator(client, pageSize)
iter := linear.NewTeamIterator(client, pageSize)
iter := linear.NewProjectIterator(client, pageSize)
iter := linear.NewCommentIterator(client, pageSize)
```

**Pattern:**
```go
for {
    item, err := iter.Next(ctx)
    if errors.Is(err, io.EOF) {
        break  // Done
    }
    if err != nil {
        return err
    }
    // Process item
}
```

**Thread Safety:** Iterators NOT safe for concurrent use. Create separate iterators:
```go
// Create separate iterator per goroutine
for i := 0; i < 10; i++ {
    go func() {
        iter := linear.NewIssueIterator(client, 50)  // Separate
        for {
            issue, err := iter.Next(ctx)
            if errors.Is(err, io.EOF) { return }
            if err != nil { return }
            process(issue)
        }
    }()
}
```

---

## Error Handling

**All methods return errors.** Check `err != nil` first.

### Error Types

Use `errors.As()` to check specific error types. Error chains are preserved:

```go
var rateLimitErr *linear.RateLimitError
if errors.As(err, &rateLimitErr) {
    // Rate limited - retry after N seconds
    fmt.Printf("Retry after: %d seconds\n", rateLimitErr.RetryAfter)
    fmt.Printf("Requests remaining: %d/%d\n",
        rateLimitErr.RequestsRemaining,
        rateLimitErr.RequestsLimit)
}

var authErr *linear.AuthenticationError
if errors.As(err, &authErr) {
    // Invalid API key
    return fmt.Errorf("authentication failed: %w", err)
}

var forbiddenErr *linear.ForbiddenError
if errors.As(err, &forbiddenErr) {
    // Missing permission (check API key scopes)
    return fmt.Errorf("permission denied: %w", err)
}

// Check underlying gqlgenc errors
var linearErr *linear.LinearError
if errors.As(err, &linearErr) {
    // Access wrapped error
    underlying := linearErr.Unwrap()
    // Can check underlying error types via errors.Is()
}
```

### Available Error Types

- `LinearError` - Base error with status code, trace ID, wrapped error chain
- `RateLimitError` - Rate limit hit (requests + complexity limits)
- `AuthenticationError` - Invalid/expired API key (401)
- `ForbiddenError` - Missing permission (403)

**Error Chain Preservation:** All errors implement `Unwrap()` for proper `errors.As()` and `errors.Is()` support.

---

## Configuration Options

### Client Options

| Option | Purpose | Example |
|--------|---------|---------|
| `WithTimeout(duration)` | Request timeout | `linear.WithTimeout(30*time.Second)` |
| `WithRetry(max, initial, max backoff)` | Automatic retry | `linear.WithRetry(5, 500*time.Millisecond, 60*time.Second)` |
| `WithMaxRetryDuration(duration)` | Max total retry time | `linear.WithMaxRetryDuration(90*time.Second)` |
| `WithLogger(logger)` | Structured logging | `linear.WithLogger(slog.Default())` |
| `WithMetrics()` | Prometheus metrics | `linear.WithMetrics()` |
| `WithRateLimitCallback(func)` | Monitor rate limits | `linear.WithRateLimitCallback(metricsFunc)` |
| `WithTLSConfig(config)` | TLS settings | `linear.WithTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12})` |
| `WithTransport(transport)` | Custom RoundTripper | `linear.WithTransport(customTransport)` |
| `WithHTTPClient(client)` | Custom HTTP client | `linear.WithHTTPClient(customClient)` |
| `WithBaseURL(url)` | API endpoint | `linear.WithBaseURL("https://api.linear.app/graphql")` |
| `WithUserAgent(ua)` | User agent string | `linear.WithUserAgent("myapp/1.0")` |

### Production Features

**Automatic Retry:**
- Exponential backoff with jitter
- Retries 429 (rate limit) and 5xx errors
- Respects `Retry-After` header
- Context cancellation support
- Bounded total retry time (default: 90s)

**Rate Limiting:**
- Parses `X-RateLimit-*` headers
- Tracks request + complexity limits
- Callback for metrics integration
- Automatic backoff on 429

**Observability:**
- Structured logging with `log/slog`
- Request/response logging with request_id
- Prometheus metrics (RED + rate limits)
- Error logging with operation context

---

## Monitoring & Observability

### Prometheus Metrics

**Single Client (Shared Metrics):**
```go
linear.EnableMetrics()

client, _ := linear.NewClient(apiKey,
    linear.WithMetrics(),
    linear.WithLogger(logger),
)

// Expose metrics at /metrics
http.Handle("/metrics", promhttp.Handler())
http.ListenAndServe(":2112", nil)
```

**Multi-Client (Isolated Metrics per Workspace):**
```go
// Production workspace
prodReg := prometheus.NewRegistry()
prodClient, _ := linear.NewClient(prodKey,
    linear.WithMetricsRegistry(prodReg, "prod"))

// Staging workspace
stageReg := prometheus.NewRegistry()
stageClient, _ := linear.NewClient(stageKey,
    linear.WithMetricsRegistry(stageReg, "staging"))

// Metrics are isolated:
// - linear_prod_requests_total{operation="IssueCreate"}
// - linear_staging_requests_total{operation="IssueCreate"}
```

**Metrics collected (per-operation visibility):**
- `linear_requests_total{operation, status_code}` - Request counts by operation (Viewer, IssueCreate, ListIssues)
- `linear_request_duration_seconds{operation}` - Request latency histogram per operation
- `linear_errors_total{operation, error_type}` - Error counts by operation and type
- `linear_retries_total{reason}` - Retry counts (rate_limited, server_error, network_error)
- `linear_rate_limit_remaining{limit_type}` - Rate limit capacity (requests, complexity)

**Operation names extracted automatically** from GraphQL requests. No more generic "graphql" label - you can see which specific operations are slow or failing.

**See:** [examples/prometheus/main.go](examples/prometheus/main.go) for complete integration

### Recommended Alerts

**Critical (Page):**
- Error rate > 50% for 5 minutes
- All requests failing for 3 minutes

**Warning (Slack):**
- Rate limit < 10% remaining for 5 minutes
- Retry rate > 10% for 10 minutes
- p95 latency > 5s for 10 minutes

**See:** [docs/MONITORING.md](docs/MONITORING.md) for complete alert rules and PromQL queries

### Request Correlation

All requests log `request_id` for correlation with Linear support during incidents:

```json
{
  "level": "INFO",
  "msg": "request completed",
  "method": "POST",
  "url": "https://api.linear.app/graphql",
  "status": 200,
  "duration": "245ms",
  "request_id": "req-abc-123"
}
```

When debugging issues with Linear support, provide the `request_id` from your logs.

**Incident Response:** [docs/RUNBOOK.md](docs/RUNBOOK.md) - Covers common incidents with mitigation steps
**Monitoring Setup:** [docs/MONITORING.md](docs/MONITORING.md) - Complete Prometheus queries and alert rules

---

## Input Types

### Common Patterns

**Required fields:** Non-pointer types
**Optional fields:** Pointer types, `nil` = omit (Create) or unchanged (Update)

### IssueCreateInput

```go
type IssueCreateInput struct {
    TeamID      string    // Required: Team UUID
    Title       *string   // Optional: Issue title
    Description *string   // Optional: Markdown description
    Priority    *int64    // Optional: 0=none, 1=urgent, 2=high, 3=normal, 4=low
    AssigneeID  *string   // Optional: User UUID
    StateID     *string   // Optional: WorkflowState UUID
    LabelIds    []string  // Optional: Label UUIDs
    ParentID    *string   // Optional: Parent issue UUID (for sub-issues)
    DueDate     *string   // Optional: YYYY-MM-DD format
    // ... 20+ more fields in internal/graphql/models.go
}
```

### IssueUpdateInput

```go
type IssueUpdateInput struct {
    Title       *string   // nil = unchanged, value = update
    Description *string
    Priority    *int64
    AssigneeID  *string   // Empty string "" = unassign
    StateID     *string
    LabelIds    []string  // Replace all labels
    AddedLabelIds   []string  // Add these labels
    RemovedLabelIds []string  // Remove these labels
    // ... all fields optional
}
```

**All mutation input types exported:**
- Issue: `IssueCreateInput`, `IssueUpdateInput`
- Comment: `CommentCreateInput`, `CommentUpdateInput`
- Label: `IssueLabelCreateInput`, `IssueLabelUpdateInput`
- Team: `TeamCreateInput`, `TeamUpdateInput`
- Project: `ProjectCreateInput`, `ProjectUpdateInput`

**See:** `pkg/linear/types.go` for complete documentation with field descriptions

---

## Testing

### Run Tests

```bash
# Mock tests (no API key required)
make test

# Live read-only tests (requires LINEAR_API_KEY with Read permission)
LINEAR_API_KEY=lin_api_xxx make test-read

# Live mutation tests (requires Write permission, modifies data)
LINEAR_API_KEY=lin_api_xxx make test-write

# All tests
LINEAR_API_KEY=lin_api_xxx make test-all
```

### Build Tags

Tests use build tags for permission levels:

- `//go:build read` - Read-only tests (Viewer, Issues, Teams)
- `//go:build write` - Mutation tests (IssueCreate, IssueUpdate)
- No tag - Mock tests (httptest.Server, no API key needed)

---

## Development

### Prerequisites

- Go 1.24+
- Linear API key (for live tests)

### Setup

```bash
git clone https://github.com/eslerm/go-linear
cd go-linear
make dev  # Install tools and dependencies
```

### Common Commands

```bash
make help          # Show all commands
make check         # Run all checks (fmt, vet, lint, test)
make fmt           # Format code
make lint          # Run linters
make test          # Run tests
make test-coverage # Generate coverage report
```

### Code Generation

GraphQL client is auto-generated from `schema.graphql`:

```bash
make generate  # Regenerate internal/graphql/*
```

**Config:** See `gqlgenc.yaml` for generation settings

---

## Architecture

```
go-linear/
├── pkg/linear/              # Public API
│   ├── client.go           # Main client (45 methods)
│   ├── errors.go           # Error types
│   ├── transport.go        # Retry/rate limiting
│   ├── options.go          # Client configuration
│   ├── pagination.go       # Iterators
│   └── types.go            # Re-exported input types
├── internal/graphql/       # Generated GraphQL client
│   ├── client.go          # Generated by gqlgenc
│   └── models.go          # Generated types (30k+ lines)
├── queries/                # GraphQL query definitions
│   ├── issues.graphql
│   ├── teams.graphql
│   └── mutations/
└── examples/production/    # Production example code
```

**Design:** Type-safe GraphQL via [genqlient](https://github.com/Khan/genqlient), generated code in `internal/`, clean public API in `pkg/linear/`

---

## FAQ

### How do I get required IDs for mutations?

**TeamID:** `teams, _ := client.Teams(ctx, nil, nil); teamID := teams.Nodes[0].ID`
**UserID:** `viewer, _ := client.Viewer(ctx); userID := viewer.ID`
**StateID:** `states, _ := client.WorkflowStates(ctx, nil, nil); stateID := states.Nodes[0].ID`
**LabelID:** `labels, _ := client.IssueLabels(ctx, nil, nil); labelID := labels.Nodes[0].ID`

### What permissions does my API key need?

**Read:** Query methods (Viewer, Issues, Teams, etc.)
**Write:** Mutation methods (IssueCreate, IssueUpdate, etc.)

Check/configure at: https://linear.app/settings/account/security

### How do I handle rate limits?

**Automatic:** Client retries 429 errors with exponential backoff
**Monitoring:** Use `WithRateLimitCallback` to track limits
**Manual:** Check `RateLimitError.RequestsRemaining` and back off proactively

### Does this support webhooks?

**No.** This is an API client library. For webhooks, use Linear's webhook API directly with your HTTP server.

### How do I contribute?

See development section above. Open issues for bugs/features. PRs welcome with tests.

---

## Resources

- **API Documentation:** https://pkg.go.dev/github.com/eslerm/go-linear
- **Linear API Docs:** https://developers.linear.app
- **GraphQL Schema:** https://studio.apollographql.com/public/Linear-API/variant/current/home
- **Examples:** [examples/production/main.go](examples/production/main.go)
- **Issue Tracker:** https://github.com/eslerm/go-linear/issues

---

## License

Apache 2.0 - See [LICENSE](LICENSE) for details

## Security

For security vulnerabilities, please email [mark.esler@chainguard.dev](mailto:mark.esler@chainguard.dev).

See [SECURITY.md](SECURITY.md) for our vulnerability disclosure policy.

---

## Status

**Pre-1.0:** API may change. Production-ready for use, but breaking changes possible until v1.0.0.

**Current Features:**
- ✅ Type-safe GraphQL operations
- ✅ Automatic retry with exponential backoff
- ✅ Rate limit detection and handling
- ✅ Bounded retry time (prevents request hangs)
- ✅ Structured logging (slog) with request_id correlation
- ✅ **Per-operation Prometheus metrics** (IssueCreate, TeamUpdate, etc.)
- ✅ **Multi-tenancy support** (instance-scoped metrics)
- ✅ **Error chain preservation** (errors.As/Unwrap works)
- ✅ Context support (timeout/cancellation)
- ✅ Comprehensive API coverage (Issues, Teams, Projects, etc.)
- ✅ Automatic pagination iterators (thread-safety documented)
- ✅ Production-ready error handling with operation context
- ✅ **All mutation input types exported** (Team, Project, etc.)
- ✅ TLS configuration
- ✅ Operational documentation (runbook + monitoring guide)
- ✅ 60%+ test coverage (mock + live tests)

**Upstream Sync:** Schema automatically synced from [Linear TypeScript SDK](https://github.com/linear/linear)
