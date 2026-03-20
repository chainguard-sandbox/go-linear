# Go SDK Documentation

Production-ready Go client for the Linear GraphQL API.

```bash
go get github.com/chainguard-sandbox/go-linear/v2
```

---

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

func main() {
    client, err := linear.NewClient("lin_api_xxx")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    viewer, err := client.Viewer(ctx)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Authenticated as: %s", viewer.Email)

    // Create issue
    title := "Payment processing timeout"
    priority := int64(1) // Urgent

    teams, _ := client.Teams(ctx, nil, nil)
    issue, err := client.IssueCreate(ctx, linear.IssueCreateInput{
        TeamID:   teams.Nodes[0].ID,
        Title:    &title,
        Priority: &priority,
    })
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Created: %s", issue.URL)
}
```

---

## Production Configuration

```go
import (
    "context"
    "crypto/tls"
    "os"
    "time"

    "github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// Use the convenience logger or create your own with clog.New()
logger := linear.NewLogger()
ctx := context.Background()

client, err := linear.NewClient(os.Getenv("LINEAR_API_KEY"),
    // Retry configuration
    linear.WithRetry(5, 1*time.Second, 30*time.Second),
    linear.WithTimeout(30*time.Second),

    // Circuit breaker: fail-fast after 5 consecutive failures
    linear.WithCircuitBreaker(&linear.CircuitBreaker{
        MaxFailures:  5,
        ResetTimeout: 60*time.Second,
    }),

    // Security
    linear.WithTLSConfig(&tls.Config{
        MinVersion: tls.VersionTLS12,
    }),

    // Observability
    linear.WithLogger(logger),
    linear.WithMetrics(),
    linear.WithRateLimitCallback(func(info *linear.RateLimitInfo) {
        logger.InfoContext(ctx, "rate_limit",
            "requests_remaining", info.RequestsRemaining,
            "requests_limit", info.RequestsLimit,
        )
    }),
)
```

---

## Client Options

| Option | Default | Description |
|--------|---------|-------------|
| `WithTimeout(d)` | 30s | Request timeout |
| `WithRetry(max, initial, maxBackoff)` | 3, 1s, 30s | Exponential backoff |
| `WithMaxRetryDuration(d)` | 90s | Total retry limit |
| `WithCircuitBreaker(config)` | nil | Fail-fast config |
| `WithLogger(logger)` | nil | clog logger |
| `WithMetrics()` | disabled | Prometheus metrics |
| `WithMetricsRegistry(reg, ns)` | global, "linear" | Multi-tenant metrics |
| `WithRateLimitCallback(f)` | nil | Rate limit monitoring |
| `WithTLSConfig(config)` | TLS 1.2+ | TLS settings |
| `WithCredentialProvider(p)` | static | Experimental rotation |
| `WithBaseURL(url)` | api.linear.app | Custom endpoint |

---

## Error Handling

```go
_, err := client.IssueCreate(ctx, input)
if err != nil {
    var rateLimitErr *linear.RateLimitError
    if errors.As(err, &rateLimitErr) {
        log.Printf("Rate limited. Retry after %ds", rateLimitErr.RetryAfter)
        // Client already retried with backoff
    }

    var authErr *linear.AuthenticationError
    if errors.As(err, &authErr) {
        // Invalid API key (401)
    }

    var forbiddenErr *linear.ForbiddenError
    if errors.As(err, &forbiddenErr) {
        // Permission denied (403)
    }
}
```

All errors preserve underlying errors for `errors.Unwrap()`.

---

## Transport Behavior

### Retry

- Retries: 429 (rate limit), 5xx (server errors), timeouts
- No retry: 4xx (except 429), 2xx responses
- Respects `Retry-After` header
- Bounded total time (90s default)

### Circuit Breaker

- Opens after N failures (default: 5)
- Half-open after timeout (default: 60s)
- Closes on successful request

### Rate Limiting

- Parses `X-RateLimit-*` headers
- Tracks request and complexity limits
- Automatic backoff on 429

---

## Pagination

### Manual

```go
first := int64(50)
issues, _ := client.Issues(ctx, &first, nil)

for _, issue := range issues.Nodes {
    fmt.Printf("%s: %s\n", issue.ID, issue.Title)
}

if issues.PageInfo.HasNextPage {
    issues, _ = client.Issues(ctx, &first, issues.PageInfo.EndCursor)
}
```

### Automatic (Thread-Safe)

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
    process(issue)
}

// Concurrent iteration safe (mutex-protected)
for i := 0; i < 10; i++ {
    go func() {
        for {
            issue, _ := iter.Next(ctx)  // Safe
            process(issue)
        }
    }()
}
```

---

## Method Signatures

### Query Methods

All query methods take `context.Context` as the first parameter and return `(result, error)`.

**Get Single Resource:**
```go
func (c *Client) Issue(ctx context.Context, id string) (*Issue, error)
func (c *Client) Team(ctx context.Context, id string) (*Team, error)
func (c *Client) Project(ctx context.Context, id string) (*Project, error)
```

**Parameters:**
- `id`: Resource identifier (string) - can be user-facing ID (e.g., "ENG-123") or UUID
- Returns `nil` for both result and error when resource not found (not an error)
- Returns error only for API failures, authentication issues, or network errors

**List Resources (Paginated):**
```go
func (c *Client) Issues(ctx context.Context, first *int64, after *string) (*IssueConnection, error)
func (c *Client) Teams(ctx context.Context, first *int64, after *string) (*TeamConnection, error)
func (c *Client) Projects(ctx context.Context, first *int64, after *string) (*ProjectConnection, error)
```

**Parameters:**
- `first`: Page size (*int64) - max 250, recommended 50, nil uses API default (~50)
- `after`: Pagination cursor (*string) - nil for first page, use `PageInfo.EndCursor` for next page
- Returns `*Connection` with `Nodes` array and `PageInfo` for pagination state

**Nil Semantics:**
```go
// Pattern 1: Not found (nil result, nil error)
issue, err := client.Issue(ctx, "ENG-999")
if err != nil {
    return err  // API failure
}
if issue == nil {
    // Not found - this is NOT an error
}

// Pattern 2: Empty results (empty array, no error)
issues, err := client.Issues(ctx, &first, nil)
if err != nil {
    return err  // API failure
}
if len(issues.Nodes) == 0 {
    // No results found - this is NOT an error
}

// Pattern 3: End of pagination (nil after is valid)
var after *string  // nil for first page
for {
    issues, err := client.Issues(ctx, &first, after)
    if err != nil {
        return err
    }
    if len(issues.Nodes) == 0 {
        break  // No more results
    }
    // Process issues...
    if !issues.PageInfo.HasNextPage {
        break
    }
    after = issues.PageInfo.EndCursor  // *string for next page
}
```

### Mutation Methods

All mutations return `(result, error)` where result contains the created/updated resource.

**Create:**
```go
func (c *Client) IssueCreate(ctx context.Context, input IssueCreateInput) (*IssuePayload, error)
```

**Update:**
```go
func (c *Client) IssueUpdate(ctx context.Context, id string, input IssueUpdateInput) (*IssuePayload, error)
```

**Delete:**
```go
func (c *Client) IssueDelete(ctx context.Context, id string, permanentlyDelete *bool) error
// permanentlyDelete: nil = trash (30-day grace), &true = permanent
```

**Archive/Unarchive:**
```go
func (c *Client) IssueArchive(ctx context.Context, id string, trash *bool) error
func (c *Client) IssueUnarchive(ctx context.Context, id string) error
func (c *Client) InitiativeArchive(ctx context.Context, id string) error
func (c *Client) InitiativeUnarchive(ctx context.Context, id string) error
func (c *Client) ProjectArchive(ctx context.Context, id string) error
func (c *Client) ProjectUnarchive(ctx context.Context, id string) error
func (c *Client) TeamUnarchive(ctx context.Context, id string) error
func (c *Client) DocumentUnarchive(ctx context.Context, id string) error
```

**Parameters:**
- Input structs use pointers for optional fields (nil = don't change)
- Returns payload with `.Issue` field containing the result
- Error if mutation fails, authentication issues, or permissions denied

---

## API Coverage

### Query Methods

| Resource | Get | List | Search | Iterator |
|----------|-----|------|--------|----------|
| Issues | `Issue(ctx, id)` | `Issues(ctx, first, after)` | `SearchIssues(...)` | `NewIssueIterator(...)` |
| Teams | `Team(ctx, id)` | `Teams(ctx, first, after)` | - | `NewTeamIterator(...)` |
| Projects | `Project(ctx, id)` | `Projects(ctx, first, after)` | - | `NewProjectIterator(...)` |
| Users | `User(ctx, id)` | `Users(ctx, first, after)` | - | - |
| Comments | `Comment(ctx, id)` | `Comments(ctx, first, after)` | - | `NewCommentIterator(...)` |
| Labels | `IssueLabel(ctx, id)` | `IssueLabels(ctx, first, after)` | - | - |
| Workflow States | `WorkflowState(ctx, id)` | `WorkflowStates(ctx, first, after)` | - | - |
| Cycles | `Cycle(ctx, id)` | `Cycles(ctx, first, after)` | - | - |
| Other | Roadmaps, Documents, Attachments, Initiatives, Templates | | | |

Special: `Viewer(ctx)` - authenticated user, `Organization(ctx)` - workspace info

### Mutation Methods

| Resource | Create | Update | Delete |
|----------|--------|--------|--------|
| Issues | `IssueCreate(ctx, input)` | `IssueUpdate(ctx, id, input)` | `IssueDelete(ctx, id)` |
| Comments | `CommentCreate(ctx, input)` | `CommentUpdate(ctx, id, input)` | `CommentDelete(ctx, id)` |
| Labels | `IssueLabelCreate(ctx, input)` | `IssueLabelUpdate(ctx, id, input)` | `IssueLabelDelete(ctx, id)` |
| Teams | `TeamCreate(ctx, input)` | `TeamUpdate(ctx, id, input)` | `TeamDelete(ctx, id)` |
| Projects | `ProjectCreate(ctx, input)` | `ProjectUpdate(ctx, id, input)` | `ProjectDelete(ctx, id)` |
| Cycles | `CycleCreate(ctx, input)` | `CycleUpdate(ctx, id, input)` | `CycleArchive(ctx, id)` |
| Initiatives | `InitiativeCreate(ctx, input)` | `InitiativeUpdate(ctx, id, input)` | `InitiativeDelete(ctx, id)` |

**Additional Methods:**
- `SubInitiatives(ctx, id, first, after)` - List sub-initiatives of a parent
- `TeamMemberships(ctx, teamID, first, after)` - Get team memberships for member management

---

## Metrics

```go
// Per-operation metrics (not generic "graphql")
linear_requests_total{operation="IssueCreate", status_code="200"}
linear_request_duration_seconds{operation="IssueCreate", quantile="0.95"}
linear_errors_total{operation="IssueCreate", error_type="RateLimited"}
linear_retries_total{reason="rate_limited"}
linear_rate_limit_remaining{limit_type="requests"}

// Multi-tenant: isolated registries
prodClient, _ := linear.NewClient(prodKey,
    linear.WithMetricsRegistry(prodRegistry, "prod"))
```

### Prometheus Alerts

**Critical:**

```promql
# Error rate > 50%
rate(linear_errors_total[5m]) / rate(linear_requests_total[5m]) > 0.5

# All requests failing
rate(linear_requests_total{status_code=~"2.."}[3m]) == 0
  and rate(linear_requests_total[3m]) > 0
```

**Warning:**

```promql
# Rate limit < 10%
linear_rate_limit_remaining{limit_type="requests"} /
  linear_rate_limit_remaining{limit_type="requests"} offset 5m < 0.1

# Retry rate > 10%
rate(linear_retries_total[10m]) / rate(linear_requests_total[10m]) > 0.1

# p95 latency > 5s
histogram_quantile(0.95,
  rate(linear_request_duration_seconds_bucket[10m])) > 5
```

---

## Structured Logging

```json
{
  "time": "2025-01-15T10:30:00Z",
  "level": "INFO",
  "msg": "request_completed",
  "operation": "IssueCreate",
  "status": 200,
  "duration_ms": 234,
  "request_id": "req-abc-123",
  "rate_limit_remaining": 1450
}

{
  "time": "2025-01-15T10:30:15Z",
  "level": "WARN",
  "msg": "retry_attempt",
  "operation": "SearchIssues",
  "attempt": 2,
  "backoff_ms": 2000,
  "reason": "rate_limited",
  "request_id": "req-def-456"
}
```

---

## Performance

| Operation | p50 | p95 | Notes |
|-----------|-----|-----|-------|
| `Viewer()` | 80ms | 150ms | Cached by Linear |
| `Issues(50)` | 120ms | 250ms | ~500 complexity points |
| `IssueCreate()` | 200ms | 400ms | Includes webhooks |
| `SearchIssues()` | 150ms | 350ms | ElasticSearch backend |

**Rate Limits:**

- API Keys: 250,000 complexity points/hour
- Users: 1,500 requests/hour
- Leaky bucket algorithm

---

## Security

### TLS Configuration

```go
linear.WithTLSConfig(&tls.Config{
    MinVersion: tls.VersionTLS12,
    VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
        // Optional: certificate pinning
    },
})
```

### Credential Rotation (Experimental)

```go
type SecretsProvider struct {
    secretName string
    manager    *secretsmanager.SecretsManager
}

func (p *SecretsProvider) GetCredential(ctx context.Context) (string, error) {
    result, _ := p.manager.GetSecretValue(&secretsmanager.GetSecretValueInput{
        SecretId: aws.String(p.secretName),
    })
    return *result.SecretString, nil
}

// Client refreshes on 401 and retries once
// NOTE: Requires production testing before deployment
client, _ := linear.NewClient("", linear.WithCredentialProvider(&SecretsProvider{...}))
```

### Audit Logging

All requests include `request_id` for correlation with Linear support.

---

## Testing

```bash
# Unit tests (no API key)
make test

# Integration - read-only
LINEAR_API_KEY=lin_api_xxx make test-read

# Integration - mutations
LINEAR_API_KEY=lin_api_xxx make test-write

# Coverage
LINEAR_API_KEY=lin_api_xxx make test-coverage
```

**Build Tags:**

- No tag: Mock tests (httptest.Server)
- `//go:build read`: Read-only integration
- `//go:build write`: Mutation integration

Coverage: 60%+ (target: 80%)

---

## Examples

| Task | File | Notes |
|------|------|-------|
| Create issue | [create_issue](../examples/tasks/create_issue/main.go) | Basic |
| Search issues | [search_issues](../examples/tasks/search_issues/main.go) | Operators |
| Pagination | [list_issues_iterator](../examples/tasks/list_issues_iterator/main.go) | Auto-pagination |
| Rate limits | [handle_rate_limits](../examples/tasks/handle_rate_limits/main.go) | Backoff |
| Credential rotation | [credential_rotation](../examples/tasks/credential_rotation/main.go) | Experimental |
| Circuit breaker | [handle_circuit_breaker](../examples/tasks/handle_circuit_breaker/main.go) | Fail-fast |
| Concurrent requests | [concurrent_requests](../examples/tasks/concurrent_requests/main.go) | Goroutines |

---

## FAQ

**Q: Official Linear SDK?**
A: No. This is a third-party Go client. Official Linear SDKs at https://github.com/linear

**Q: API key permissions?**
A: Read for queries, Write for mutations. Configure at https://linear.app/settings/account/security

**Q: Webhooks supported?**
A: No. Use Linear's webhook API directly.

**Q: Thread-safe iterators?**
A: Yes. Mutex-protected, safe for concurrent use.

**Q: Credential rotation production-ready?**
A: No. `WithCredentialProvider` is experimental. Test thoroughly before deployment.

**Q: Circuit breaker behavior?**
A: Opens after 5 failures, resets after 60s. Returns errors immediately while open.

**Q: Multi-workspace support?**
A: Yes. Use `WithMetricsRegistry()` for isolated metrics per workspace.
