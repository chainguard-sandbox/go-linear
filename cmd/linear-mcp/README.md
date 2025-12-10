# Linear MCP Server

An MCP (Model Context Protocol) server that exposes Linear API operations as tools for AI agents.

## Quick Start

### Build

```bash
go build -o linear-mcp
```

### Run

```bash
export LINEAR_API_KEY=lin_api_xxx
./linear-mcp
```

The server communicates over stdio using JSON-RPC 2.0.

## Using with AI Agents

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "linear": {
      "command": "/absolute/path/to/linear-mcp",
      "env": {
        "LINEAR_API_KEY": "lin_api_xxx"
      }
    }
  }
}
```

Restart Claude Desktop to activate.

### Custom Integration

The server implements the MCP protocol over stdio. See [examples/mcp-client](../../examples/mcp-client) for a Go client example.

## Available Tools

- `linear_get_viewer` - Get authenticated user
- `linear_list_teams` - List teams
- `linear_list_issues` - List issues with pagination
- `linear_get_issue` - Get single issue by ID
- `linear_search_issues` - Search issues
- `linear_create_issue` - Create new issue
- `linear_update_issue` - Update issue
- `linear_delete_issue` - Delete issue
- `linear_list_workflow_states` - List workflow states
- `linear_list_users` - List users
- `linear_list_labels` - List labels
- `linear_create_comment` - Add comment
- `linear_list_comments` - List comments

## Documentation

- **Full Documentation:** [docs/MCP.md](../../docs/MCP.md)
- **Tool Definitions:** [mcp/tools.json](../../mcp/tools.json)
- **JSON Schema:** [pkg/linear/schema.json](../../pkg/linear/schema.json)

## Testing

Test with the MCP inspector:

```bash
npx @modelcontextprotocol/inspector ./linear-mcp
```

Or manually via stdio:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | ./linear-mcp
```

## Requirements

- Go 1.24+
- Linear API key with appropriate permissions (Read for queries, Write for mutations)

## License

Apache 2.0 - See [LICENSE](../../LICENSE)
