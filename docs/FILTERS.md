# Issue Filtering Guide

## Overview

`go-linear issue list` supports 64 filter flags across 44 filter types, enabling precise issue queries for production environments.

## Filter Categories

### Core Filters (6)
- `--team` - Team name, key, or UUID (e.g., "Engineering", "ENG")
- `--assignee` - User email, name, or 'me'
- `--state` - State name or UUID (e.g., "In Progress")
- `--priority` - 0=none, 1=urgent, 2=high, 3=normal, 4=low
- `--label` - Label names (repeatable, OR logic)
- `--creator` - Who created (email, name, or 'me')

### Date Filters (17 ranges, 34 flags)
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

### Entity Filters (12)
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

### Text Filters (3)
- `--title` - Text in title (contains)
- `--description` - Text in description (contains)
- `--comment-contains` - Text in comments (contains)

### Numeric Filters (4)
- `--estimate` - Story points
- `--number` - Issue number
- `--customer-count` - Number of linked customers
- `--customer-important-count` - Important customer count

### Relation Filters (4)
- `--has-blocking` - Blocks other issues
- `--has-blocked-by` - Blocked by other issues
- `--has-duplicate` - Has duplicate issues
- `--has-related` - Has related issues

### Collection Filters (6)
Repeatable, OR logic:
- `--attachment-by=me,alice` - Has attachments by any of these users
- `--comment-by=me,bob` - Has comments by any of these users
- `--subscriber=me,team` - Subscribed by any of these users

Boolean checks:
- `--has-children` - Has sub-issues
- `--has-needs` - Has customer needs
- `--has-reactions` - Has emoji reactions

### AI Suggestion Filters (4)
- `--has-suggested-teams` - Linear AI suggested a team
- `--has-suggested-assignees` - Linear AI suggested an assignee
- `--has-suggested-projects` - Linear AI suggested a project
- `--has-suggested-labels` - Linear AI suggested labels

### Source Type Filter (1)
- `--attachment-source-type` - uploaded, url, github, slack

## Combining Filters

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

## Examples

### Find Triage Issues Needing Approval
```bash
go-linear issue list \
  --state=Triage \
  --has-suggested-teams \
  --creator=colleague@company.com \
  --created-after=7d \
  --count
```

### Find Blocked Issues
```bash
go-linear issue list \
  --has-blocked-by \
  --assignee=me \
  --output=json
```

### Find Overdue Issues
```bash
go-linear issue list \
  --due-before=today \
  --state=In Progress \
  --team=Engineering
```

### Find Issues with Customer Impact
```bash
go-linear issue list \
  --customer-count-gt=0 \
  --has-needs \
  --priority=1
```

### Find Stale Issues
```bash
go-linear issue list \
  --updated-before=30d \
  --state=In Progress \
  --has-children=false
```

## Filters Not Supported

**Internal fields (9):** accumulatedStateUpdatedAt, ageTime, cycleTime, leadTime, triageTime, hasSuggestedRelatedIssues, hasSuggestedSimilarIssues, searchableContent, suggestions

**Compound logic (2):** and, or (all flags are implicitly ANDed)

**Alpha/Experimental (1):** recurringIssueTemplate

**Complex nested (5):** Full nested collection filtering beyond user/text matching

Total: 17 filters not supported (73% coverage)

## Performance

**Aggregation:**
- Use `--count` to get totals without fetching full results
- 99% token reduction (348 chars → 17 chars for counts)

**Field selection:**
- Use `--fields=defaults` for minimal output (80% reduction)
- Use `--fields=id` for ID-only queries (95% reduction)

See: docs/FIELDS.md for field selection details.
