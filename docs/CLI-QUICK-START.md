# Linear CLI Quick Start

The Linear CLI provides 26 commands for interacting with Linear's API, optimized for both humans and AI agents.

## Installation

```bash
# Build the CLI
make build-cli

# The binary is created at:
bin/linear
```

## Authentication

Set your Linear API key:
```bash
export LINEAR_API_KEY=lin_api_xxx
```

Or pass it as a flag:
```bash
linear --api-key=lin_api_xxx issue list
```

## Quick Examples

### List Issues
```bash
# My urgent issues
linear issue list --assignee=me --priority=1

# Engineering team's completed work from yesterday
linear issue list --team=Engineering --completed-after=yesterday --completed-before=today

# Issues in specific state
linear issue list --state="In Progress" --output=json
```

### Complex Query (AI-Optimized)
```bash
# Who from Engineering completed work yesterday?
linear user completed --team=Engineering --completed-after=yesterday --completed-before=today

# My completed work from last 7 days
linear user completed --user=me --completed-after=7d
```

### Create & Update
```bash
# Create issue
linear issue create --team=Engineering --title="Fix bug" --description="Details" --priority=1

# Update issue
linear issue update ENG-123 --state=Done --add-label=fixed

# Create project
linear project create --name="Q1 Platform" --description="Platform improvements"
```

### Teams & Users
```bash
# List teams
linear team list

# List team members
linear team members --team=Engineering

# Get user info
linear user get alice@company.com
linear user get me
```

### Relationships
```bash
# Create issue relation
linear issue relate ENG-123 ENG-124 --type=blocks
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
- `linear user list` - List all users
- `linear user get <name|email>` - Get user details
- `linear user completed` - Get completed work ⭐

**Teams (2)**:
- `linear team list` - List teams
- `linear team members --team=X` - List members

**Projects (4)**:
- `linear project list` - List projects
- `linear project get <id>` - Get project
- `linear project create` - Create project
- `linear project update <id>` - Update project

**Cycles (2)**:
- `linear cycle list` - List cycles/sprints
- `linear cycle get <id>` - Get cycle

**Comments (4)**:
- `linear comment list` - List comments
- `linear comment create` - Add comment
- `linear comment update <id>` - Update comment
- `linear comment delete <id>` - Delete comment ⚠️

**Labels (2)**:
- `linear label list` - List labels
- `linear label create` - Create label

**States & Info (2)**:
- `linear state list` - List workflow states
- `linear viewer` - Get current user

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
linear issue list \
  --team=Engineering \
  --priority=1 \
  --state="In Progress" \
  --created-after=7d \
  --label=bug \
  --assignee=alice@company.com \
  --output=json
```

## MCP Server Mode

Start as MCP server for Claude Desktop:

```bash
./bin/linear mcp start
```

Configure in Claude Desktop:
```bash
./bin/linear mcp claude setup
```

All 26 commands automatically become MCP tools via ophis!

## Tips for AI Agents

1. **Always use `--output=json`** for programmatic parsing
2. **Use relative dates** when possible (`yesterday` vs `2025-12-09`)
3. **Prefer named parameters** (`--team=Engineering` vs team IDs)
4. **Check `--help`** for complete flag lists
5. **Complex queries**: Use `user completed` instead of multiple list commands

## Getting Help

```bash
linear --help                    # Show all commands
linear issue --help              # Show issue subcommands
linear issue list --help         # Show all flags for issue list
linear mcp --help                # Show MCP integration options
```
