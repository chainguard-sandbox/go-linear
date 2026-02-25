# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - Unreleased

### Breaking Changes
- **JSON-only output**: Dropped table output — all commands output JSON exclusively. The `--output` flag is removed (~4,000 lines removed). See [docs/MIGRATION.md](docs/MIGRATION.md).
- **SDK**: `IssueDelete(ctx, id)` → `IssueDelete(ctx, id, permanentlyDelete *bool)`. Pass `nil` to preserve v1.x behavior.

### Added

**Lifecycle Management** (12 commands):
- `issue archive` / `issue unarchive` / `issue delete --permanent`
- `initiative archive` / `initiative unarchive` / `initiative list-sub`
- `project archive` / `project unarchive` / `project milestone-list` / `project status-list`
- `team unarchive` / `team remove-member`
- `document unarchive`

**Resolver Improvements**:
- Resolvers now show available options on "not found" errors (e.g., lists valid project statuses)
- Issue ID resolution added to `delete`, `archive`, and `unarchive` commands

**Project Updates**:
- Comprehensive `project update` flags for all mutable fields

**Templates**:
- `template get` now includes `templateData` field showing pre-filled template values
- `issue create --use-default-template` to apply the team's default template

**CLI**:
- `--version` flag with build metadata (version, commit, tree state, build date)
- `--assignee=none` to unassign issues
- Priority validation (0–4) across all commands

**Error Handling**:
- Structured error handling infrastructure with improved messages for fetch failures and empty input

### Changed
- Sync upstream schema to `@linear/sdk@75.0.0`
- Go 1.25.7 (fixes GO-2026-4337 TLS vulnerability)
- Removed goreleaser; simplified Makefile sync check

### Fixed
- Config-default labels no longer override template labels when `--label` is not explicitly passed
- Resolve gosec findings: G704 (subprocess), G703 (error string), G117 (HTTP redirect), G706 (env var log sanitization)
- Skip ambiguous state name tests on multi-team workspaces

### Documentation
- Added `CONTRIBUTING.md`
- Added `docs/MIGRATION.md` for v1→v2 upgrade guide
- Removed static counts and outdated docs

### Security
- Bumped `step-security/harden-runner`, `actions/checkout`, `actions/setup-go`, `aquasecurity/trivy-action`
- Bumped `github.com/olekukonko/tablewriter` and minor-and-patch dependency group

## [1.4.1] - 2026-01-16

### Fixed
- **Issue Resolution**: Fixed `ResolveIssue` using fuzzy search instead of direct lookup, which could return wrong issues (e.g., "PSEC-78" returning PSEC-128). Now uses Linear's `issue(id:)` query for exact identifier matching. Fixes #22.
- **Issue Update**: Fixed `issue update --project` silently failing when using project names instead of UUIDs. The command now properly resolves issue identifiers (e.g., ENG-123) to UUIDs before API calls. Fixes #20.
- **Name Resolution**: Added consistent cycle and project name resolution across all issue commands (`update`, `create`, `batch-update`). Previously these flags only accepted UUIDs, inconsistent with other fields like `--assignee` and `--state` that accept human-readable names.

## [1.4.0] - 2026-01-12

### Added

**Notification Inbox** (3 commands):
- `notification list`: View inbox notifications with filtering
- `notification get`: Get notification details
- `notification unarchive`: Restore archived notifications to inbox

**AI Suggestions** (1 command):
- `issue suggestions`: View AI-recommended assignees, teams, labels, projects, and related issues

**Comment Threading**:
- `comment create --parent`: Create threaded replies
- `comment get`: Shows parent context and child replies with pagination (max 5000)

**Issue Templates**:
- `issue create --template`: Apply templates during creation (title becomes optional)

**Team Management** (1 command):
- `team add-member`: Add users to teams (admin required)

**Total**: 94 commands (was 89), 94 MCP tools

### Changed
- Roadmap commands marked as deprecated (Linear deprecated in favor of initiatives)
- Migrate `multicache` to `fido` (dependency renamed upstream)

### Fixed
- Fix broken test related to sub-second TTL in cache
- Fix concurrent cache race condition

