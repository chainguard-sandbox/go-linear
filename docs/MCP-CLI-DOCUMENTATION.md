# MCP CLI Documentation Strategy

## Core Insight

MCP servers load ALL tool descriptions once per session and keep them in the AI agent's context permanently. With 72 tools, verbose documentation consumes thousands of tokens that could be used for actual conversation.

**Our approach:** System-wide coherence and minimal error output.

### Optimizations for AI Agents

1. **Concise tool descriptions** - Tools complement each other, not repeat information
2. **Silent usage on errors** - No help text dumped on every error (Cobra `SilenceUsage: true`)
3. **Clean error messages** - Extract user-friendly messages from verbose API responses

---

## Philosophy

### 1. Explain Once, Reference Everywhere

**Don't:**
```
issue_create: "Priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low"
issue_update: "Priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low"
issue_list:   "Priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low"
```
Total: ~150 chars × 3 = 450 chars

**Do:**
```
issue_create: "Priority (0=none, 1=urgent, 2=high, 3=normal, 4=low)"  [CANONICAL]
issue_update: "Priority (0-4, see issue_create)"                      [REFERENCE]
issue_list:   "Priority (0-4)"                                        [MINIMAL]
```
Total: ~140 chars (69% savings)

### 2. Core Tools Teach Patterns

**Foundation tools** (teach reusable concepts):
- `issue_list` - Teaches: filtering, 'me' keyword, date formats, pagination
- `issue_create` - Teaches: priority scale, required vs optional
- `team_list` - Teaches: team discovery, team keys
- `user_list` - Teaches: user discovery

**Dependent tools** reference foundation tools:
- `issue_update` → "priority (0-4, see issue_create)"
- `cycle_create` → "date formats (see issue_list)"
- `comment_create` → "issue ID (from issue_list)"

### 3. Show, Don't Tell

**Instead of:**
```
--completed-after: Supports ISO8601 (2025-12-10), relative (yesterday, today), or duration (7d, 2w, 3m)
```

**Use:**
```
--created-after=yesterday|7d|2025-12-10
```

AI agents learn formats from examples. No need to explain "supports" or enumerate all options.

---

## Documentation Pattern

### Structure Template

```
{One-line summary}. Returns {N} default fields. {Key feature}.

{Most important capability}: {concise syntax}

Example: go-linear {entity} {action} {key-flags} --output=json

{Returns/Count/Special}: {output structure}
Related: {entity}_{related-actions}
```

### Field Count

Always state field count (helps AI predict response size):
- "Returns 8 default fields per issue"
- "Returns 7 default fields per team (includes issueCount)"

### Examples

One comprehensive example showing multiple features:
```
Example: go-linear issue list --team=ENG --assignee=me --priority=1 --output=json
```

This ONE example teaches:
- Team filtering with keys
- 'me' keyword
- Priority values
- JSON output

### Related Commands

Concise references (AI can check their descriptions):
```
Related: issue_get, issue_create, team_list
```

Not:
```
Related Commands:
  - linear issue get - Get single issue details
  - linear issue create - Create new issue
```

---

## Token Budget

### Target Sizes

**Simple commands** (get, delete): 100-200 chars
```
Get {entity} by {id-type}. Returns {N} default fields.

Example: go-linear {entity} get <id> --output=json

Related: {entity}_list, {entity}_update
```

**Medium commands** (list, create, update): 250-400 chars
```
{Action} {entity}. Returns {N} default fields. {Key feature}.

{Primary capability}: {concise list}

Example: go-linear {entity} {action} {key-example} --output=json

Related: {comma-separated}
```

**Complex commands** (issue_list, user_completed): 400-600 chars
- Worth extra tokens to explain properly
- Teach patterns other commands reference
- Multiple capabilities documented

### Results Achieved

- **Before:** 47,673 chars (~11,918 tokens)
- **After:** 15,121 chars (~3,780 tokens)
- **Reduction:** 69% (8,138 tokens saved per MCP session)

---

