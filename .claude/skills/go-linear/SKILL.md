---
name: go-linear
description: Linear issue tracking CLI with smart filtering and aggregation (project)
---

# go-linear MCP Skill

Linear CLI for AI agents. 74 tools across 17 entities. Requires `LINEAR_API_KEY` env var.

## Decision Guide

| Goal | Tool | Key flags |
|------|------|-----------|
| Count issues | `issue list` | `--team=ENG --count` |
| Find by criteria | `issue list` | `--assignee=me --state="In Progress"` |
| Find by text | `issue search` | `"auth bug"` |
| Get one issue | `issue get` | `ENG-123` |
| Get comments for issue | `comment list` | `--issue=ENG-123` |
| Who completed work? | `user completed` | `--team=ENG --completed-after=7d` |
| Team workload | `team get` | `ENG` (returns issueCount) |
| Rate limits | `status` | |

## Core Patterns

### Filtering (names, not UUIDs)
```
--team=Engineering       # or team key: --team=ENG
--assignee=me            # or email: --assignee=alice@example.com
--creator=me             # who created it
--state="In Progress"    # workflow state name
--priority=1             # 0=none, 1=urgent, 2=high, 3=normal, 4=low
--label=bug              # repeatable: --label=bug --label=security
```

### Date Filters
```
--created-after=yesterday
--created-after=7d       # 7 days ago
--created-after=2w       # 2 weeks ago
--created-after=2025-01-15
--completed-before=today
```

### Output Optimization
```
--output=json            # ALWAYS use for parsing
--count                  # Returns {"count": N} - 99% fewer tokens
--fields=defaults        # 8 fields vs 50+ (80% reduction)
--fields=id,title,url    # Specific fields only
--fields=defaults,labels # Add to defaults
```

### Pagination
```
--limit=50               # Default, max varies
--after=<cursor>         # From pageInfo.endCursor
```

## Discovery Commands

| Need | Command |
|------|---------|
| Team names/keys | `team list --output=json` |
| User emails | `user list --output=json` |
| Workflow states | `state list --output=json` |
| State types | `state list --type=started` (triage/backlog/unstarted/started/completed/canceled) |
| Labels | `label list --output=json` |
| Labels by team | `label list --team=ENG --output=json` |
| Projects | `project list --output=json` |
| Active cycles | `cycle list --active --output=json` |
| Current user | `viewer --output=json` |
| Workspace | `organization --output=json` |
| Rate limits | `status --output=json` |

## Workflows

### Issue Investigation
```bash
# Get issue details
go-linear issue get ENG-123 --output=json

# Get comments on issue
go-linear comment list --issue=ENG-123 --output=json

# Check sub-issues
go-linear issue list --parent=ENG-123 --output=json
```

### Triage Check
```bash
# Step 1: Count
go-linear issue list --state=Triage --team=ENG --count

# Step 2: If count > 0, get details
go-linear issue list --state=Triage --team=ENG --output=json
```

### Team Productivity
```bash
# Who completed work this week?
go-linear user completed --team=ENG --completed-after=7d --output=json
# Returns: [{user: {name, email}, count: N}, ...]

# Single user
go-linear user completed --user=me --completed-after=yesterday --output=json
```

### Create Issue
```bash
go-linear issue create \
  --team=ENG \
  --title="Fix auth bug" \
  --priority=1 \
  --assignee=me \
  --label=bug \
  --output=json
```

### Update Issue
```bash
# Change state and priority
go-linear issue update ENG-123 --state=Done --priority=2 --output=json

# Add/remove labels
go-linear issue update ENG-123 --add-label=reviewed --remove-label=needs-review

# Make sub-issue (set parent)
go-linear issue update ENG-456 --parent=ENG-123

# Remove from parent (use 'none')
go-linear issue update ENG-456 --parent=none
```

### Batch Update (max 50)
```bash
# Preview first (--dry-run)
go-linear issue batch-update \
  --state=Triage \
  --has-suggested-teams \
  --set-state=Backlog \
  --dry-run

# Apply (--yes skips confirmation)
go-linear issue batch-update \
  --state=Triage \
  --has-suggested-teams \
  --set-state=Backlog \
  --yes
```

### Link GitHub PR
```bash
# Via attachment
go-linear attachment link-github \
  --issue=ENG-123 \
  --url=https://github.com/owner/repo/pull/42 \
  --output=json

# Or via issue update
go-linear issue update ENG-123 --link-pr=owner/repo#42
```

### Link External Resources
```bash
# Link URL
go-linear attachment link-url \
  --issue=ENG-123 \
  --url=https://docs.example.com/spec \
  --title="Design Spec" \
  --output=json

# Link Slack (requires Slack integration)
go-linear attachment link-slack \
  --issue=ENG-123 \
  --url=https://workspace.slack.com/archives/C123/p1234567890 \
  --output=json

# Custom attachment with metadata
go-linear attachment create \
  --issue=ENG-123 \
  --title="Build #42" \
  --url=https://ci.example.com/42 \
  --metadata='{"status":"passed"}' \
  --output=json
```

