# CLI Quick Start (For Humans)

Command-line tool for managing Linear issues, projects, and teams.

**Note:** AI agents should use MCP mode (`go-linear mcp start`) - see main README.

## Installation

```bash
# Build
make build-cli

# Creates one binary:
bin/go-linear
```

## Authentication

Set your Linear API key:
```bash
export LINEAR_API_KEY=lin_api_xxx
```

Or pass it as a flag:
```bash
go-linear --api-key=lin_api_xxx issue list
```

## Quick Examples

### List Issues
```bash
# My urgent issues
go-linear issue list --assignee=me --priority=1

# Engineering team's completed work from yesterday
go-linear issue list --team=Engineering --completed-after=yesterday --completed-before=today

# Issues in specific state
go-linear issue list --state="In Progress"
```

### Complex Query (AI-Optimized)
```bash
# Who from Engineering completed work yesterday?
go-linear user completed --team=Engineering --completed-after=yesterday --completed-before=today

# My completed work from last 7 days
go-linear user completed --user=me --completed-after=7d
```

### Create & Update
```bash
# Create issue
go-linear issue create --team=Engineering --title="Fix bug" --description="Details" --priority=1

# Update issue
go-linear issue update ENG-123 --state=Done --add-label=fixed

# Create project
go-linear project create --name="Q1 Platform" --description="Platform improvements"
```

### Teams & Users
```bash
# List teams
go-linear team list

# List team members
go-linear team members --team=Engineering

# Get user info
go-linear user get alice@company.com
go-linear user get me
```

### Relationships
```bash
# Create issue relation
go-linear issue relate ENG-123 ENG-124 --type=blocks
```

### Status Updates & Metrics
```bash
# Create project status update
go-linear project status-update-create --project="Q1 Platform" --body="Week 1: On track" --health=onTrack

# List project updates
go-linear project status-update-list --project="Q1 Platform"

# Create initiative status update
go-linear initiative status-update-create --initiative="Shrink Wolfi" --body="Q1 progress update" --health=atRisk

# Calculate team velocity
go-linear team velocity --team=ENG --cycles=3

# View enhanced project metrics
go-linear project get <uuid>  # Shows: progress %, health, dates, lead, teams, linked initiatives

# View enhanced initiative metrics
go-linear initiative get <uuid>  # Shows: status, health, owner, parent, linked projects
```

### Initiative-Project Linking
```bash
# Link project to initiative
go-linear initiative add-project --initiative="Shrink Wolfi" --project="Cloud cost optimization"

# Unlink project from initiative
go-linear initiative remove-project --initiative="Shrink Wolfi" --project="Cloud cost optimization"
```

### Documents
```bash
# Create team document
go-linear document create --title="API Guide" --team=ENG --content="# Documentation..."

# Create project document
go-linear document create --title="Architecture" --project="Q1 Platform" --content="## Design..."

# Update document
go-linear document update <uuid> --title="Updated Guide" --content="New content"

# Delete document
go-linear document delete <uuid>
```

## All Available Commands

**Issues (15)**:
- `go-linear issue list` - List with 64 filters, --count mode
- `go-linear issue get <id>` - Get single issue
- `go-linear issue search <query>` - Full-text search, --count mode
- `go-linear issue suggestions <id>` - View AI suggestions
- `go-linear issue create` - Create issue. Supports --template
- `go-linear issue update <id>` - Update issue
- `go-linear issue batch-update` - Update multiple (max 50) ⭐
- `go-linear issue delete <id>` - Delete issue (--permanent for no grace period) ⚠️
- `go-linear issue archive <id>` - Archive issue (--trash for 30-day auto-delete) ⭐
- `go-linear issue unarchive <id>` - Restore archived/trashed issue ⭐
- `go-linear issue relate <id1> <id2>` - Create relationship
- `go-linear issue unrelate <id1> <id2>` - Remove relationship
- `go-linear issue update-relation <id>` - Change relationship type
- `go-linear issue add-label <id>` - Add label to issue
- `go-linear issue remove-label <id>` - Remove label from issue

**Users (3)**:
- `go-linear user list` - List all users
- `go-linear user get <name|email>` - Get user details
- `go-linear user completed` - Get completed work ⭐

**Teams (10)**:
- `go-linear team list` - List teams
- `go-linear team get <id>` - Get team details
- `go-linear team create` - Create team
- `go-linear team update <id>` - Update team
- `go-linear team delete <id>` - Delete team ⚠️
- `go-linear team members --team=X` - List members
- `go-linear team add-member` - Add user to team
- `go-linear team remove-member` - Remove user from team ⭐
- `go-linear team unarchive <id>` - Restore archived team ⭐
- `go-linear team velocity --team=X` - Calculate performance metrics ⭐

**Projects (15)**:
- `go-linear project list` - List projects
- `go-linear project get <id>` - Get project with metrics
- `go-linear project create` - Create project
- `go-linear project update <id>` - Update project
- `go-linear project delete <id>` - Delete project ⚠️
- `go-linear project archive <id>` - Archive project (Linear recommends delete) ⭐
- `go-linear project unarchive <id>` - Restore archived project ⭐
- `go-linear project milestone-create` - Create milestone
- `go-linear project milestone-update <id>` - Update milestone
- `go-linear project milestone-delete <id>` - Delete milestone ⚠️
- `go-linear project milestone-list` - List project milestones ⭐
- `go-linear project status-update-create` - Create status update ⭐
- `go-linear project status-update-list` - List updates ⭐
- `go-linear project status-update-get <id>` - Get update ⭐
- `go-linear project status-update-delete <id>` - Delete update ⭐