### Defensive Coding
- Request body size limiting (default 10MB, configurable via LINEAR_MAX_BODY_SIZE)
- Prevents OOM from bugs in query generation (hardening, not fixing exploitable vulnerability)

### Developer Tools
- Added `make modernize` target for modern Go pattern analysis (min/max, range-int, etc.)

## [1.3.0] - 2026-01-09

### Added

**Status Updates** (8 commands):
- `project status-update-{create,list,get,delete}`: Track project progress with health indicators
- `initiative status-update-{create,list,get,archive}`: Track initiative progress
- Health states: onTrack, atRisk, offTrack

**Team Metrics** (1 command):
- `team velocity`: Calculate performance from completed cycles (avg points/issues per cycle)

**Command Enhancements**:
- `initiative get`: Accepts name or UUID; displays ID, status, health, target date, owner, parent initiative, linked projects (with progress), description, URL
- `project get`: Accepts name or UUID; displays ID, progress, health, all dates (target/started/completed/canceled), lead, team(s), linked initiatives (with status), description, URL
- `cycle get`: Displays scope history and issue count metrics
- `initiative delete`: Delete initiatives with confirmation

Name resolution in get commands simplifies workflows.
Enhanced displays show strategic relationships and full context for weekly updates.
All metrics from existing Linear API fields.

**Initiative-Project Linking** (2 commands):
- `initiative add-project`: Link projects to strategic initiatives
- `initiative remove-project`: Unlink projects from initiatives

**Document Management** (3 commands):
- `document create`: Create knowledge base documents (requires --team, --project, --initiative, or --issue)
- `document update`: Modify document title and content
- `document delete`: Remove documents with confirmation

**Total**: 89 commands (was 74), 89 MCP tools

### Changed
- GraphQL queries updated with progress, health, and history fields for metrics
- `initiative create/update`: Fixed field names (OwnerId→OwnerID), removed unsupported parent filtering, fixed status type casting
- `initiative get`: Fixed description pointer handling
- `initiative remove-project`: Uses pagination to handle workspaces with 100+ links (prevents silent failures)

### Fixed
- Initiative delete command compilation errors
- Nil pointer checks in initiative and project commands
- Lint issues (nil checks, prealloc)

## [1.2.1] - 2026-01-07

### Fixed
- **Attachments**: Fixed `attachment list` command failure caused by JSONObject scalar mapping. Linear's `source` field returns native JSON objects (TypeScript behavior) but gqlgenc defaulted to string mapping. Now properly maps `JSONObject` to `map[string]interface{}` in gqlgenc.yaml, preserving all source metadata (imageUrl, syncedCommentId, channelId, etc.). Fixes #14.
- **CLI**: Fixed metadata flag parsing in `attachment create` to properly convert JSON string to map.

### Added
- **Tests**: Comprehensive attachment integration tests with full lifecycle coverage (create, list, filter, get, delete) and all link types (URL, GitHub PR, Slack).

### Dependencies
- Bump golang.org/x/sync from 0.17.0 to 0.19.0

## [1.2.0] - 2026-01-06

### Overview

**New CLI layer**: Introduced `go-linear` CLI between the SDK and MCP server. The CLI is context-engineered—it absorbs GraphQL complexity so agents (and humans) work at the semantic level.

### Architecture

```
v1.1.0: Agent → MCP → SDK → GraphQL
v1.2.0: Agent → MCP → CLI → SDK → GraphQL
```

The CLI operates at the right altitude—high enough to hide GraphQL mechanics, specific enough to express precise intent. This protects the agent's attention budget by absorbing:
- **UUID resolution**: `--team=ENG` not `--team-id=abc-123-...`
- **Field defaults**: 8 fields returned, not 50+
- **Pagination**: Automatic, or `--count` for totals
- **Date parsing**: `--created-after=7d`
- **Batch operations**: Update 50 issues in one call