### Issue Relationships
```bash
# Create relationship
go-linear issue relate ENG-123 ENG-124 --type=blocks
# Types: blocks, blocked-by, duplicate, related

# Find blocking issues
go-linear issue list --has-blocking --team=ENG --output=json

# Find blocked issues
go-linear issue list --has-blocked-by --team=ENG --output=json
```

### Comments
```bash
# Add comment (supports markdown)
go-linear comment create --issue=ENG-123 --body="Fixed in PR #42" --output=json

# List comments for issue
go-linear comment list --issue=ENG-123 --output=json

# List my recent comments
go-linear comment list --creator=me --created-after=7d --output=json
```

### Reactions
```bash
# React to issue
go-linear reaction create --issue=ENG-123 --emoji=👍 --output=json

# React to comment
go-linear reaction create --comment=<uuid> --emoji=🎉 --output=json
```

### Favorites (Stars)
```bash
# Star issue
go-linear favorite create --issue=ENG-123 --output=json

# Star project
go-linear favorite create --project="Q1 Platform" --output=json
```

### Cycles (Sprints)
```bash
# List active cycles
go-linear cycle list --active --team=ENG --output=json

# Create cycle
go-linear cycle create \
  --team=ENG \
  --starts-at=2025-01-15 \
  --ends-at=14d \
  --name="Sprint 42" \
  --output=json

# Assign issue to cycle
go-linear issue update ENG-123 --cycle=<cycle-uuid>

# Remove from cycle
go-linear issue update ENG-123 --cycle=none
```

### Projects & Milestones
```bash
# List projects
go-linear project list --health=onTrack --output=json

# Create milestone
go-linear project milestone-create \
  --project="Q1 Platform" \
  --name="Alpha Release" \
  --target-date=2025-02-15 \
  --output=json
```

### Notifications
```bash
# Subscribe to project updates
go-linear notification subscribe --project="Q1 Platform" --output=json

# Subscribe to team updates
go-linear notification subscribe --team=ENG --output=json
```

## Advanced Filters (issue list)

### AI Triage Suggestions
```bash
--has-suggested-teams      # AI suggested team transfer
--has-suggested-assignees  # AI suggested assignee
--has-suggested-labels     # AI suggested labels
--has-suggested-projects   # AI suggested project
```

### Relationships
```bash
--has-children             # Has sub-issues (is parent)
--has-blocking             # Blocks other issues
--has-blocked-by           # Blocked by other issues
--has-duplicate            # Has duplicates
--has-related              # Has related issues
--parent=ENG-123           # Sub-issues of specific parent
```

### People
```bash
--comment-by=me            # Issues I commented on
--comment-by=alice@co.com  # Issues they commented on
--subscriber=me            # Issues I'm subscribed to
--attachment-by=me         # Issues I attached files to
--delegate=me              # Delegated to agent
```

### SLA & Customer
```bash
--sla-status=Breached      # SLA status: Breached/HighRisk/MediumRisk/OnTrack
--has-needs                # Has customer needs
--customer-count=5         # Affected customer count
```

### Text Search
```bash
--title="auth"             # Title contains
--description="login"      # Description contains
--comment-contains="fixed" # Comments contain
```

## Error Recovery

| Error | Solution |
|-------|----------|
| "team not found" | Run `team list` to discover valid names/keys |
| "ambiguous match: found X teams" | Use team key (ENG) not full name |
| "user not found" | Run `user list` to discover emails |
| "state not found" | Run `state list` to discover state names |
| "must specify either --user or --team" | Flags are mutually exclusive |
| Rate limited | Check `status`, wait and retry |
| "LINEAR_API_KEY required" | Set environment variable |

## Tool Categories

### Safe (Read-Only) - Execute freely
`*_list`, `*_get`, `*_search`, `status`, `viewer`, `organization`

### Creates - Safe, reversible via delete
`*_create`, `*_relate`, `*_link-*`, `*_add-label`, `*_subscribe`

### Updates - Modifies data
`*_update`, `*_remove-label`, `*_archive`, `*_unsubscribe`

### Deletes - Permanent, confirm with user
`*_delete`, `*_unrelate`

## Output Structures

### List commands
```json
{
  "nodes": [{...}, {...}],
  "pageInfo": {"hasNextPage": true, "endCursor": "cursor..."}
}
```

### With --count
```json
{"count": 23}
```

### Get commands
```json
{"id": "...", "identifier": "ENG-123", "title": "...", ...}
```

### Mutations
```json
{"success": true, "id": "...", ...}
```

## Efficiency Tips

1. **Start with --count** - Know scope before fetching details
2. **Use --fields=defaults** - 80% token reduction
3. **user completed** - Replaces 5-step workflow
4. **batch-update** - Update 50 issues in one call
5. **Always --output=json** - Machine-readable output
6. **Use team keys** - `ENG` not `Engineering` to avoid ambiguity