## Guidelines for New Commands

### 1. Identify Command Complexity

**Simple** (read-only, single resource):
- get, delete commands
- No new patterns to teach
- Minimal docs (~150 chars)

**Medium** (CRUD operations):
- create, update, list commands
- May teach new patterns
- Moderate docs (~300 chars)

**Complex** (multi-step workflows):
- Aggregations, cross-entity queries
- Teach important patterns
- Detailed docs (~500 chars)

### 2. Check for Redundancy

Before documenting a pattern, ask:
1. Does another command already explain this?
2. Can I reference that command instead?
3. If I must explain, is this the canonical place?

### 3. Example First

Lead with the example, not explanations:
```
Example: go-linear issue list --team=ENG --priority=1 --count
```

The example shows:
- How to use flags
- What values to pass
- Output format

No need to say "you can filter by team" - the example shows it.

### 4. Preserve Useful Information

**Keep:**
- Unique capabilities of this command
- Parameter formats by example
- Output structure if non-standard
- Cross-references for discovery

**Remove:**
- Generic framing ("This operation creates...")
- Repeated pattern explanations
- Multiple examples showing same concept
- Verbose parameter descriptions (flags have their own descriptions)

---

## Pattern Examples

### Canonical Tool (Teaches Pattern)

**issue_list** (533 chars):
```
List issues with filtering. Returns 8 default fields per issue. Use --count for totals (99% fewer tokens).

Filters: --team (name/key), --assignee=me (current user) or email, --state (name), --priority (0-4), --label (name), --created-after=yesterday|7d|2w|2025-12-10

Pagination: --limit (default 50), --after (cursor from pageInfo.endCursor)
Count: --count returns just {"count": N} instead of full results

Example: go-linear issue list --team=ENG --assignee=me --priority=1 --output=json

Returns: {nodes: [{8 issue fields}...], pageInfo: {hasNextPage, endCursor}} or {"count": N} with --count
Related: issue_get, issue_create, team_list, user_list
```

**Teaches:**
- 'me' keyword for current user
- Date formats (yesterday, 7d, 2w, ISO8601)
- Pagination structure
- --count flag pattern
- Filter composition

### Referencing Tool

**issue_update** (220 chars):
```
Update issue. Modifies existing data.

Fields: --title, --description, --assignee=me, --state, --priority (0-4, see issue_create), --add-label, --remove-label

Example: go-linear issue update ENG-123 --assignee=me --priority=1 --state=Done --output=json

Related: issue_get, issue_create
```

**References:**
- Priority scale → "see issue_create"
- 'me' keyword → just uses it (taught in issue_list)

### Simple Tool

**issue_delete** (120 chars):
```
⚠️ Delete issue. Cannot be undone. Prompts unless --yes.

Example: go-linear issue delete ENG-123

Related: issue_list, issue_get
```

**Minimal:** No new patterns, just essential info.

---

## Maintenance Checklist

When adding a new command:

1. **Categorize:** Simple / Medium / Complex
2. **Check redundancy:** Does this pattern exist elsewhere?
3. **Document appropriately:**
   - Simple: ~150 chars
   - Medium: ~300 chars
   - Complex: ~500 chars
4. **Test final size:** Ensure no bloat
5. **Verify MCP export:** `make build && ./bin/go-linear mcp tools`
6. **Measure:** Check total description size stays reasonable

---

## Error Handling for AI Agents

### Problem: Verbose Errors Bloat Context

**Before optimization:**
```
Error: LINEAR_API_KEY environment variable or --api-key flag required
Usage:
  go-linear issue list [flags]

Flags:
      --added-to-cycle-after string ...
      --added-to-cycle-before string ...
      [80+ lines of help text]
```

**After optimization:**
```
Error: LINEAR_API_KEY environment variable or --api-key flag required
```

### Implementation

**1. Silence usage on errors** (`cmd/linear/commands/root.go`):
```go
rootCmd := &cobra.Command{
    Use: "go-linear",
    SilenceUsage: true, // Don't print usage on errors
    // ...
}
```

