# Filtering Guide

## Overview

go-linear supports filtering across multiple entity types. Date formats and relative dates work consistently across all entities.

### Date Format Reference
All date filters support:
- ISO8601: `2025-01-15`, `2025-01-15T10:30:00Z`
- Relative: `yesterday`, `today`, `tomorrow`
- Duration: `7d` (7 days ago), `2w` (2 weeks ago), `30d`

---

## Issue Filters

`go-linear issue list` supports 64 filter flags across 44 filter types, enabling precise issue queries for production environments.

### Filter Categories

#### Core Filters (6)
- `--team` - Team name, key, or UUID (e.g., "Engineering", "ENG")
- `--assignee` - User email, name, or 'me'
- `--state` - State name or UUID (e.g., "In Progress")
- `--priority` - 0=none, 1=urgent, 2=high, 3=normal, 4=low
- `--label` - Label names (repeatable, OR logic)
- `--creator` - Who created (email, name, or 'me')

#### Date Filters (17 ranges, 34 flags)
All support relative dates (yesterday, 7d, 2w) and ISO8601.

- `--created-after`, `--created-before`
- `--updated-after`, `--updated-before`
- `--completed-after`, `--completed-before`
- `--archived-after`, `--archived-before`
- `--auto-archived-after`, `--auto-archived-before`
- `--auto-closed-after`, `--auto-closed-before`
- `--canceled-after`, `--canceled-before`
- `--started-after`, `--started-before`
- `--triaged-after`, `--triaged-before`
- `--due-after`, `--due-before`
- `--snoozed-until-after`, `--snoozed-until-before`
- `--added-to-cycle-after`, `--added-to-cycle-before`

#### Entity Filters (12)
- `--cycle` - Cycle/sprint UUID
- `--project` - Project UUID
- `--parent` - Parent issue ID (for sub-issues)
- `--project-milestone` - Milestone UUID
- `--delegate` - Delegated agent (email, name, or 'me')
- `--snoozed-by` - Who snoozed (email, name, or 'me')
- `--last-applied-template` - Template UUID
- `--added-to-cycle-period` - before, during, after
- `--id` - Issue UUID
- `--sla-status` - Breached, HighRisk, MediumRisk, LowRisk, Completed, Failed

#### Text Filters (3)
- `--title` - Text in title (contains)
- `--description` - Text in description (contains)
- `--comment-contains` - Text in comments (contains)

#### Numeric Filters (4)
- `--estimate` - Story points
- `--number` - Issue number
- `--customer-count` - Number of linked customers
- `--customer-important-count` - Important customer count

#### Relation Filters (4)
- `--has-blocking` - Blocks other issues
- `--has-blocked-by` - Blocked by other issues
- `--has-duplicate` - Has duplicate issues
- `--has-related` - Has related issues

#### Collection Filters (6)
Repeatable, OR logic:
- `--attachment-by=me,alice` - Has attachments by any of these users
- `--comment-by=me,bob` - Has comments by any of these users
- `--subscriber=me,team` - Subscribed by any of these users

Boolean checks:
- `--has-children` - Has sub-issues
- `--has-needs` - Has customer needs
- `--has-reactions` - Has emoji reactions

#### AI Suggestion Filters (4)
- `--has-suggested-teams` - Linear AI suggested a team
- `--has-suggested-assignees` - Linear AI suggested an assignee
- `--has-suggested-projects` - Linear AI suggested a project
- `--has-suggested-labels` - Linear AI suggested labels

#### Source Type Filter (1)
- `--attachment-source-type` - uploaded, url, github, slack

### Combining Filters

**All filters are ANDed together:**
```bash
go-linear issue list \
  --team=Engineering \
  --state=Triage \
  --creator=me \
  --has-suggested-teams \
  --created-after=7d \
  --comment-by=manager@company.com
```

Returns: Issues in Engineering team, in Triage state, created by me, with AI team suggestions, from last 7 days, with comments by manager.

**Repeatable filters use OR logic:**
```bash
--label=bug,urgent          # Has bug OR urgent label
--comment-by=me,alice,bob   # Has comments by me OR alice OR bob
```

### Batch Operations

Use `go-linear issue batch-update` with filters to update multiple issues:

```bash
# Batch approve triage (dry run first)
go-linear issue batch-update \
  --state=Triage \
  --has-suggested-teams \
  --creator=colleague@company.com \
  --created-after=7d \
  --set-state=Backlog \
  --dry-run

# Actually apply (after reviewing dry run)
go-linear issue batch-update \
  --state=Triage \
  --has-suggested-teams \
  --set-state=Backlog \
  --yes

# Batch assign
go-linear issue batch-update \
  --team=Engineering \
  --state=Todo \
  --set-assignee=me \
  --batch-limit=20
```

Max 50 issues per batch (API limit).

### Examples

