# MCP Architecture: ophis Integration

This document explains the MCP (Model Context Protocol) integration architecture and the decision to use the [ophis](https://github.com/njayp/ophis) library.

## Overview

go-linear exposes its CLI commands as MCP tools, allowing AI assistants (Claude, Cursor, VSCode Copilot) to interact with Linear programmatically.

## Why ophis?

We chose ophis for MCP server generation because it:

1. **Zero-modification integration**: Wraps existing Cobra commands without changes to command implementations
2. **Automatic schema generation**: Converts flags/arguments to JSON Schema for MCP tool definitions
3. **Selector-based filtering**: Fine-grained control over which commands become tools
4. **Multiple transport support**: stdio (for Claude) and HTTP (for other clients)

### Alternatives Considered

| Library | Pros | Cons |
|---------|------|------|
| **ophis** | Zero-touch CLI wrapping, mature Cobra support | Subprocess per call (no shared state) |
| **go-sdk/mcp** (direct) | Full control, single process | Requires rewriting all commands as MCP handlers |
| **Custom wrapper** | Full control | Significant development effort |

## Architecture Details

### Subprocess Model

ophis spawns a **new subprocess** for each MCP tool call:

```
AI Assistant → MCP Protocol → ophis server (long-lived)
                                    ↓
                          Each tool call spawns:
                          exec.Command("go-linear", "issue", "list", ...)
                                    ↓
                          CLI subprocess runs, exits
```

This design has tradeoffs:

**Pros:**
- Clean isolation between calls
- No memory leaks across calls
- Works with any CLI without modification

**Cons:**
- No in-memory state sharing between calls
- Requires file-based caching for performance (see below)

### Caching Strategy

Because each MCP call spawns a new process, we use [fido](https://github.com/codeGROOVE-dev/fido) for the resolver:

- **TieredCache**: memory + local filesystem for optimal performance
- Location: `~/.cache/go-linear/resolver/` (sharded directory structure)
- TTL: 5 minutes
- Contents: name→UUID mappings for teams, users, states, labels, projects

This allows repeated MCP calls like `issue list --team=Engineering` to resolve "Engineering" to its UUID without API calls. The tiered approach provides fast in-memory lookups within a single CLI invocation while persisting to disk for cross-process cache hits.

## Stability Assessment

### ophis Package Maturity

As of December 2025:
- GitHub: [njayp/ophis](https://github.com/njayp/ophis)
- License: MIT
- Active development
- Used by: go-linear (this project)

**Risk Assessment:** Medium
- Small community (single maintainer)
- API is simple and stable
- Dependencies are minimal (Cobra, go-sdk/mcp)

### Migration Path

If ophis becomes unmaintained, migration options:

1. **Fork and maintain**: ophis is ~500 LOC, easily maintainable
2. **Direct go-sdk/mcp**: Rewrite commands as MCP handlers (~2-3 days effort)
3. **Alternative wrapper**: Evaluate new libraries as ecosystem matures

### Vendor Lock-in Mitigation

- All business logic is in `pkg/linear` (pure Go, no MCP dependency)
- CLI commands are standard Cobra (work independently of MCP)
- MCP integration is isolated to `cmd/linear/commands/mcp/` and ophis config
- Switching MCP implementations affects only the wrapper layer

## Configuration

MCP is configured via ophis in `cmd/linear/main.go`:

```go
ophis.Merge(rootCmd, &ophis.Config{
    Selectors: []ophis.Selector{{
        CmdSelector: ophis.AllowCmdsContaining("get", "list", "search", "create", "update", "delete"),
        // ... flag selectors
    }},
})
```

See the CLI commands for usage details.
