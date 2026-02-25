# Migration Guide: v1.x to v2.0

## CLI: `--output` flag removed (JSON-only)

All commands now output JSON exclusively. The `--output` flag no longer exists.

### What changed

| v1.x | v2.0 |
|------|------|
| `--output=json` (opt-in) | JSON is the default and only format |
| `--output=table` (default) | Removed entirely |
| `--output=invalid` returns error | No `--output` flag to validate |

### Migration steps

**1. Remove `--output=json` from scripts**

The flag is no longer recognized. Remove it from all invocations:

```bash
# Before
go-linear issue list --team=ENG --output=json

# After
go-linear issue list --team=ENG
```

**2. Remove `--output=table` usage**

Table output is gone. Use `jq` for human-readable formatting:

```bash
# Before
go-linear issue list --output=table

# After (compact JSON)
go-linear issue list

# After (human-readable)
go-linear issue list | jq .

# After (table-like with jq)
go-linear issue list | jq -r '.nodes[] | "\(.identifier)\t\(.title)\t\(.state.name)"'
```

**3. Update error handling for output validation**

If your code checked for "unsupported output format" errors, remove that handling. The error can no longer occur.

**4. `--fields` flag is unchanged**

Field filtering works exactly as before:

```bash
go-linear issue get ENG-123 --fields=id,title,url
go-linear issue list --fields=defaults,estimate
go-linear issue get ENG-123 --fields=none   # all fields
```

### Why this change

- **Consistency**: Table output varied across commands (some used `tablewriter`, most used inline `fmt.Fprintf`). JSON is uniform.
- **Maintenance**: ~4,000 lines of table formatting code removed. Every new command required both formats.
- **Agents**: AI agents always use JSON. Table output was never consumed programmatically.
- **Simplicity**: One output format means predictable behavior. No format negotiation, no format-specific bugs.

## SDK: `IssueDelete` signature change

`IssueDelete` now accepts an optional `permanentlyDelete` parameter:

```go
// Before
err := client.IssueDelete(ctx, id)

// After
err := client.IssueDelete(ctx, id, nil)           // trash (default, 30-day grace)
err := client.IssueDelete(ctx, id, boolPtr(true))  // permanent, no grace period
```

Pass `nil` to preserve v1.x behavior (trash with 30-day grace period).

## Internal changes (library consumers)

If you imported internal packages:

| v1.x | v2.0 |
|------|------|
| `cli.OutputFlags{Output, Fields}` | `cli.FieldFlags{Fields}` |
| `cli.OutputOnlyFlags{Output}` | Deleted |
| `cli.OutputFlags.Validate()` | Deleted |
| `cli.OutputFlags.Bind(cmd, help)` | `cli.FieldFlags.Bind(cmd, help)` (no `--output` flag) |
| `formatter.FormatIssuesTable(...)` | Deleted |
| `formatter.FormatTeamsTable(...)` | Deleted |
| `formatter.FormatUsersTable(...)` | Deleted |
| `formatter.FormatCSV(...)` | Deleted |
| `formatter.FormatJSON(...)` | Unchanged |
| `formatter.FormatJSONFiltered(...)` | Unchanged |
