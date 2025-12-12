# go-linear-mcp

An MCP (Model Context Protocol) server that exposes Linear API operations as tools for AI agents.

## Quick Start

### Build

```bash
go build -o go-linear-mcp
```

### Run

```bash
export LINEAR_API_KEY=lin_api_xxx
./go-linear-mcp
```

The server communicates over stdio using JSON-RPC 2.0.

## Configuration

The server is configured via environment variables with production-appropriate defaults tuned for Linear's API characteristics.

### Required

- **`LINEAR_API_KEY`** - Linear API key (get from https://linear.app/settings/account/security)

### Optional Configuration

All settings have sane defaults optimized for Linear's API. Only override if you have specific requirements.

| Variable | Default | Description |
|----------|---------|-------------|
| `LINEAR_BASE_URL` | `https://api.linear.app/graphql` | Custom API endpoint (for self-hosted instances) |
| `LINEAR_TIMEOUT` | `30s` | Request timeout (e.g., "30s", "1m") |
| `LINEAR_RETRY_ATTEMPTS` | `3` | Number of retry attempts (0 to disable) |
| `LINEAR_RETRY_INITIAL` | `1s` | Initial retry backoff delay |
| `LINEAR_RETRY_MAX` | `30s` | Maximum retry backoff delay |
| `LINEAR_CIRCUIT_BREAKER_FAILURES` | `5` | Failures before circuit opens |
| `LINEAR_CIRCUIT_BREAKER_TIMEOUT` | `60s` | Circuit breaker reset timeout |
| `LINEAR_TLS_MIN_VERSION` | `1.2` | Minimum TLS version ("1.2" or "1.3") |

**Defaults are tuned for Linear's API:**
- Rate limits: 250,000 points/hour for API keys, 1,500 requests/hour per user
- Max query complexity: 10,000 points per query
- Uses leaky bucket algorithm for rate limiting

See [Linear API Rate Limiting](https://developers.linear.app/docs/graphql/working-with-the-graphql-api/rate-limiting) for details.

## Using with AI Agents

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "linear": {
      "command": "/absolute/path/to/go-linear-mcp",
      "env": {
        "LINEAR_API_KEY": "lin_api_xxx"
      }
    }
  }
}
```

**With custom configuration:**

```json
{
  "mcpServers": {
    "linear": {
      "command": "/absolute/path/to/go-linear-mcp",
      "env": {
        "LINEAR_API_KEY": "lin_api_xxx",
        "LINEAR_TIMEOUT": "45s",
        "LINEAR_RETRY_ATTEMPTS": "5"
      }
    }
  }
}
```

Restart Claude Desktop to activate.

### Other MCP Clients

The server implements the MCP protocol over stdio. Any MCP-compatible client can use it. See [examples/mcp-client](../../examples/mcp-client) for a Go client example.

## Available Tools

**Read-Only (9 tools):**
- `linear_get_viewer` - Get authenticated user
- `linear_list_teams` - List teams
- `linear_list_issues` - List issues with pagination
- `linear_get_issue` - Get single issue by ID
- `linear_search_issues` - Search issues
- `linear_list_workflow_states` - List workflow states
- `linear_list_users` - List users
- `linear_list_labels` - List labels
- `linear_list_comments` - List comments

**Write (3 tools):**
- `linear_create_issue` - ⚠️ Create new issue
- `linear_update_issue` - ⚠️ Update issue
- `linear_create_comment` - ⚠️ Add comment

**Destructive (1 tool):**
- `linear_delete_issue` - ⚠️ Delete issue (PERMANENT)

## Documentation

- **Full Documentation:** [docs/MCP.md](../../docs/MCP.md)
- **Tool Definitions:** [mcp/tools.json](../../mcp/tools.json)
- **JSON Schema:** [pkg/linear/schema.json](../../pkg/linear/schema.json)

## Testing

Test with the MCP inspector:

```bash
npx @modelcontextprotocol/inspector ./go-linear-mcp
```

Or manually via stdio:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | ./go-linear-mcp
```

## Requirements

- Go 1.25+
- Linear API key with appropriate permissions (Read for queries, Write for mutations)

## License

Apache 2.0 - See [LICENSE](../../LICENSE)
