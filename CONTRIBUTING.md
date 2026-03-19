# Contributing to go-linear

## Setup

```bash
git clone https://github.com/chainguard-sandbox/go-linear
cd go-linear
make dev    # installs tools + downloads deps
```

## Testing

| Tier | Command | API key | What it tests |
|------|---------|---------|---------------|
| Mock | `make test` | No | Unit tests, filters, parsing |
| Read | `make test-read` | Read | Live queries against Linear |
| Write | `make test-write` | Write | Creates/updates/deletes real data |

Mock tests run in CI. Read and write tests require `LINEAR_API_KEY`.

## Code Style

```bash
make check   # fmt + vet + lint + test + tidy
```

This is the same gate CI runs. Fix any issues before opening a PR.

## Schema Sync

The SDK is generated from Linear's upstream GraphQL schema:

```bash
git submodule update --init upstream
make sync-upstream   # fetch latest schema, regenerate code, run tests
```

Review the diff carefully — new types are added to `internal/graphql/models.go` and `schema.graphql`. Verify that removed or renamed upstream types don't break existing queries in `queries/`.

## MCP Tool Documentation

Every token in MCP tool descriptions is permanent overhead for the entire session. Optimize ruthlessly.

### Principles

1. **Explain once, reference everywhere** — define concepts in one canonical tool, reference from others:
   ```
   issue_create: "Priority (0=none, 1=urgent, 2=high, 3=normal, 4=low)"  [CANONICAL]
   issue_update: "Priority (0-4, see issue_create)"                      [REFERENCE]
   issue_list:   "Priority (0-4)"                                        [MINIMAL]
   ```

2. **Core tools teach patterns** — foundation tools define reusable concepts:
   - `issue_list` — filtering, `me` keyword, date formats, pagination, `--count`
   - `issue_create` — priority scale, required vs optional fields
   - `team_list` / `user_list` — entity discovery

   Dependent tools reference them: `"date formats (see issue_list)"`

3. **Show, don't tell** — lead with examples, not explanations:
   ```
   --created-after=yesterday|7d|2025-12-10
   ```
   Not: "Supports ISO8601, relative dates (yesterday, today), or durations (7d, 2w, 3m)"

4. **Silent on errors** — no help text on failure (Cobra `SilenceUsage: true`), clean error messages instead of raw API JSON

### Size Targets

| Complexity | Example | Target |
|------------|---------|--------|
| Simple (get, delete) | `issue_get`, `issue_delete` | ~150 chars |
| Medium (list, create, update) | `issue_update`, `cycle_create` | ~300 chars |
| Complex (multi-capability) | `issue_list`, `user_completed` | ~500 chars |

### Template

```
{One-line summary}. Returns {N} default fields.

{Key capabilities}: {concise inline syntax}

Example: go-linear {entity} {action} {key-flags}
Related: {entity}_{related-actions}
```

### Anti-Patterns

**Don't** repeat explanations across tools — reference or show by example.
**Don't** use multiple examples showing the same concept — one comprehensive example.
**Don't** use verbose parameter sections — inline: `--team (name/key from team_list)`.

## Pull Requests

1. Fork and branch from `main`
2. Keep changes focused — one concern per PR
3. Add tests for new functionality (mock tier at minimum)
4. Run `make check` before pushing
5. PR description should explain *why*, not just *what*

## Security

Report vulnerabilities privately to mark.esler@chainguard.dev. See [SECURITY.md](SECURITY.md).
