# Adding go-linear to Claude Code

Configure the go-linear MCP server using Claude Code's CLI.

## Prerequisites

1. **Claude Code CLI** - Install from [code.claude.com](https://code.claude.com)
2. **Linear API key** - Get from [linear.app/settings/api](https://linear.app/settings/api)
3. **go-linear binary** - Build with `make build-cli`

## Quick Setup

```bash
# 1. Get your binary path
cd /path/to/go-linear
make build-cli
pwd
# Example output: /Users/yourname/code/go-linear
# Binary is at: /Users/yourname/code/go-linear/bin/go-linear

# 2. Add to Claude Code (choose scope below)
```

### Choose Installation Scope

**Option A: User-wide (Recommended)** - Available in all projects
```bash
claude mcp add --scope user --transport stdio go-linear \
  --env LINEAR_API_KEY=lin_api_YOUR_KEY_HERE \
  -- /absolute/path/to/go-linear/bin/go-linear mcp start
```

**Option B: Local** - Only available in current folder
```bash
claude mcp add --scope local --transport stdio go-linear \
  --env LINEAR_API_KEY=lin_api_YOUR_KEY_HERE \
  -- /absolute/path/to/go-linear/bin/go-linear mcp start
```

**Option C: Project** - Shared via `.mcp.json` (team collaboration)
```bash
claude mcp add --scope project --transport stdio go-linear \
  --env LINEAR_API_KEY=lin_api_YOUR_KEY_HERE \
  -- /absolute/path/to/go-linear/bin/go-linear mcp start
```
*Note: Don't commit `.mcp.json` with API keys*

**Important:**
- Replace `lin_api_YOUR_KEY_HERE` with your actual Linear API key
- Replace `/absolute/path/to/...` with your actual binary path from step 1
- The `--` separator distinguishes Claude's flags from the server command

## Get Your Linear API Key

1. Go to [linear.app/settings/api](https://linear.app/settings/api)
2. Click "Create key"
3. Copy the key (starts with `lin_api_`)

## Verify Installation

```bash
# List configured servers
claude mcp list

# Check server details
claude mcp get go-linear

# Test in Claude Code
# Use /mcp command to see server status
```

## Configuration Scopes

Choose where the server is available:

```bash
# Set API key first (do this once)
export LINEAR_API_KEY=lin_api_xxx

# User-wide (all projects)
claude mcp add --scope user --transport stdio go-linear \
  -- /path/to/go-linear mcp start

# Project-only (shared via .mcp.json)
claude mcp add --scope project --transport stdio go-linear \
  -- /path/to/go-linear mcp start

# Local (default, private to current project)
claude mcp add --transport stdio go-linear \
  -- /path/to/go-linear mcp start
```

**Note:** When using `--scope project`, don't commit `.mcp.json` with secrets. Keep API keys in environment variables.

## Persistent Configuration

Add to your shell profile for automatic setup:

**~/.bashrc or ~/.zshrc:**
```bash
# Linear MCP configuration
export LINEAR_API_KEY=lin_api_xxx
```

Then reload:
```bash
source ~/.bashrc  # or source ~/.zshrc
```

## Optional: Enable Logging

Add logging environment variable:

```bash
claude mcp add --transport stdio go-linear \
  --env LINEAR_API_KEY=lin_api_xxx \
  --env LINEAR_LOG_LEVEL=info \
  -- /path/to/go-linear/bin/go-linear mcp start
```

## Troubleshooting

### "Server failed to start"

**Test the binary directly:**
```bash
/absolute/path/to/go-linear/bin/go-linear mcp start
```

If it errors:
- Check path is correct and absolute
- Verify binary has execute permissions: `chmod +x /path/to/go-linear`

### "API key invalid"

**Verify your key works:**
```bash
export LINEAR_API_KEY=lin_api_YOUR_KEY
./bin/go-linear viewer
```

Should show your Linear user info. If not, regenerate your API key.

### View server details

```bash
# See configured command and environment
claude mcp get go-linear

# List all servers
claude mcp list
```

### Test with MCP Inspector

Test locally before adding to Claude:

```bash
npx @modelcontextprotocol/inspector /path/to/go-linear -- mcp start
```

Opens web UI at http://localhost:5173.

## What You Can Do

Once configured, Claude can:

- **List and search issues** - "Show my urgent issues", "Find bugs in Engineering"
- **Create and update issues** - "Create a bug for slow login", "Mark ENG-123 as done"
- **Batch operations** - "Update all triage issues to backlog"
- **Team info** - "Who's on the Engineering team?"
- **Completed work** - "What did Alice complete yesterday?"
- **Count and filter** - "How many P1 bugs are there?"

See [FILTERS.md](FILTERS.md) for all 44 filter types.

## Removing go-linear

```bash
claude mcp remove go-linear
```

## Next Steps

- Read [CLI Quick Start](CLI-QUICK-START.md) to use go-linear from terminal
- See [FILTERS.md](FILTERS.md) for advanced filtering
- Check [FIELDS.md](FIELDS.md) for field selection options
