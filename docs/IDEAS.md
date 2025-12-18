# Future Feature Ideas

Ideas and proposals for enhancing go-linear CLI and MCP.

---

## Aggregation Extensions

**Implemented:** --count flag (99% token reduction)

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

**Value:** Further token reduction, dashboard queries

---

## Triage Automation

**Implemented:** 44 filters including AI suggestions, batch-update

**Future ideas:**

### Auto-apply AI Suggestions
```bash
go-linear issue batch-update \
  --has-suggested-teams \
  --apply-suggested-team \
  --dry-run
```

Requires: Access to Issue.suggestions API

### Conditional Auto-actions
```bash
# Auto-accept with conditions
go-linear issue batch-update \
  --label=bug \
  --priority=1 \
  --set-state=Backlog \
  --set-assignee=oncall
```

### SLA-based Actions
```bash
# Auto-escalate SLA breaches
go-linear issue batch-update \
  --sla-status=Breached \
  --set-priority=1 \
  --add-label=urgent
```

**Value:** Reduce manual triage work

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

### Batch Delete
```bash
# Clean up test issues
go-linear issue batch-delete \
  --title="Test:" \
  --created-after=today \
  --dry-run
```

### Cross-entity Batch
```bash
# Update multiple entities
go-linear batch-update \
  --entity=issue \
  --filter="team:Old" \
  --set-team=New
```

**Value:** More bulk operations

---

## Context Optimization

**Implemented:** Field defaults (80% reduction), --count (99% reduction)

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

**Value:** Reduce "my issues" from 2 calls to 1

---

## Advanced Filtering

**Implemented:** 44 filters

**Future ideas:**

### Compound Logic
```bash
go-linear issue list --filter='(team=X AND priority=1) OR (team=Y AND priority=2)'
```

### Nested Collection Filters
```bash
# Issues with comments containing specific text by specific user
go-linear issue list \
  --comment-by=me \
  --comment-contains="approved" \
  --comment-after=7d
```

Currently: comment-by AND comment-contains work together
Future: Add comment-after date filtering

### Query Builder
```bash
# Save complex queries
go-linear query save urgent-bugs \
  --priority=1 \
  --label=bug \
  --state=Backlog

go-linear query run urgent-bugs
```

**Value:** Reusable complex queries

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

### AI-learned Defaults
```bash
# CLI learns from usage patterns
go-linear learn --from-history
# Adjusts defaults based on most-used filters
```

**Value:** Smarter defaults over time

---

## Integration Ideas

### Sentry Integration
```bash
go-linear issue create \
  --from-sentry=<event-id> \
  --auto-populate
```

### GitHub PR Linking
```bash
go-linear issue update ENG-123 \
  --link-pr=owner/repo#123
```

### Slack Integration
```bash
go-linear issue create \
  --from-slack-thread=<url> \
  --title-from-first-message
```

**Value:** Streamlined integrations

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

**Value:** 95% token reduction vs manual aggregation

---

## Sources

Ideas synthesized from:
- AGGREGATION.md - Aggregation design proposals
- CONTEXT-SAVERS.md - Token optimization analysis
- TRIAGE-AUTOMATION.md - Triage workflow proposals
- Session discussions

---

*Document created: 2025-12-17*
