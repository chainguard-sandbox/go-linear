# go-linear

Go SDK, CLI, and MCP server for Linear. Generated from upstream schema.

Built for humans and AI agents who want to work with Linear without wrestling GraphQL.

[![Go Reference](https://pkg.go.dev/badge/github.com/chainguard-sandbox/go-linear.svg)](https://pkg.go.dev/github.com/chainguard-sandbox/go-linear)
[![Go Report Card](https://goreportcard.com/badge/github.com/chainguard-sandbox/go-linear)](https://goreportcard.com/report/github.com/chainguard-sandbox/go-linear)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

---

## Architecture

```
Linear GraphQL API
       ↓
   Go SDK (pkg/linear)     ← Type-safe client, generated from schema
       ↓
   CLI (go-linear)         ← Semantic commands, smart defaults
       ↓
   MCP Server              ← Same binary, JSON-RPC mode (via ophis)
```

Each layer adds value:

| Layer | What it provides |
|-------|------------------|
| **SDK** | Type safety, retries, circuit breakers, metrics (for Go developers) |
| **CLI** | Name resolution, field defaults, filtering, batching (for humans and agents) |
| **MCP** | Exposes CLI as 100+ tools for AI agents |

---

## Installation

### From Source

```bash
go install github.com/chainguard-sandbox/go-linear/cmd/linear@latest
```

Note: `go install` names the binary `linear`. Examples below use `go-linear` (the name from `make build`).

Or:

```bash
git clone https://github.com/chainguard-sandbox/go-linear
cd go-linear && make build
```

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/chainguard-sandbox/go-linear/releases).

### Go SDK

```bash
go get github.com/chainguard-sandbox/go-linear
```

---

## Quick Start

### MCP

```bash
claude mcp add --scope user --transport stdio go-linear \
  --env LINEAR_API_KEY=lin_api_xxx \
  -- /path/to/go-linear mcp start
```

Then ask Claude: *"What are my urgent issues?"* or *"What did the team complete this week?"*

See [Claude Setup Guide](docs/CLAUDE-SETUP.md) for details.

### CLI

```bash
export LINEAR_API_KEY=lin_api_xxx

go-linear issue list --assignee=me --priority=1
go-linear user completed --team=ENG --completed-after=7d
go-linear issue create --team=ENG --title="Fix bug" --priority=2
```

#### Common Tasks

| I want to... | Command |
|--------------|---------|
| See my assigned issues | `issue list --assignee=me` |
| Find urgent issues | `issue list --priority=1` |
| Create a new issue | `issue create --team=ENG --title="Fix bug"` |
| Update issue status | `issue update ENG-123 --state=Done` |
| Add comment to issue | `comment create --issue=ENG-123 --body="Fixed"` |
| Link PR to issue | `issue update ENG-123 --link-pr=owner/repo#123` |
| Search issues | `issue search "authentication"` |
| List team members | `team members --team=ENG` |
| Bulk update issues | `issue batch-update --state=Triage --set-state=Backlog` |
| Get completion stats | `user completed --team=ENG --completed-after=7d` |
| Filter by multiple criteria | `issue list --team=ENG --priority=1 --state="In Progress"` |
| List with AI suggestions | `issue list --has-suggested-teams` |

See [CLI Quick Start](docs/CLI-QUICK-START.md) for all commands and [Filters Guide](docs/FILTERS.md) for filter options.

### SDK (Go Developers)

```go
client, _ := linear.NewClient("lin_api_xxx")
defer client.Close()

issues, _ := client.Issues(ctx, nil, nil)
for _, issue := range issues.Nodes {
    fmt.Printf("%s: %s\n", issue.Identifier, issue.Title)
}
```

See [SDK Documentation](docs/SDK.md) for production configuration.

---

## Design Philosophy

For AI agents, the CLI provides a semantic interface to Linear's GraphQL API, keeping context clean.

Working directly with GraphQL consumes an agent's **attention budget** on low-signal work:

- Schema exploration
- Query composition
- Pagination management
- Response parsing

This causes **context rot** - as tokens accumulate, accuracy degrades.

The CLI operates at the **right altitude**: high enough to hide GraphQL mechanics, specific enough to express precise intent:

```bash
go-linear issue list --team=ENG --priority=1
```

One command, one response. The agent's context stays **high-signal**:

| What the CLI absorbs | What the agent sees |
|----------------------|---------------------|
| UUID resolution | `--team=ENG` |
| 50+ field responses | 8 default fields |
| Pagination | Automatic, or `--count` for totals |
| Date parsing | `--created-after=7d` |

The CLI also ships with a [Claude skill](.claude/skills/go-linear/SKILL.md) (~10 KB) - decision guides and workflows that provide **just-in-time context** for complex operations.

The MCP server is auto-generated from the CLI using [ophis](https://github.com/njayp/ophis).

---

## Documentation

| Doc | Audience |
|-----|----------|
| [CLI Quick Start](docs/CLI-QUICK-START.md) | Terminal users |
| [Claude Setup](docs/CLAUDE-SETUP.md) | MCP/AI agent users |
| [SDK Documentation](docs/SDK.md) | Go developers |
| [Skill Reference](.claude/skills/go-linear/SKILL.md) | AI agents (loaded as context) |
| [Filters](docs/FILTERS.md) | Power users |
| [Migration](docs/MIGRATION.md) | Upgrading from v1.x |

---

## Code Generation

The SDK is generated from Linear's GraphQL schema:

```
upstream/linear (git submodule)
       ↓
   schema.graphql
       ↓
   gqlgenc (code generator)
       ↓
   internal/graphql/models.go (~30k lines)
   internal/graphql/client.go
       ↓
   pkg/linear/ (public API)
```

To regenerate after schema updates:

```bash
make sync-upstream  # Fetch latest schema + regenerate
```

---

## Project Info

**Status**: Stable. Semantic versioning.

**Not official**: This is a third-party client. Official Linear SDKs at https://github.com/linear

**Security**: mark.esler@chainguard.dev (72-hour acknowledgment). See [SECURITY.md](SECURITY.md).

**License**: Apache 2.0

**Built with**:

- [gqlgenc](https://github.com/Yamashou/gqlgenc) - GraphQL code generation
- [ophis](https://github.com/njayp/ophis) - Cobra to MCP server
- [clog](https://github.com/chainguard-dev/clog) - Structured logging
- [fido](https://github.com/codeGROOVE-dev/fido) - Tiered caching (memory + filesystem)
