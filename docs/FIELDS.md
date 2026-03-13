# Field Selection Guide

## Overview

All `list` and `get` commands support field selection via the `--fields` flag to control JSON output verbosity.

## Syntax

```bash
--fields=<spec>
```

Where `<spec>` can be:

| Syntax | Behavior | Example |
|--------|----------|---------|
| *(empty)* | Use smart defaults | `go-linear issue list` |
| `defaults` | Explicitly use defaults | `--fields=defaults` |
| `none` | Show all available fields | `--fields=none` |
| `defaults,field,...` | Add fields to defaults | `--fields=defaults,estimate,number` |
| `field1,field2,...` | Show only specified fields | `--fields=id,title,url` |

## Default Field Sets

**These defaults are part of the stable API.** Fields may be added in minor versions (non-breaking) but will not be removed within a major version.

### Issue Commands

**`issue list`** (8 fields):
```
id, identifier, title, url,
state.name, team.key, priority, createdAt
```

**`issue get`** (10 fields):
```
+ description, assignee.name
```

### Team Commands

**`team list`** (6 fields):
```
id, name, key, description, icon, createdAt
```

**`team get`** (8 fields):
```
+ color, private
```

### User Commands

**`user list`** (6 fields):
```
id, name, displayName, email, active, avatarUrl
```

**`user get`** (7 fields):
```
+ admin
```

### Other Resources

**Comment**: 5-6 fields (id, body, createdAt, user.name, url, [editedAt])
**Label**: 4-5 fields (id, name, color, createdAt, [description])
**WorkflowState**: 5 fields (id, name, type, color, position)
**Cycle**: 5-6 fields (id, name, startsAt, endsAt, createdAt, [description])
**Attachment**: 5 fields (id, title, url, source, createdAt)
**Project**: 4-6 fields
**Roadmap**: 4 fields
**Initiative**: 4 fields

## Nested Field Access

Use dot notation to access nested object fields:

```bash
# Access nested user fields
--fields=id,assignee.name,assignee.email

# Access nested state fields
--fields=id,state.name,state.type,state.color

# Access nested team fields
--fields=id,team.key,team.name,team.icon
```

## Examples

### Use Defaults
```bash
linear issue get ENG-123# Returns 10 fields

linear issue list# Returns 8 fields per issue
```

### Add to Defaults
```bash
# Need description and estimate too?
linear issue get ENG-123 --fields=defaults,estimate,number# Returns: 10 default fields + estimate + number = 12 fields

# Multiple additions
linear issue get ENG-123 --fields=defaults,estimate,labels,dueDate```

### Show Only Specific Fields
```bash
# Just ID and title
linear issue get ENG-123 --fields=id,title# Returns: {"id": "...", "title": "..."}

# With nested field
linear issue get ENG-123 --fields=id,title,assignee.name# Returns: {"id": "...", "title": "...", "assignee": {"name": "..."}}
```

### Show All Fields
```bash
# See everything available
linear issue get ENG-123 --fields=none# Returns: ~50 fields (varies by resource type)
```

## User Configuration

Create `~/.config/linear/config.yaml` for personal defaults:

```yaml
# Command defaults (applies to issue create)
defaults:
  team: PSEC           # Default team
  labels:              # Default labels
    - triage

# Field display defaults
field_defaults:
  issue.list: "id,identifier,title,state.name,priority"
  issue.get: "id,title,description,state.name,assignee.name,estimate"
  team.list: "id,name,key"

# MCP-specific defaults (for AI agents)
mcp:
  field_defaults:
    issue.list: "id,identifier,title,url,state.name"
```

**Setup:**
```bash
mkdir -p ~/.config/linear
# Edit ~/.config/linear/config.yaml with your defaults
```

**Resolution order:**
1. Explicit flags (highest priority)
2. Workspace config (`.linear-workspace.yaml`)
3. User config (`~/.config/linear/config.yaml`)
4. Built-in defaults (fallback)

## Available Fields

### Issue (~50 fields)

**Essential:**
- id, identifier, number, title, description, url
- priority, priorityLabel, estimate
- createdAt, updatedAt, completedAt, canceledAt

**Nested objects:**
- state.{id, name, type, color}
- team.{id, name, key, icon, color}
- assignee.{id, name, displayName, email}
- creator.{id, name, email}
- cycle.{id, name, startsAt, endsAt}
- project.{id, name, color}

**Expensive** (avoid in defaults):
- comments, attachments, history (require pagination)

### Team (~80 fields)

**Essential:**
- id, name, key, displayName, description
- icon, color, timezone, private
- createdAt, updatedAt, archivedAt

**Expensive:**
- members, issues, cycles, labels (require pagination)

### User (~40 fields)

**Essential:**
- id, name, displayName, email
- active, admin, guest, avatarUrl
- createdAt, updatedAt

**Expensive:**
- assignedIssues, createdIssues, teams (require pagination)

### Full Schema

See Linear's GraphQL schema: https://studio.apollographql.com/public/Linear-API/variant/current/schema/reference

Or introspect locally:
```bash
cat schema.graphql | grep "type Issue " -A 200
```

## Best Practices

### For AI Agents

**Use defaults first:**
```bash
linear issue list# Returns manageable dataset (~50 lines per issue)
```

**Add fields as needed:**
```bash
linear issue list --fields=defaults,description,labels# Only request additional fields when necessary
```

**Avoid expensive fields:**
- Don't include `comments`, `attachments`, `history` unless specifically needed
- These require separate API calls and pagination

### For Humans

**Pipe to jq for readable output:**
```bash
linear issue list | jq '.nodes[] | "\(.identifier): \(.title)"'
```

**Extract specific fields:**
```bash
linear issue list | jq '.nodes[] | .identifier'
```

**Use none for debugging:**
```bash
linear issue get ENG-123 --fields=none | jq .
# See everything available
```

