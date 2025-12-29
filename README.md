# go-linear

Type-safe Go client for the Linear GraphQL API with CLI for humans and MCP server for AI agents.

[![Go Reference](https://pkg.go.dev/badge/github.com/chainguard-sandbox/go-linear.svg)](https://pkg.go.dev/github.com/chainguard-sandbox/go-linear)
[![Go Report Card](https://goreportcard.com/badge/github.com/chainguard-sandbox/go-linear)](https://goreportcard.com/report/github.com/chainguard-sandbox/go-linear)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

```bash
go get github.com/chainguard-sandbox/go-linear
```

---

## What's Inside

**Go SDK (`pkg/linear`)** - Use Linear from your Go applications
- Type-safe GraphQL client (~7400 LOC)
- Complete API: Issues, Teams, Projects, Comments, Labels, Workflows
- Production features: retry, circuit breaker, rate limiting, metrics

**CLI (`go-linear`)** - Command-line tool for humans
- Interactive commands: `go-linear issue list --team=ENG --priority=1`
- Smart defaults, filtering, batch operations
- Works standalone or in scripts

**MCP Server (`go-linear mcp start`)** - API for AI agents
- Same binary, different mode - powered by [ophis](https://github.com/njayp/ophis)
- Ophis automatically converts all CLI commands into 72 MCP tools
- AI agents call commands via JSON-RPC instead of using the CLI directly
- Optimized output for token efficiency

### Design

- **Type Safety**: Generated from GraphQL schema. Schema changes caught at compile time.
- **Reliability**: Circuit breakers open after 5 consecutive failures. Retry bounded to 90s total.
- **Observability**: All operations logged with request_id. Metrics include operation name (not generic "graphql").
- **Error Handling**: Typed errors with `errors.As()` support. Error chains preserved with `Unwrap()`.

---

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"

    "github.com/chainguard-sandbox/go-linear/pkg/linear"
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

### Production Configuration

```go
import (
    "crypto/tls"
    "log/slog"
    "os"
    "time"

    "github.com/chainguard-sandbox/go-linear/pkg/linear"
)

logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

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
        logger.Info("rate_limit",
            "requests_remaining", info.RequestsRemaining,
            "requests_limit", info.RequestsLimit,
        )
    }),
)
```

**See also:**
- [Production example](examples/production/main.go) - Complete setup
- [Prometheus integration](examples/prometheus/main.go) - Metrics collection
- [Error handling](examples/tasks/handle_auth_errors/main.go) - Typed error patterns

---

## CLI for Humans

Quick start:

```bash
# Build
make build-cli

# Set API key
export LINEAR_API_KEY=lin_api_xxx

# Use it
go-linear issue list --assignee=me --priority=1
go-linear issue create --team=Engineering --title="Fix bug"
go-linear team list
```

Features:
- Smart field defaults (returns 8 fields instead of 50+)
- 44 filters for precise queries
- Batch operations (update up to 50 issues)
- Count mode (`--count` returns just totals)
- Relative dates (`yesterday`, `7d`, `2w`)
- Name resolution (`--team=Engineering` not UUIDs)

See [CLI Quick Start](docs/CLI-QUICK-START.md) for more.

---

## MCP Server for AI Agents

The **same binary** works as an MCP server for AI agents like Claude.

**How it works:** [Ophis](https://github.com/njayp/ophis) automatically converts your CLI into an MCP server by:
1. Walking the Cobra command tree
2. Generating JSON schemas from flags
3. Creating 72 MCP tool definitions
4. Executing tools by spawning CLI subprocesses

**Architecture:**
```
AI Agent (Claude) ─stdio─► go-linear mcp start ─spawns─► go-linear issue list ─► Linear API
                           (JSON-RPC 2.0)               (regular CLI)
```

**Why this approach:**
- Agents work with high-level commands, not raw GraphQL
- CLI optimizations (defaults, filtering) reduce token usage
- Same battle-tested logic for humans and AI
- No separate codepaths to maintain

### Security Model

The MCP server exposes 72 tools from the CLI. Tool categories:

| Category | Examples | User Confirmation |
|----------|----------|-------------------|
| Read-Only | `list`, `get` operations | No |
| Mutable | `create`, `update` operations | Yes |
| Destructive | `delete` operations | Yes |

Mutable operations require user confirmation. Destructive operations cannot be undone.

### Quick Setup

**1. Build:**
```bash
make build-cli
```

**2. Add to Claude Code:**
```bash
claude mcp add --transport stdio go-linear \
  --env LINEAR_API_KEY=lin_api_xxx \
  -- /absolute/path/to/bin/go-linear mcp start
```

**3. Verify:**
```bash
claude mcp list
```

**Need help?** See [Claude Code Setup Guide](docs/CLAUDE-SETUP.md) for:
- Step-by-step instructions
- Getting your Linear API key
- Configuration scopes
- Troubleshooting

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `LINEAR_API_KEY` | Required | Linear API key |
| `LINEAR_BASE_URL` | api.linear.app/graphql | Custom API endpoint |
| `LINEAR_TIMEOUT` | 30s | Request timeout |
| `LINEAR_RETRY_ATTEMPTS` | 3 | Retry count (0 to disable) |
| `LINEAR_RETRY_INITIAL` | 1s | Initial retry backoff |
| `LINEAR_RETRY_MAX` | 30s | Maximum retry backoff |
| `LINEAR_CIRCUIT_BREAKER_FAILURES` | 5 | Failures before circuit opens |
| `LINEAR_CIRCUIT_BREAKER_TIMEOUT` | 60s | Circuit reset timeout |
| `LINEAR_TLS_MIN_VERSION` | 1.2 | Minimum TLS version (1.2 or 1.3) |
| `LINEAR_LOG_LEVEL` | disabled | Logging level (debug, info, warn, error) |
| `LINEAR_METRICS_ENABLED` | false | Enable Prometheus metrics (true/false) |

**Note:** All environment variables are read once at CLI startup for optimal performance.

### Available Commands

The MCP server exposes 72 tools auto-generated from CLI commands. Major command groups:

| Command Group | Examples | Description |
|---------------|----------|-------------|
| `issue` | `list`, `get`, `create`, `update`, `delete` | Issue management |
| `comment` | `list`, `get`, `create`, `update`, `delete` | Comment operations |
| `team` | `list`, `get` | Team discovery |
| `project` | `list`, `get`, `create`, `update` | Project management |
| `label` | `list`, `get`, `create`, `update`, `delete` | Label operations |
| `user` | `list`, `get`, `completed` | User information |
| `state` | `list`, `get` | Workflow states |
| `cycle` | `list`, `get`, `create`, `update` | Sprint cycles |
| `viewer` | `get` | Authenticated user |
| `attachment` | `list`, `get`, `create`, `delete` | File attachments |
| `document` | `list`, `get` | Document management |
| `roadmap` | `list`, `get` | Roadmap planning |
| `initiative` | `list`, `get` | Strategic initiatives |
| `template` | `list`, `get` | Issue templates |
| `favorite` | `list`, `create`, `delete` | Favorites |
| `reaction` | `create`, `delete` | Emoji reactions |
| `notification` | `list`, `subscribe`, `unsubscribe` | Notifications |

### Testing

```bash
# View all CLI commands (becomes MCP tools)
./bin/go-linear --help

# Test MCP server with inspector
npx @modelcontextprotocol/inspector ./bin/go-linear -- mcp start
```

Opens web UI at http://localhost:5173 to test tools interactively.

### Logging & Metrics

Add environment variables when configuring:

```bash
claude mcp add --transport stdio go-linear \
  --env LINEAR_API_KEY=lin_api_xxx \
  --env LINEAR_LOG_LEVEL=info \
  --env LINEAR_METRICS_ENABLED=true \
  -- /path/to/go-linear mcp start
```

See [Environment Variables](#environment-variables) for all options.

### Why AI Agents Love This CLI

These CLI features reduce token usage for AI agents:

**1. Field defaults** - Returns 10 fields instead of 50+
```bash
go-linear issue get ENG-123 --output=json
# Returns: id, title, state, team, priority, assignee, url, createdAt

# Custom fields
go-linear issue get ENG-123 --fields=id,title,priority --output=json

# Add to defaults
go-linear issue get ENG-123 --fields=defaults,estimate,labels --output=json
```

**2. Count mode** - Get totals without full data
```bash
# Instead of returning 50 full issues...
go-linear issue list --team=Engineering --priority=1 --count
# Returns: {"count": 5}
```

**3. Smart filtering** - 44 filters for precise queries
```bash
# Complex query in one call
go-linear issue list \
  --state=Triage \
  --has-suggested-teams \
  --created-after=7d \
  --comment-by=me,manager@company.com \
  --output=json

# Batch operations
go-linear issue batch-update \
  --state=Triage \
  --has-suggested-teams \
  --set-state=Backlog \
  --dry-run
```

See [FILTERS.md](docs/FILTERS.md) for all 44 filters.

**Rate limit monitoring**:
```bash
# Check API quota
linear status

# Output:
# LINEAR API RATE LIMIT STATUS
#
# REQUEST-BASED LIMITS
#   Remaining: 2,485 / 2,500 (99.4%)
#   Resets in: 54m32s (at 2025-12-11T19:47:30Z)
#
# COMPLEXITY-BASED LIMITS
#   Remaining: 248,750 / 250,000 (99.5%)
#   Resets in: 54m32s (at 2025-12-11T19:47:30Z)

# JSON format for scripting
linear status --output=json
```

---

## Architecture

### Project Structure

```
go-linear/
├── pkg/linear/              # Public API
│   ├── client.go           # 45 methods
│   ├── transport.go        # Retry, circuit breaker, metrics
│   ├── errors.go           # Typed error hierarchy
│   ├── pagination.go       # Thread-safe iterators
│   ├── options.go          # Functional options
│   └── metrics.go          # Prometheus collectors
├── internal/graphql/       # Generated code (gqlgenc)
│   ├── client.go
│   └── models.go          # ~30k lines
├── cmd/linear/             # MCP server & CLI
│   └── main.go
├── queries/                # GraphQL definitions
└── examples/               # 25+ examples
```

### Code Generation

```
schema.graphql → gqlgenc → internal/graphql/* → pkg/linear
```

Changes to `schema.graphql` require `make generate`. Schema synced from Linear's TypeScript SDK.

### Error Handling

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

### Metrics

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

### Performance

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

### Threat Mitigation

| Threat | Mitigation |
|--------|------------|
| Credential theft | TLS 1.2+, no plaintext storage |
| API abuse | Circuit breakers, rate monitoring, bounded retries |
| Injection | Typed GraphQL parameters |
| MITM | Certificate validation, optional pinning |
| Supply chain | Dependency scanning (govulncheck, trivy) |

### Features

**TLS Configuration:**
```go
linear.WithTLSConfig(&tls.Config{
    MinVersion: tls.VersionTLS12,
    VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
        // Optional: certificate pinning
    },
})
```

**Credential Rotation (⚠️ Experimental - not production-tested):**
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

**Audit Logging:**
```json
{
  "level": "INFO",
  "msg": "request_completed",
  "operation": "IssueCreate",
  "status": 200,
  "duration_ms": 234,
  "request_id": "req-abc-123"
}
```

All requests include `request_id` for correlation with Linear support.

### Compliance

**GDPR:** Issue data may contain PII. `IssueDelete` supports data deletion. Linear hosted in US.

**SOC 2:** Structured audit logs, circuit breakers for availability.

**Vulnerability Disclosure:** mark.esler@chainguard.dev (72-hour acknowledgment SLA). See [SECURITY.md](SECURITY.md).

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

### Pagination

**Manual:**
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

**Automatic (Thread-Safe):**
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

## Monitoring

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

### Structured Logging

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

**Documentation:**
- [Incident Response](docs/RUNBOOK.md)
- [Monitoring Setup](docs/MONITORING.md)

---

## Examples

| Task | File | Notes |
|------|------|-------|
| Create issue | [create_issue](examples/tasks/create_issue/main.go) | Basic |
| Search issues | [search_issues](examples/tasks/search_issues/main.go) | Operators |
| Pagination | [list_issues_iterator](examples/tasks/list_issues_iterator/main.go) | Auto-pagination |
| Rate limits | [handle_rate_limits](examples/tasks/handle_rate_limits/main.go) | Backoff |
| Credential rotation | [credential_rotation](examples/tasks/credential_rotation/main.go) | ⚠️ Experimental |
| Circuit breaker | [handle_circuit_breaker](examples/tasks/handle_circuit_breaker/main.go) | Fail-fast |
| Concurrent requests | [concurrent_requests](examples/tasks/concurrent_requests/main.go) | Goroutines |

---

## Configuration

### Client Options

| Option | Default | Description |
|--------|---------|-------------|
| `WithTimeout(d)` | 30s | Request timeout |
| `WithRetry(max, initial, maxBackoff)` | 3, 1s, 30s | Exponential backoff |
| `WithMaxRetryDuration(d)` | 90s | Total retry limit |
| `WithCircuitBreaker(config)` | nil | Fail-fast config |
| `WithLogger(logger)` | nil | slog logger |
| `WithMetrics()` | disabled | Prometheus metrics |
| `WithMetricsRegistry(reg, ns)` | global, "linear" | Multi-tenant metrics |
| `WithRateLimitCallback(f)` | nil | Rate limit monitoring |
| `WithTLSConfig(config)` | TLS 1.2+ | TLS settings |
| `WithCredentialProvider(p)` | static | ⚠️ Experimental rotation |
| `WithBaseURL(url)` | api.linear.app | Custom endpoint |

### Transport Behavior

**Retry:**
- Retries: 429 (rate limit), 5xx (server errors), timeouts
- No retry: 4xx (except 429), 2xx responses
- Respects `Retry-After` header
- Bounded total time (90s default)

**Circuit Breaker:**
- Opens after N failures (default: 5)
- Half-open after timeout (default: 60s)
- Closes on successful request

**Rate Limiting:**
- Parses `X-RateLimit-*` headers
- Tracks request and complexity limits
- Automatic backoff on 429

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

## Development

### Setup

```bash
git clone https://github.com/chainguard-sandbox/go-linear
cd go-linear
make dev
```

### Commands

```bash
make check         # fmt, vet, lint, test
make fmt           # Format code
make lint          # golangci-lint
make generate      # Regenerate GraphQL client
make test-coverage # Coverage report
```

### Code Generation

```bash
make generate
# 1. Fetches Linear schema (or uses local)
# 2. Runs gqlgenc with queries/*.graphql
# 3. Generates internal/graphql/*
```

Config: `gqlgenc.yaml`

### Contributing

1. Open issue
2. Fork repository
3. Create feature branch
4. Add tests
5. Run `make check`
6. Submit PR

**Review criteria:**
- Type safety maintained
- Tests included
- golangci-lint passes
- Coverage maintained

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

---

## Resources

**Documentation:**
- [Go Package Docs](https://pkg.go.dev/github.com/chainguard-sandbox/go-linear)
- [Linear GraphQL API](https://developers.linear.app)
- [Linear Schema](https://studio.apollographql.com/public/Linear-API/variant/current/home)
- [Model Context Protocol](https://modelcontextprotocol.io/)

**Repositories:**
- [go-linear](https://github.com/chainguard-sandbox/go-linear)
- [Linear TypeScript SDK](https://github.com/linear/linear)
- [MCP Specification](https://github.com/anthropics/mcp)

**Support:**
- [GitHub Issues](https://github.com/chainguard-sandbox/go-linear/issues)
- [Linear Status](https://status.linear.app)
- Security: mark.esler@chainguard.dev

---

## Status

### v1.0.0 (Stable)

**Features:**
- ✅ Type-safe GraphQL operations (7400 LOC)
- ✅ Complete API coverage (45 methods)
- ✅ Automatic retry, exponential backoff
- ✅ Rate limit detection and handling
- ✅ Circuit breakers
- ✅ Bounded retry time
- ✅ Structured logging (slog)
- ✅ Per-operation Prometheus metrics
- ✅ Multi-tenancy support
- ✅ Error chain preservation
- ✅ Thread-safe pagination iterators
- ✅ TLS 1.2+ enforcement
- ✅ MCP server (72 tools)
- ✅ Field defaults (minimizes output verbosity)
- ✅ Advanced filtering (44 filters)
- ✅ Batch operations (up to 50 issues)
- ✅ Count aggregation (efficient "how many?" queries)
- ✅ 60%+ test coverage

**Experimental:**
- ⚠️ Credential rotation (not production-tested)

**Stability:**
- Semantic versioning
- No breaking changes in v1.x
- Schema changes require minor bump

### Roadmap

**v1.1.0:**
- [ ] OpenTelemetry tracing
- [ ] Query complexity estimation
- [ ] Webhook validation helpers
- [ ] Credential rotation testing

**v1.2.0:**
- [ ] Batch mutations
- [ ] GraphQL subscriptions
- [ ] Performance benchmarks

---

## License

Apache 2.0 - See [LICENSE](LICENSE)

**Dependencies:** See [go.mod](go.mod). All Apache 2.0 or MIT licensed.

---

**Built with:**
- [gqlgenc](https://github.com/Yamashou/gqlgenc) - GraphQL generation
- [Prometheus client_golang](https://github.com/prometheus/client_golang)
- [slog](https://pkg.go.dev/log/slog)

Chainguard, Inc. | [chainguard.dev](https://chainguard.dev)