**2. Clean API error messages** (`pkg/linear/errors.go`):
```go
// Before: {"networkErrors":{"code":401,"message":"Response body {...}"},...}
// After: linear: AuthenticationError: invalid or expired API key

func wrapGraphQLError(operation string, err error) error {
    errStr := err.Error()
    if strings.Contains(errStr, `"code":401`) {
        return &AuthenticationError{
            LinearError: &LinearError{
                Message: "invalid or expired API key",
                // ...
            },
        }
    }
    // ...
}
```

**Result:**
- Missing API key: 2 lines (was 80+ lines)
- Invalid API key: 1 line (was 500+ characters of JSON)
- Other errors: Operation context only, no verbose JSON

---

## Anti-Patterns

### ❌ DON'T: Repeat Explanations

```
issue_create: "Date formats: ISO8601 (2025-12-10), relative (yesterday), duration (7d)"
cycle_create: "Date formats: ISO8601 (2025-12-10), relative (yesterday), duration (7d)"
```

### ✅ DO: Reference or Show

```
issue_create: "--created-after=yesterday|7d|2025-12-10"  [TEACH BY EXAMPLE]
cycle_create: "Date formats: see issue_list"             [REFERENCE]
```

### ❌ DON'T: Multiple Examples

```
Examples:
  go-linear issue list --team=ENG
  go-linear issue list --priority=1
  go-linear issue list --team=ENG --priority=1 --output=json
```

### ✅ DO: One Comprehensive Example

```
Example: go-linear issue list --team=ENG --priority=1 --output=json
```

### ❌ DON'T: Verbose Sections

```
Parameters:
  --team: Team name (e.g., "Engineering") or UUID - use 'linear team list' to discover names
  --assignee: User email, name, or 'me' - use 'linear user list' to discover users

Common Errors:
  - "team not found": Check spelling or use 'linear team list'
```

### ✅ DO: Inline and Concise

```
Filters: --team (name/key from team_list), --assignee=me or email
```

---

## Measuring Success

### Token Metrics

**Per command average:**
- Before: ~670 chars (~168 tokens)
- After: ~213 chars (~53 tokens)
- Reduction: 68%

**Total (71 tools):**
- Before: 47,673 chars (~11,918 tokens)
- After: 15,121 chars (~3,780 tokens)
- Reduction: 69% (~8,138 tokens saved)

### Quality Metrics

**Can AI agents:**
- ✅ Discover the right tool for the job?
- ✅ Understand what it does?
- ✅ Construct correct queries?
- ✅ Learn patterns from core tools?
- ✅ Navigate to related tools?

If yes to all: Documentation is sufficient.

---

## Real-World Impact

**Example query:** "How many urgent bugs in Engineering?"

**Without optimization:**
- MCP context: ~12,000 tokens (all 71 tool descriptions)
- Query result: ~500 tokens (50 issues with details)
- Total: ~12,500 tokens

**With optimization:**
- MCP context: ~3,780 tokens (optimized descriptions)
- Query result: ~4 tokens (`{"count": 5}`)
- Total: ~3,784 tokens

**Savings: 70% (~8,716 tokens)**

---

## Future Additions

When adding aggregation to other entities:

**Priority:**
1. ✅ `issue list --count` (done)
2. ✅ `issue search --count` (done)
3. ✅ `team members --count` (done)
4. Consider: `comment list --count`, `label list --count`, `cycle list --count`

**Pattern to follow:**
- Add --count flag to existing list/search commands
- Don't create standalone count commands
- Document in existing tool description (1-2 lines)
- Reference canonical documentation (issue_list)

---

## Key Principle

**Every token in MCP context is permanent overhead for the entire session.**

Optimize ruthlessly while preserving:
- Unique capabilities
- Discovery paths (related commands)
- Pattern teaching (in canonical tools only)
- Essential examples

When in doubt: Can the AI use this tool correctly with just this info? If yes, it's sufficient.