MCP server auto-generated from CLI via [ophis](https://github.com/njayp/ophis).

### Added

#### CLI (`go-linear`) - 74 commands
New semantic interface for humans and agents:
- **Issue**: `list`, `get`, `create`, `update`, `delete`, `search`, `batch-update`, `add-label`, `remove-label`, `relate`, `unrelate`
- **Attachment**: `list`, `get`, `create`, `delete`, `link-github`, `link-slack`, `link-url`
- **Comment**: `list`, `get`, `create`, `update`, `delete`
- **Cycle**: `list`, `get`, `create`, `update`, `archive`
- **Project**: `list`, `get`, `create`, `update`, `delete`, `milestone-create`, `milestone-update`, `milestone-delete`
- **Team**: `list`, `get`, `create`, `update`, `delete`, `members`
- **User**: `list`, `get`, `completed`
- **Label**: `list`, `get`, `create`, `update`, `delete`
- **State**: `list`, `get`
- **Notification**: `subscribe`, `unsubscribe`, `archive`, `update`
- **Reaction**: `create`, `delete`
- **Favorite**: `create`, `delete`
- **Initiative**: `list`, `get`, `create`, `update`
- **Document/Roadmap/Template**: `list`, `get`
- **Utility**: `viewer`, `organization`, `status`

#### Filtering (44 flags)
- Issue: SLA status, customer counts, AI suggestions, relationships, collections
- Cycle: 15 filters (active, past, future, date ranges)
- Project/Comment/Label/State/Team/User: 5-16 filters each

#### Output optimization
- `--fields=defaults` (8 fields vs 50+)
- `--count` returns `{"count": N}` (99% token reduction)
- `Team.issueCount` without pagination

#### Resolvers
- `ResolveCycle`, `ResolveDocument`, `ResolveInitiative`, `ResolveProject`
- Case-insensitive, fuzzy matching

#### Infrastructure
- GitHub Actions release workflow (5 platforms)
- Performance profiling
- Multicache resolver for MCP persistence
- clog structured logging

#### Documentation
- Claude skill (`.claude/skills/go-linear/SKILL.md`)
- SDK docs (`docs/SDK.md`)
- Filter docs (`docs/FILTERS.md`)
- Claude setup (`docs/CLAUDE-SETUP.md`)

### Changed
- **MCP rebuilt**: Now wraps CLI instead of SDK directly
- **README**: Rewritten with context engineering principles
- **Client**: Split into entity-focused files

### Fixed
- Resolver cache: synchronous writes
- Context cancellation
- Rate limit parsing
- Identifier resolution
- Nil safety across filter system

### Testing
- Command coverage: 61-87%
- Filter tests: all 48 issue filters
- Internal packages: 60%+

### Security
- gosec and nilaway in CI
- Dependency updates

### Upstream
- @linear/sdk@68.1.0

## [1.1.0] - 2025-12-10

### Added
- **Model Context Protocol (MCP) server** for AI agent integration
  - 13 tools exposing Linear API operations (viewer, teams, issues, search, comments, labels, workflow states, users)
  - Read-only operations: 9 safe tools for querying Linear data
  - Mutable operations: 4 tools (create_issue, update_issue, create_comment, delete_issue) with confirmation warnings
  - JSON-RPC 2.0 over stdio following MCP specification
  - Complete tool specifications in `mcp/tools.json` with safety markers
  - Automated test script in `mcp/test-mcp.sh`
  - Example Go client in `examples/mcp-client/`
  - MCP server binary in `cmd/linear-mcp/` with separate module
- **JSON Schema** for AI-friendly type discovery
  - Complete schema definitions in `pkg/linear/schema.json`
  - Embedded schema accessor: `linear.GetSchema()`
  - Input/output types with validation rules and field constraints
- **AI Integration Documentation**
  - Complete MCP guide in `docs/MCP.md` with setup, workflows, and safety details
  - Claude Desktop configuration instructions
  - Linear API permission documentation (Read, Write, Admin, Create issues, Create comments)
  - Permission detection via API responses (401 Unauthorized, 403 Forbidden)

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

### Security
- MCP server includes safety warnings for all mutable operations
- Destructive operations (delete) marked with `x-dangerous` and `x-requires-confirmation` flags
- MCP server binary added to `.gitignore`
- All mutable tools require user confirmation with ⚠️ warnings in descriptions

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
