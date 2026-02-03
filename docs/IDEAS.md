# Future Feature Ideas

Ideas and proposals for enhancing go-linear CLI and MCP.

---

## Aggregation Extensions

**Implemented:** --count flag

**Future ideas:**

### Group-by Aggregation
```bash
go-linear issue list --group-by=priority --output=json
# Returns: [{"priority": 1, "count": 5}, {"priority": 2, "count": 23}]
```

### Multi-dimensional Grouping
```bash
go-linear issue list --group-by=priority,state --output=json
# Returns matrix: priority × state with counts
```

### Statistical Aggregations
```bash
go-linear issue list --stats=estimate --output=json
# Returns: {"avg": 5.2, "min": 1, "max": 13, "median": 5}
```

### Time-series
```bash
go-linear issue list --group-by=created-date --interval=day
# Returns: issues per day
```

**Status:** Not blocked - requires client-side aggregation code
**Value:** Further token reduction, dashboard queries

---

## Triage Automation

**Implemented:**
- 44 filters including AI suggestions, batch-update
- Conditional auto-actions (via batch-update with filters)
- SLA-based actions (via --sla-status filter)

**Examples:**
```bash
# Batch approve with conditions
go-linear issue batch-update \
  --state=Triage \
  --label=bug \
  --priority=1 \
  --set-state=Backlog \
  --dry-run

# Auto-escalate SLA breaches
go-linear issue batch-update \
  --sla-status=Breached \
  --set-priority=1 \
  --add-label=urgent
```

**Future ideas:**

### Auto-apply AI Suggestions
```bash
go-linear issue batch-update \
  --has-suggested-teams \
  --apply-suggested-team \
  --dry-run
```

**Blocked:** Issue.suggestions API is marked [Internal] in schema - Linear doesn't expose what AI suggested, only that suggestions exist

---

## Batch Operation Extensions

**Implemented:** issue batch-update (up to 50 issues)

**Future ideas:**

### Batch Create
```bash
# Create multiple issues from template
go-linear issue batch-create \
  --template=bug-report \
  --count=5 \
  --team=Engineering
```

**Status:** Not blocked - `issueBatchCreate` mutation exists in schema

### Batch Delete
```bash
# Clean up test issues
go-linear issue batch-delete \
  --title="Test:" \
  --created-after=today \
  --dry-run
```

**Status:** Unknown if Linear API supports batch delete (not found in schema)

### Cross-entity Batch
```bash
# Update multiple entities
go-linear batch-update \
  --entity=issue \
  --filter="team:Old" \
  --set-team=New
```

**Status:** Not blocked - just needs implementation for other entities

---

## Context Optimization

**Implemented:** Field defaults, --count

**Future ideas:**

### Context Commands
```bash
# Return minimal context for AI caching
go-linear context me
# Returns: {"id": "...", "defaultTeam": "..."}

go-linear context teams
# Returns: {"teamIds": ["..."]}
```

### Convenience Commands
```bash
go-linear issue mine
# Equivalent to: issue list --assignee=me

go-linear team mine
# Returns default team from config
```

**Status:** Not blocked - just CLI sugar
**Value:** Reduce "my issues" from 2 calls to 1

---

## Advanced Filtering

**Implemented:** 44 filters

**Future ideas:**

### Compound Logic
```bash
go-linear issue list --filter='(team=X AND priority=1) OR (team=Y AND priority=2)'
```

**Status:** Not blocked - IssueFilter has `and: [IssueFilter!]` and `or: [IssueFilter!]` fields in schema. Needs CLI flag parsing and recursive filter building.

### Nested Collection Filters
```bash
# Issues with comments containing specific text by specific user
go-linear issue list \
  --comment-by=me \
  --comment-contains="approved" \
  --comment-after=7d
```

**Status:** Partially blocked - comment collection filter doesn't have date fields in schema. Would need client-side filtering.

Currently: comment-by AND comment-contains work together

### Query Builder
```bash
# Save complex queries
go-linear query save urgent-bugs \
  --priority=1 \
  --label=bug \
  --state=Backlog

go-linear query run urgent-bugs
```

**Status:** Not blocked - just CLI config feature

---

## Smart Defaults

**Implemented:** User config (~/.config/linear/config.yaml)

**Future ideas:**

### Team/Workspace Defaults
```yaml
# .linear-workspace.yaml (checked into repo)
defaults:
  team: Engineering
  labels: [bug, feature]

filters:
  active_issues: --state=In Progress --assignee=me
```

**Status:** Not blocked - just config file extension

### AI-learned Defaults
```bash
# CLI learns from usage patterns
go-linear learn --from-history
# Adjusts defaults based on most-used filters
```

**Status:** Not blocked - analyze shell history or config usage

---

## Integration Ideas

### Sentry Integration
```bash
go-linear issue create \
  --from-sentry=<event-id> \
  --auto-populate
```

**Status:** Not blocked - fetch from Sentry API, populate issue fields

### GitHub PR Linking
```bash
go-linear issue update ENG-123 \
  --link-pr=owner/repo#123
```

**Status:** Not blocked - use attachmentLinkGitHubPR mutation (already in SDK)

### Slack Integration
```bash
go-linear issue create \
  --from-slack-thread=<url> \
  --title-from-first-message
```

**Status:** Not blocked - fetch from Slack API, populate issue fields

---

## Analysis & Metrics

### Triage Metrics
```bash
go-linear triage summary --period=7d
# Returns: acceptance rate, avg time, top sources
```

### Team Velocity
```bash
go-linear team velocity --team=ENG --period=sprint
# Returns: completed points, cycle time, etc.
```

### SLA Dashboard
```bash
go-linear sla dashboard --output=json
# Returns: breaches, at-risk, by team
```

**Status:** Not blocked - client-side aggregation over filtered issue lists
**Value:** Significant token reduction vs manual aggregation

