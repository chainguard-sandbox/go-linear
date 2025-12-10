# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Improved variable naming in examples and documentation for clarity and consistency following Go best practices:
  - Variables with `new*` prefix (e.g., `newTitle`, `newName`) renamed to more descriptive names using `updated*` prefix (e.g., `updatedTitle`, `updatedName`) or context-specific names (e.g., `updatedTargetDate`)
  - `unassign` renamed to `emptyAssignee` to clarify its purpose
  - `added` renamed to `labelIDsToAdd` to be more explicit

  **Affected files:**
  - Documentation examples in `pkg/linear/client.go`
  - Example programs in `examples/tasks/`
  - Test files in `pkg/linear/`

  **Note**: This is a documentation-only change. If you copied code from examples, update variable names to match the new patterns. The API itself is unchanged.

## [1.0.0] - 2025-12-10

### Overview

First stable release of go-linear, a production-ready Go client for the Linear API. This release marks API stability - future changes will follow semantic versioning.

### Features

#### Core API Coverage
- Comprehensive GraphQL client with 45+ methods covering Issues, Teams, Projects, Comments, Labels, Attachments, and more
- Type-safe GraphQL operations via genqlient
- Full CRUD operations for Issues, Comments, Labels, Teams, Projects, and Cycles
- Advanced operations: Issue relationships, attachments (URL, GitHub PR, Slack), reactions, favorites
- Search functionality with `SearchIssues` for full-text search with filters

#### Pagination & Iteration
- Cursor-based pagination for all list operations
- Automatic pagination iterators (IssueIterator, TeamIterator, ProjectIterator, CommentIterator)
- Thread-safe iterators with mutex protection for concurrent access
- Redesigned iterator API returning values instead of pointers

#### Production Features
- **Automatic retry** with exponential backoff and jitter
- **Rate limit handling** with Retry-After header support and monitoring callbacks
- **Circuit breaker** pattern for fail-fast during outages
- **Bounded retry time** prevents request hangs (default: 90s max)
- **Request timeout** support with context cancellation
- **TLS configuration** for security requirements (enforce TLS 1.2+)
- **Dynamic credential management** with auto-refresh on 401 errors
- **HTTP connection pooling** tuned for Linear rate limits

#### Observability & Monitoring
- **Structured logging** with log/slog integration
- **Request ID correlation** for incident tracking with Linear support
- **Per-operation Prometheus metrics** (not just generic "graphql" label)
  - Request counts, duration histograms, error rates by operation
  - Rate limit tracking (requests + complexity)
  - Retry metrics by reason (rate_limited, server_error, network_error)
- **Multi-tenancy support** with instance-scoped metrics registries
- **OpenTelemetry tracing** support for distributed tracing

#### Developer Experience
- Comprehensive API documentation with godoc comments
- 18 task-based examples for common operations
- Production deployment example with best practices
- Prometheus metrics integration example
- Error handling examples with retry patterns
- Operational runbook and monitoring guide

#### Error Handling
- Structured error types: `RateLimitError`, `AuthenticationError`, `ForbiddenError`, `LinearError`
- **Error chain preservation** for proper `errors.As()` and `errors.Is()` support
- Improved GraphQL error extraction with operation context
- Helpful error messages with troubleshooting guidance

#### Input/Output Types
- All mutation input types exported (IssueCreateInput, TeamCreateInput, ProjectCreateInput, etc.)
- Clean public API with internal implementation details hidden
- Pointer fields for optional parameters (nil = omit or unchanged)

### Testing & Quality
- 60%+ test coverage with mock and live integration tests
- Build tags for test isolation (read-only vs mutation tests)
- Race detection in all test runs
- Comprehensive transport layer tests
- Example tests for documentation validation

### Security
- Hardened GitHub Actions workflows (zizmor audit compliance)
- Credential isolation with `persist-credentials: false`
- Minimal token permissions with job-level grants
- Template injection prevention via environment variables
- Gitleaks secret scanning (pre-commit + CI)
- Dependabot for automated security updates
- Trivy and govulncheck for vulnerability scanning

### Documentation
- Comprehensive README with quick start, common tasks, and troubleshooting
- Production deployment guide with configuration options
- Operational runbook for incident response
- Monitoring guide with Prometheus queries and alerts
- Apache 2.0 license and security policy
- Agent-friendly documentation structure

### Infrastructure
- golangci-lint v2 with comprehensive linter configuration
- Pre-commit hooks for formatting, vetting, and linting
- CI workflows for testing, verification, and security scanning
- Upstream schema sync automation
- Automated dependency updates via Dependabot

### Fixed
- golangci-lint v2 migration and configuration compatibility
- Import ordering and code style issues
- Parameter type combinations for cleaner signatures
- HTTP test request bodies using `http.NoBody`
- Builtin shadowing in retry backoff calculation
- Deprecated `issueSearch` replaced with `searchIssues`
- Error chain wrapping for proper error type detection
- go.mod dependency classification (direct vs indirect)

### Changed
- Client struct simplified with config separation
- Module paths standardized after fork
- Iterator API redesigned for better ergonomics
- Removed duplicate error types from internal package

### Removed
- Example binaries from version control
- Unused transitive dependencies

## Notes

**API Stability**: Starting with v1.0.0, this library follows semantic versioning:
- MAJOR version for incompatible API changes
- MINOR version for backwards-compatible functionality additions
- PATCH version for backwards-compatible bug fixes

**Upstream Sync**: GraphQL schema automatically synced from [Linear TypeScript SDK](https://github.com/linear/linear)

**License**: Apache 2.0

[1.1.0]: https://github.com/eslerm/go-linear/releases/tag/v1.1.0
[1.0.0]: https://github.com/eslerm/go-linear/releases/tag/v1.0.0