#### Find Triage Issues Needing Approval
```bash
go-linear issue list \
  --state=Triage \
  --has-suggested-teams \
  --creator=colleague@company.com \
  --created-after=7d \
  --count
```

#### Find Blocked Issues
```bash
go-linear issue list \
  --has-blocked-by \
  --assignee=me \
  --output=json
```

#### Find Overdue Issues
```bash
go-linear issue list \
  --due-before=today \
  --state=In Progress \
  --team=Engineering
```

#### Find Issues with Customer Impact
```bash
go-linear issue list \
  --customer-count-gt=0 \
  --has-needs \
  --priority=1
```

#### Find Stale Issues
```bash
go-linear issue list \
  --updated-before=30d \
  --state=In Progress \
  --has-children=false
```

### Filters Not Supported

**Internal fields (9):** accumulatedStateUpdatedAt, ageTime, cycleTime, leadTime, triageTime, hasSuggestedRelatedIssues, hasSuggestedSimilarIssues, searchableContent, suggestions

**Compound logic (2):** and, or (all flags are implicitly ANDed)

**Alpha/Experimental (1):** recurringIssueTemplate

**Complex nested (5):** Full nested collection filtering beyond user/text matching

Total: 17 filters not supported (73% coverage)

### Performance

**Aggregation:**
- Use `--count` to get totals without fetching full results
- 99% token reduction (348 chars → 17 chars for counts)

**Field selection:**
- Use `--fields=defaults` for minimal output (80% reduction)
- Use `--fields=id` for ID-only queries (95% reduction)

See: docs/FIELDS.md for field selection details.

---

## Cycle Filters

`go-linear cycle list` supports 15 filter flags for querying sprints/cycles.

### Filter Categories

#### Date Filters (10 flags)
- `--created-after`, `--created-before` - Creation date
- `--updated-after`, `--updated-before` - Last update date
- `--completed-after`, `--completed-before` - Completion date
- `--starts-after`, `--starts-before` - Sprint start date
- `--ends-after`, `--ends-before` - Sprint end date

#### State Filters (6 boolean flags)
- `--active` - Currently active cycles
- `--future` - Future cycles (not yet started)
- `--past` - Completed cycles
- `--next` - Next upcoming cycle
- `--previous` - Previous cycle
- `--in-cooldown` - Cycles in cooldown period

#### Entity Filters (4)
- `--id` - Cycle UUID
- `--name` - Name contains (case-insensitive)
- `--number` - Cycle number
- `--team` - Team name or key

### Example

```bash
go-linear cycle list --active --team=ENG --output=json
```

Returns active cycles for the Engineering team.

### Combining Filters

All filters are ANDed:
```bash
go-linear cycle list --past --created-after=30d --team=Platform
```

Returns past cycles from Platform team created in the last 30 days.

---

## Document Filters

`go-linear document list` supports 9 filter flags for querying knowledge base documents.

### Filter Categories

#### Date Filters (4 flags)
- `--created-after`, `--created-before` - Creation date
- `--updated-after`, `--updated-before` - Last update date

#### Entity Filters (5)
- `--id` - Document UUID
- `--creator` - Creator name, email, or 'me'
- `--initiative` - Initiative name or UUID
- `--project` - Project name or UUID
- `--issue` - Issue identifier or UUID

#### Text Filters (2)
- `--title` - Title contains (case-insensitive)
- `--slug-id` - Document slug ID (exact match)

### Example

```bash
go-linear document list --created-after=30d --output=json
```

Returns documents created in the last 30 days.

---

## Initiative Filters

`go-linear initiative list` supports 11 filter flags for querying strategic initiatives.

### Filter Categories

#### Date Filters (4 flags)
- `--created-after`, `--created-before` - Creation date
- `--target-after`, `--target-before` - Target completion date

#### Entity Filters (3)
- `--id` - Initiative UUID
- `--creator` - Creator name, email, or 'me'
- `--owner` - Owner name, email, or 'me'

#### Text Filters (2)
- `--name` - Name contains (case-insensitive)
- `--slug-id` - Initiative slug ID (exact match)

#### State Filters (2)
- `--health` - Health status: onTrack, atRisk, offTrack
- `--status` - Initiative status: Planned, Active, Completed

### Example

```bash
go-linear initiative list --status=Active --health=onTrack --output=json
```

Returns active initiatives that are on track.

---

## Comment Filters

`go-linear comment list` supports 8 filter flags for querying issue comments.

### Filter Categories

#### Date Filters (4 flags)
- `--created-after`, `--created-before` - Creation date
- `--updated-after`, `--updated-before` - Last update date

#### Entity Filters (3)
- `--id` - Comment UUID
- `--creator` - Creator name, email, or 'me'
- `--issue` - Issue identifier or UUID

#### Text Filters (1)
- `--body` - Body contains (case-insensitive)

### Example

```bash
go-linear comment list --created-after=7d --issue=ENG-123 --output=json
```

Returns comments on issue ENG-123 created in the last 7 days.
