# MCP (Model Context Protocol) Server for Linear

This directory contains an MCP server implementation that exposes Linear API operations as tools for AI agents.

## What is MCP?

[Model Context Protocol (MCP)](https://modelcontextprotocol.io/) is an open protocol by Anthropic that enables AI applications to securely connect to external data sources and tools. The go-linear MCP server allows AI agents (like Claude) to interact with Linear's issue tracking system.

## ⚠️ Safety & Security

**Current Implementation Status**: The MCP server implements **13 tools** covering all major Linear operations:

**Read-Only (Safe)**:
- ✅ `linear_get_viewer` - Get authenticated user
- ✅ `linear_list_teams` - List teams
- ✅ `linear_list_issues` - List issues
- ✅ `linear_get_issue` - Get single issue
- ✅ `linear_search_issues` - Search issues
- ✅ `linear_list_workflow_states` - List states
- ✅ `linear_list_users` - List users
- ✅ `linear_list_labels` - List labels
- ✅ `linear_list_comments` - List comments

**Write (Non-Destructive)**:
- ✅ `linear_create_issue` - Create new issue
- ✅ `linear_update_issue` - Update issue
- ✅ `linear_create_comment` - Add comment

**Destructive**:
- ⚠️ `linear_delete_issue` - Delete issue permanently

**Dangerous Operations**: The `linear_delete_issue` tool:
- Be marked with `"x-dangerous": true` and `"x-requires-confirmation": true`
- Include ⚠️ warnings in descriptions
- Require explicit user confirmation before execution
- Cannot be undone once executed

**API Key Permissions**: Linear API keys support these permission scopes (configured at https://linear.app/settings/account/security):
- **Read**: Query operations only (Issues, Teams, Users, etc.)
- **Write**: Create and update operations
- **Admin**: Administrative operations
- **Create issues**: Specific permission for issue creation
- **Create comments**: Specific permission for comments

**Permission Detection**: Linear's API doesn't expose which scopes a key has. The API returns:
- `401 Unauthorized` - Invalid/expired API key
- `403 Forbidden` - Valid key, but missing required permission

The `Viewer` query returns an `admin` boolean field indicating if the user is an admin.

**Best Practice**: Create API keys with minimal permissions needed:
- AI agents reading data: **Read only**
- AI agents creating issues/comments: **Read + Create issues + Create comments**
- Full AI assistant: **Read + Write** (avoid Admin unless necessary)

**Sources**:
- [Linear API Documentation](https://linear.app/developers/graphql)
- [Linear Security & Access](https://linear.app/docs/security-and-access)

## Quick Start

### 1. Build the MCP Server

```bash
cd cmd/linear-mcp
go build -o linear-mcp
```

### 2. Configure Your Environment

```bash
export LINEAR_API_KEY=lin_api_xxx
```

Get your API key from: https://linear.app/settings/account/security

### 3. Run the Server

```bash
./linear-mcp
```

The server communicates over stdio using JSON-RPC 2.0 protocol.

## Using with Claude Desktop

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "linear": {
      "command": "/path/to/go-linear/cmd/linear-mcp/linear-mcp",
      "env": {
        "LINEAR_API_KEY": "lin_api_xxx"
      }
    }
  }
}
```

Restart Claude Desktop, and the Linear tools will be available.

## Available Tools

### Authentication & Discovery

#### `linear_get_viewer`
Get authenticated user information. Use this to verify authentication.

**Parameters:** None

**Example:**
```json
{
  "name": "linear_get_viewer",
  "arguments": {}
}
```

#### `linear_list_teams`
List all teams in the workspace. Essential for getting team IDs needed to create issues.

**Parameters:**
- `first` (number, optional): Number of teams to return (default: 50, max: 250)
- `after` (string, optional): Cursor for pagination

**Example:**
```json
{
  "name": "linear_list_teams",
  "arguments": {
    "first": 10
  }
}
```

### Issue Management

#### `linear_list_issues`
List issues with cursor-based pagination.

**Parameters:**
- `first` (number, optional): Number of issues to return (default: 50, max: 250)
- `after` (string, optional): Pagination cursor from previous response

**Example:**
```json
{
  "name": "linear_list_issues",
  "arguments": {
    "first": 20
  }
}
```

#### `linear_get_issue`
Get a single issue by UUID.

**Parameters:**
- `id` (string, required): Issue UUID

**Example:**
```json
{
  "name": "linear_get_issue",
  "arguments": {
    "id": "abc-123-uuid"
  }
}
```

#### `linear_search_issues`
Search for issues using text query with operators.

**Parameters:**
- `query` (string, required): Search query (supports operators like `priority:urgent`, `state:started`)
- `first` (number, optional): Number of results
- `after` (string, optional): Pagination cursor

**Example:**
```json
{
  "name": "linear_search_issues",
  "arguments": {
    "query": "priority:urgent assignee:me",
    "first": 10
  }
}
```

**Search Operators:**
- `priority:urgent` - Filter by priority (urgent, high, normal, low, none)
- `state:started` - Filter by state type (triage, backlog, unstarted, started, completed, canceled)
- `assignee:me` - Issues assigned to you
- `label:bug` - Issues with specific label
- Text search: Any text searches title and description

#### `linear_create_issue`
Create a new issue in Linear.

**Parameters:**
- `teamID` (string, required): Team UUID from `linear_list_teams`
- `title` (string, optional): Issue title
- `description` (string, optional): Description in markdown
- `priority` (number, optional): 0=none, 1=urgent, 2=high, 3=normal, 4=low
- `assigneeID` (string, optional): User UUID to assign
- `stateID` (string, optional): Workflow state UUID
- `labelIds` (array, optional): Array of label UUIDs
- `parentID` (string, optional): Parent issue UUID for sub-issues
- `dueDate` (string, optional): Due date in YYYY-MM-DD format

**Example:**
```json
{
  "name": "linear_create_issue",
  "arguments": {
    "teamID": "87f863b0-b00c-4059-827c-1d163b557bdd",
    "title": "Fix authentication bug",
    "description": "Users cannot log in on Safari",
    "priority": 1
  }
}
```

#### `linear_update_issue`
Update an existing issue. All fields except `id` are optional.

**Parameters:**
- `id` (string, required): Issue UUID
- `title` (string, optional): New title
- `description` (string, optional): New description
- `priority` (number, optional): New priority
- `assigneeID` (string, optional): New assignee (empty string to unassign)
- `stateID` (string, optional): New state
- `labelIds` (array, optional): Replace all labels

**Example:**
```json
{
  "name": "linear_update_issue",
  "arguments": {
    "id": "abc-123-uuid",
    "priority": 1,
    "title": "Updated title"
  }
}
```

#### `linear_delete_issue`
Delete an issue permanently.

**Parameters:**
- `id` (string, required): Issue UUID to delete

**Example:**
```json
{
  "name": "linear_delete_issue",
  "arguments": {
    "id": "abc-123-uuid"
  }
}
```

### Supporting Resources

#### `linear_list_workflow_states`
List workflow states (e.g., Todo, In Progress, Done) for getting state IDs.

**Parameters:**
- `first` (number, optional): Number of states to return
- `after` (string, optional): Pagination cursor

#### `linear_list_users`
List all users in the workspace for getting user IDs to assign issues.

**Parameters:**
- `first` (number, optional): Number of users to return
- `after` (string, optional): Pagination cursor

#### `linear_list_labels`
List all issue labels for getting label IDs.

**Parameters:**
- `first` (number, optional): Number of labels to return
- `after` (string, optional): Pagination cursor

### Comments

#### `linear_create_comment`
Add a comment to an issue.

**Parameters:**
- `issueID` (string, required): Issue UUID
- `body` (string, required): Comment text in markdown

**Example:**
```json
{
  "name": "linear_create_comment",
  "arguments": {
    "issueID": "abc-123-uuid",
    "body": "This is fixed in the latest release"
  }
}
```

#### `linear_list_comments`
List comments with pagination.

**Parameters:**
- `first` (number, optional): Number of comments to return
- `after` (string, optional): Pagination cursor

## Common Workflows

### Workflow 1: Create an Issue

```
1. Get team ID:
   linear_list_teams → teams[0].id

2. Create issue:
   linear_create_issue with:
   - teamID: (from step 1)
   - title: "Your issue title"
   - description: "Issue details"
   - priority: 1 (urgent)

3. Result contains:
   - issue.id: UUID for future operations
   - issue.url: Web link to share
   - issue.number: Issue number (e.g., 57 for HEX-57)
```

### Workflow 2: Search and Update Issues

```
1. Search for issues:
   linear_search_issues with:
   - query: "priority:urgent state:started"

2. Select issue from results → issue.id

3. Update the issue:
   linear_update_issue with:
   - id: (from step 2)
   - priority: 3 (downgrade to normal)
   - state: (move to completed)
```

### Workflow 3: Assign Issue to User

```
1. Get user ID:
   linear_get_viewer → viewer.id
   OR
   linear_list_users → users[0].id

2. Update issue with assignee:
   linear_update_issue with:
   - id: "issue-uuid"
   - assigneeID: (from step 1)
```

## Error Handling

The MCP server returns standard JSON-RPC errors:

### Authentication Error (401)
```json
{
  "error": {
    "code": -32000,
    "message": "Tool execution failed",
    "data": "authentication failed"
  }
}
```

**Solution:** Check your LINEAR_API_KEY environment variable.

### Permission Error (403)
```json
{
  "error": {
    "code": -32000,
    "message": "Tool execution failed",
    "data": "permission denied"
  }
}
```

**Solution:** Your API key lacks Write permission. Update scopes at https://linear.app/settings/account/security

### Rate Limit Error (429)
```json
{
  "error": {
    "code": -32000,
    "message": "Tool execution failed",
    "data": "rate limit exceeded"
  }
}
```

**Solution:** Linear allows ~2 requests/second sustained. The server automatically retries with backoff.

## Protocol Details

The MCP server implements JSON-RPC 2.0 over stdio:

### Initialize
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {}
}
```

### List Tools
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/list",
  "params": {}
}
```

### Call Tool
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "linear_list_teams",
    "arguments": {
      "first": 10
    }
  }
}
```

## Development

### Adding New Tools

1. Add tool definition to `mcp/tools.json`
2. Implement tool method in `cmd/linear-mcp/main.go`
3. Add to `handleToolsList()` method
4. Add case in `executeTool()` switch
5. Update this documentation

### Testing

Test the MCP server using stdio:

```bash
export LINEAR_API_KEY=lin_api_xxx
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | ./linear-mcp
```

Or use the MCP inspector:
```bash
npx @modelcontextprotocol/inspector ./linear-mcp
```

## Resources

- **MCP Specification:** https://modelcontextprotocol.io/specification
- **Linear API Docs:** https://developers.linear.app
- **go-linear Documentation:** ../README.md
- **Tool Definitions:** ../mcp/tools.json
- **JSON Schema:** ../pkg/linear/schema.json

## Security

- The MCP server requires LINEAR_API_KEY environment variable
- Never commit API keys to version control
- Use read-only API keys if you don't need write operations
- The server runs locally and communicates only with Linear's API

## Support

For issues or questions:
- GitHub Issues: https://github.com/chainguard-sandbox/go-linear/issues
- Linear API Status: https://status.linear.app
- MCP Documentation: https://modelcontextprotocol.io