**Initiatives (14)**:
- `go-linear initiative list` - List initiatives
- `go-linear initiative get <id>` - Get initiative with metrics
- `go-linear initiative create` - Create initiative
- `go-linear initiative update <id>` - Update initiative
- `go-linear initiative delete <id>` - Delete initiative ⚠️
- `go-linear initiative archive <id>` - Archive initiative ⭐
- `go-linear initiative unarchive <id>` - Restore archived initiative ⭐
- `go-linear initiative list-sub <id>` - List sub-initiatives ⭐
- `go-linear initiative status-update-create` - Create status update ⭐
- `go-linear initiative status-update-list` - List updates ⭐
- `go-linear initiative status-update-get <id>` - Get update ⭐
- `go-linear initiative status-update-archive <id>` - Archive update ⭐
- `go-linear initiative add-project` - Link project to initiative ⭐
- `go-linear initiative remove-project` - Unlink project ⭐

**Cycles (5)**:
- `go-linear cycle list` - List cycles/sprints
- `go-linear cycle get <id>` - Get cycle with metrics
- `go-linear cycle create` - Create cycle
- `go-linear cycle update <id>` - Update cycle
- `go-linear cycle archive <id>` - Archive cycle

**Comments (5)**:
- `go-linear comment list` - List comments
- `go-linear comment get <id>` - Get comment with thread
- `go-linear comment create` - Add comment. Supports --parent for threading
- `go-linear comment update <id>` - Update comment
- `go-linear comment delete <id>` - Delete comment ⚠️

**Labels (5)**:
- `go-linear label list` - List labels
- `go-linear label get <id>` - Get label
- `go-linear label create` - Create label
- `go-linear label update <id>` - Update label
- `go-linear label delete <id>` - Delete label ⚠️

**Documents (7)**:
- `go-linear document list` - List documents
- `go-linear document get <id>` - Get document
- `go-linear document create` - Create document ⭐
- `go-linear document update <id>` - Update document ⭐
- `go-linear document delete <id>` - Delete document ⚠️
- `go-linear document unarchive <id>` - Restore deleted document ⭐

**Notifications (7)**:
- `go-linear notification list` - List inbox notifications
- `go-linear notification get <id>` - Get notification details
- `go-linear notification update <id>` - Mark read or snooze
- `go-linear notification archive <id>` - Archive notification
- `go-linear notification unarchive <id>` - Restore to inbox
- `go-linear notification subscribe` - Subscribe to entity updates
- `go-linear notification unsubscribe <id>` - Remove subscription

**Attachments (7)**:
- `go-linear attachment list` - List attachments
- `go-linear attachment get <id>` - Get attachment
- `go-linear attachment create` - Create attachment
- `go-linear attachment delete <id>` - Delete attachment ⚠️
- `go-linear attachment link-github` - Link GitHub PR/issue
- `go-linear attachment link-slack` - Link Slack message
- `go-linear attachment link-url` - Link URL

**States & Info (4)**:
- `go-linear state list` - List workflow states
- `go-linear state get <id>` - Get workflow state
- `go-linear viewer` - Get current user
- `go-linear status` - Check rate limits

## Output

All commands output JSON. Use `--fields` to control which fields appear. Pipe to `jq` for human-readable formatting.

## AI Agent Features

### 1. Name Resolution
```bash
--team=Engineering          # Not: --team=550e8400-...
--assignee=alice@company.com # Not: --assignee=uuid
--state="In Progress"       # Not: --state=uuid
```

### 2. Relative Dates
```bash
--completed-after=yesterday  # Not: --completed-after=2025-12-09T00:00:00Z
--created-after=7d          # 7 days ago
--updated-before=2w         # 2 weeks ago
```

### 3. Rich Filtering
```bash
go-linear issue list \
  --team=Engineering \
  --priority=1 \
  --state="In Progress" \
  --created-after=7d \
  --label=bug \
  --assignee=alice@company.com
```

## MCP Mode (For AI Agents)

The same binary works as an MCP server. [Ophis](https://github.com/njayp/ophis) automatically converts all CLI commands into MCP tools.

**Setup for Claude Code:**

```bash
claude mcp add --transport stdio go-linear \
  --env LINEAR_API_KEY=lin_api_xxx \
  -- /path/to/go-linear/bin/go-linear mcp start
```

See [CLAUDE-SETUP.md](CLAUDE-SETUP.md) for full instructions.

## Tips for AI Agents

1. **Use relative dates** when possible (`yesterday` vs `2025-12-09`)
3. **Prefer named parameters** (`--team=Engineering` vs team IDs)
4. **Check `--help`** for complete flag lists
5. **Complex queries**: Use `user completed` instead of multiple list commands

## Getting Help

```bash
go-linear --help                    # Show all commands
go-linear issue --help              # Show issue subcommands
go-linear issue list --help         # Show all flags for issue list
go-linear mcp --help                # Show MCP integration options
```
