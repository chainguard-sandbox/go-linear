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
go-linear issue list --state="In Progress" --output=json
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

## All Available Commands

**Issues (8)**:
- `go-linear issue list` - List with 64 filters, --count mode
- `go-linear issue get <id>` - Get single issue
- `go-linear issue search <query>` - Full-text search, --count mode
- `go-linear issue create` - Create issue
- `go-linear issue update <id>` - Update issue
- `go-linear issue batch-update` - Update multiple (max 50) ⭐
- `go-linear issue delete <id>` - Delete issue ⚠️
- `go-linear issue relate <id1> <id2>` - Create relationship

**Users (3)**:
- `go-linear user list` - List all users
- `go-linear user get <name|email>` - Get user details
- `go-linear user completed` - Get completed work ⭐

**Teams (2)**:
- `go-linear team list` - List teams
- `go-linear team members --team=X` - List members

**Projects (4)**:
- `go-linear project list` - List projects
- `go-linear project get <id>` - Get project
- `go-linear project create` - Create project
- `go-linear project update <id>` - Update project

**Cycles (2)**:
- `go-linear cycle list` - List cycles/sprints
- `go-linear cycle get <id>` - Get cycle

**Comments (4)**:
- `go-linear comment list` - List comments
- `go-linear comment create` - Add comment
- `go-linear comment update <id>` - Update comment
- `go-linear comment delete <id>` - Delete comment ⚠️

**Labels (2)**:
- `go-linear label list` - List labels
- `go-linear label create` - Create label

**States & Info (2)**:
- `go-linear state list` - List workflow states
- `go-linear viewer` - Get current user

## Output Formats

All commands support multiple output formats:

```bash
--output=json   # Machine-readable (default for AI agents)
--output=table  # Human-readable formatted tables (default)
```

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
  --assignee=alice@company.com \
  --output=json
```

## MCP Mode (For AI Agents)

The same binary works as an MCP server. [Ophis](https://github.com/njayp/ophis) automatically converts all CLI commands into 72 MCP tools.

**Setup for Claude Code:**

```bash
claude mcp add --transport stdio go-linear \
  --env LINEAR_API_KEY=lin_api_xxx \
  -- /path/to/go-linear/bin/go-linear mcp start
```

See [CLAUDE-SETUP.md](CLAUDE-SETUP.md) for full instructions.

## Tips for AI Agents

1. **Always use `--output=json`** for programmatic parsing
2. **Use relative dates** when possible (`yesterday` vs `2025-12-09`)
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
