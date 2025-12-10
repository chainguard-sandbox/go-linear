// Package main implements an MCP (Model Context Protocol) server for Linear API.
//
// This server exposes Linear operations as MCP tools that AI agents can discover
// and use. It communicates over stdio using JSON-RPC 2.0 protocol.
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	linear-mcp
//
// The server listens on stdin for JSON-RPC requests and writes responses to stdout.
// This follows the MCP specification for AI agent integration.
//
// Supported Tools (13 total):
//
// Read-Only (Safe - 9 tools):
//   - linear_get_viewer: Get authenticated user
//   - linear_list_teams: List all teams
//   - linear_list_issues: List issues with pagination
//   - linear_get_issue: Get single issue by ID
//   - linear_search_issues: Search issues by query
//   - linear_list_workflow_states: List workflow states
//   - linear_list_users: List users
//   - linear_list_labels: List labels
//   - linear_list_comments: List comments
//
// Write (Non-Destructive - 3 tools):
//   - linear_create_issue: Create a new issue
//   - linear_update_issue: Update an existing issue
//   - linear_create_comment: Add comment to issue
//
// Destructive (1 tool):
//   - linear_delete_issue: ⚠️ Delete an issue (PERMANENT, marked with warnings)
//
// See: https://modelcontextprotocol.io/specification
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

const version = "0.1.0"

// Server implements an MCP server for Linear API
type Server struct {
	client *linear.Client
	reader *bufio.Reader
	writer *bufio.Writer
}

// NewServer creates a new MCP server
func NewServer(apiKey string) (*Server, error) {
	client, err := linear.NewClient(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create Linear client: %w", err)
	}

	return &Server{
		client: client,
		reader: bufio.NewReader(os.Stdin),
		writer: bufio.NewWriter(os.Stdout),
	}, nil
}

// Close cleans up resources
func (s *Server) Close() {
	if s.client != nil {
		_ = s.client.Close()
	}
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Run starts the MCP server loop
func (s *Server) Run(ctx context.Context) error {
	log.Println("Linear MCP server started")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := s.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("failed to read request: %w", err)
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(nil, -32700, "Parse error", nil)
			continue
		}

		s.handleRequest(ctx, &req)
	}
}

// handleRequest processes a single JSON-RPC request
func (s *Server) handleRequest(ctx context.Context, req *JSONRPCRequest) {
	if req.JSONRPC != "2.0" {
		s.sendError(req.ID, -32600, "Invalid Request", nil)
		return
	}

	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "tools/list":
		s.handleToolsList(req)
	case "tools/call":
		s.handleToolsCall(ctx, req)
	default:
		s.sendError(req.ID, -32601, "Method not found", nil)
	}
}

// handleInitialize handles the MCP initialize request
func (s *Server) handleInitialize(req *JSONRPCRequest) {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "linear-mcp",
			"version": version,
		},
	}
	s.sendResult(req.ID, result)
}

// handleToolsList returns the list of available tools
func (s *Server) handleToolsList(req *JSONRPCRequest) {
	// Note: Full tool definitions with safety markers are in mcp/tools.json
	// This returns simplified schemas for MCP protocol compatibility
	tools := []map[string]interface{}{
		{
			"name":        "linear_get_viewer",
			"description": "Get authenticated user information",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			"name":        "linear_list_teams",
			"description": "List all teams in the workspace",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"first": map[string]interface{}{
						"type":        "number",
						"description": "Number of teams to return (default: 50)",
					},
					"after": map[string]interface{}{
						"type":        "string",
						"description": "Cursor for pagination",
					},
				},
			},
		},
		{
			"name":        "linear_list_issues",
			"description": "List issues with pagination",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"first": map[string]interface{}{
						"type":        "number",
						"description": "Number of issues to return",
					},
					"after": map[string]interface{}{
						"type":        "string",
						"description": "Cursor for pagination",
					},
				},
			},
		},
		{
			"name":        "linear_get_issue",
			"description": "Get a single issue by ID",
			"inputSchema": map[string]interface{}{
				"type":     "object",
				"required": []string{"id"},
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Issue UUID",
					},
				},
			},
		},
		{
			"name":        "linear_search_issues",
			"description": "Search for issues by text query with operators",
			"inputSchema": map[string]interface{}{
				"type":     "object",
				"required": []string{"query"},
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query (supports priority:urgent, state:started, assignee:me, label:bug)",
					},
					"first": map[string]interface{}{
						"type":        "number",
						"description": "Number of results to return",
					},
					"after": map[string]interface{}{
						"type":        "string",
						"description": "Cursor for pagination",
					},
				},
			},
		},
		{
			"name":        "linear_create_issue",
			"description": "⚠️ MUTABLE: Create a new issue. Always confirm with user before executing.",
			"inputSchema": map[string]interface{}{
				"type":     "object",
				"required": []string{"teamID"},
				"properties": map[string]interface{}{
					"teamID": map[string]interface{}{
						"type":        "string",
						"description": "Team UUID (from linear_list_teams)",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Issue title",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Issue description in markdown",
					},
					"priority": map[string]interface{}{
						"type":        "number",
						"description": "Priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low",
						"enum":        []int{0, 1, 2, 3, 4},
					},
				},
			},
		},
		{
			"name":        "linear_update_issue",
			"description": "⚠️ MUTABLE: Update an existing issue. Always confirm with user before executing.",
			"inputSchema": map[string]interface{}{
				"type":     "object",
				"required": []string{"id"},
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Issue UUID to update",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "New title",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "New description",
					},
					"priority": map[string]interface{}{
						"type":        "number",
						"description": "New priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low",
						"enum":        []int{0, 1, 2, 3, 4},
					},
					"assigneeID": map[string]interface{}{
						"type":        "string",
						"description": "New assignee UUID (empty string to unassign)",
					},
					"stateID": map[string]interface{}{
						"type":        "string",
						"description": "New workflow state UUID",
					},
				},
			},
		},
		{
			"name":        "linear_delete_issue",
			"description": "⚠️ DESTRUCTIVE: Delete an issue permanently. Cannot be undone. Always confirm with user.",
			"inputSchema": map[string]interface{}{
				"type":     "object",
				"required": []string{"id"},
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Issue UUID to delete (WARNING: permanent)",
					},
				},
			},
		},
		{
			"name":        "linear_list_workflow_states",
			"description": "List all workflow states (e.g., Todo, In Progress, Done)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"first": map[string]interface{}{
						"type":        "number",
						"description": "Number of states to return",
					},
					"after": map[string]interface{}{
						"type":        "string",
						"description": "Cursor for pagination",
					},
				},
			},
		},
		{
			"name":        "linear_list_users",
			"description": "List all users in the workspace",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"first": map[string]interface{}{
						"type":        "number",
						"description": "Number of users to return",
					},
					"after": map[string]interface{}{
						"type":        "string",
						"description": "Cursor for pagination",
					},
				},
			},
		},
		{
			"name":        "linear_list_labels",
			"description": "List all issue labels",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"first": map[string]interface{}{
						"type":        "number",
						"description": "Number of labels to return",
					},
					"after": map[string]interface{}{
						"type":        "string",
						"description": "Cursor for pagination",
					},
				},
			},
		},
		{
			"name":        "linear_create_comment",
			"description": "⚠️ MUTABLE: Add a comment to an issue. Always confirm with user before executing.",
			"inputSchema": map[string]interface{}{
				"type":     "object",
				"required": []string{"issueID", "body"},
				"properties": map[string]interface{}{
					"issueID": map[string]interface{}{
						"type":        "string",
						"description": "Issue UUID to comment on",
					},
					"body": map[string]interface{}{
						"type":        "string",
						"description": "Comment text in markdown",
					},
				},
			},
		},
		{
			"name":        "linear_list_comments",
			"description": "List comments with pagination",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"first": map[string]interface{}{
						"type":        "number",
						"description": "Number of comments to return",
					},
					"after": map[string]interface{}{
						"type":        "string",
						"description": "Cursor for pagination",
					},
				},
			},
		},
	}

	s.sendResult(req.ID, map[string]interface{}{
		"tools": tools,
	})
}

// ToolCallParams represents parameters for a tool call
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// handleToolsCall executes a tool call
func (s *Server) handleToolsCall(ctx context.Context, req *JSONRPCRequest) {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	result, err := s.executeTool(ctx, params.Name, params.Arguments)
	if err != nil {
		s.sendError(req.ID, -32000, "Tool execution failed", err.Error())
		return
	}

	s.sendResult(req.ID, map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": result,
			},
		},
	})
}

// executeTool executes a specific Linear tool
func (s *Server) executeTool(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	switch name {
	case "linear_get_viewer":
		return s.getViewer(ctx)
	case "linear_list_teams":
		return s.listTeams(ctx, args)
	case "linear_list_issues":
		return s.listIssues(ctx, args)
	case "linear_get_issue":
		return s.getIssue(ctx, args)
	case "linear_search_issues":
		return s.searchIssues(ctx, args)
	case "linear_create_issue":
		return s.createIssue(ctx, args)
	case "linear_update_issue":
		return s.updateIssue(ctx, args)
	case "linear_delete_issue":
		return s.deleteIssue(ctx, args)
	case "linear_list_workflow_states":
		return s.listWorkflowStates(ctx, args)
	case "linear_list_users":
		return s.listUsers(ctx, args)
	case "linear_list_labels":
		return s.listLabels(ctx, args)
	case "linear_create_comment":
		return s.createComment(ctx, args)
	case "linear_list_comments":
		return s.listComments(ctx, args)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

// getViewer implements linear_get_viewer
func (s *Server) getViewer(ctx context.Context) (string, error) {
	viewer, err := s.client.Viewer(ctx)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(viewer, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// listTeams implements linear_list_teams
func (s *Server) listTeams(ctx context.Context, args map[string]interface{}) (string, error) {
	var first *int64
	if f, ok := args["first"].(float64); ok {
		val := int64(f)
		first = &val
	}

	var after *string
	if a, ok := args["after"].(string); ok {
		after = &a
	}

	teams, err := s.client.Teams(ctx, first, after)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(teams, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// listIssues implements linear_list_issues
func (s *Server) listIssues(ctx context.Context, args map[string]interface{}) (string, error) {
	var first *int64
	if f, ok := args["first"].(float64); ok {
		val := int64(f)
		first = &val
	}

	var after *string
	if a, ok := args["after"].(string); ok {
		after = &a
	}

	issues, err := s.client.Issues(ctx, first, after)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(issues, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// createIssue implements linear_create_issue
func (s *Server) createIssue(ctx context.Context, args map[string]interface{}) (string, error) {
	teamID, ok := args["teamID"].(string)
	if !ok || teamID == "" {
		return "", fmt.Errorf("teamID is required")
	}

	input := linear.IssueCreateInput{
		TeamID: teamID,
	}

	if title, ok := args["title"].(string); ok {
		input.Title = &title
	}

	if desc, ok := args["description"].(string); ok {
		input.Description = &desc
	}

	if priority, ok := args["priority"].(float64); ok {
		p := int64(priority)
		input.Priority = &p
	}

	issue, err := s.client.IssueCreate(ctx, input)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(issue, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// getIssue implements linear_get_issue
func (s *Server) getIssue(ctx context.Context, args map[string]interface{}) (string, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("id is required")
	}

	issue, err := s.client.Issue(ctx, id)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(issue, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// searchIssues implements linear_search_issues
func (s *Server) searchIssues(ctx context.Context, args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("query is required")
	}

	var first *int64
	if f, ok := args["first"].(float64); ok {
		val := int64(f)
		first = &val
	}

	var after *string
	if a, ok := args["after"].(string); ok {
		after = &a
	}

	issues, err := s.client.SearchIssues(ctx, query, first, after, nil, nil)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(issues, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// updateIssue implements linear_update_issue
func (s *Server) updateIssue(ctx context.Context, args map[string]interface{}) (string, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("id is required")
	}

	input := linear.IssueUpdateInput{}

	if title, ok := args["title"].(string); ok {
		input.Title = &title
	}

	if desc, ok := args["description"].(string); ok {
		input.Description = &desc
	}

	if priority, ok := args["priority"].(float64); ok {
		p := int64(priority)
		input.Priority = &p
	}

	if assigneeID, ok := args["assigneeID"].(string); ok {
		input.AssigneeID = &assigneeID
	}

	if stateID, ok := args["stateID"].(string); ok {
		input.StateID = &stateID
	}

	if labelIds, ok := args["labelIds"].([]interface{}); ok {
		labels := make([]string, len(labelIds))
		for i, v := range labelIds {
			if s, ok := v.(string); ok {
				labels[i] = s
			}
		}
		input.LabelIds = labels
	}

	issue, err := s.client.IssueUpdate(ctx, id, input)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(issue, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// deleteIssue implements linear_delete_issue
func (s *Server) deleteIssue(ctx context.Context, args map[string]interface{}) (string, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("id is required")
	}

	err := s.client.IssueDelete(ctx, id)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`{"success": true, "id": %q, "message": "Issue deleted permanently"}`, id), nil
}

// listWorkflowStates implements linear_list_workflow_states
func (s *Server) listWorkflowStates(ctx context.Context, args map[string]interface{}) (string, error) {
	var first *int64
	if f, ok := args["first"].(float64); ok {
		val := int64(f)
		first = &val
	}

	var after *string
	if a, ok := args["after"].(string); ok {
		after = &a
	}

	states, err := s.client.WorkflowStates(ctx, first, after)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// listUsers implements linear_list_users
func (s *Server) listUsers(ctx context.Context, args map[string]interface{}) (string, error) {
	var first *int64
	if f, ok := args["first"].(float64); ok {
		val := int64(f)
		first = &val
	}

	var after *string
	if a, ok := args["after"].(string); ok {
		after = &a
	}

	users, err := s.client.Users(ctx, first, after)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// listLabels implements linear_list_labels
func (s *Server) listLabels(ctx context.Context, args map[string]interface{}) (string, error) {
	var first *int64
	if f, ok := args["first"].(float64); ok {
		val := int64(f)
		first = &val
	}

	var after *string
	if a, ok := args["after"].(string); ok {
		after = &a
	}

	labels, err := s.client.IssueLabels(ctx, first, after)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(labels, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// createComment implements linear_create_comment
func (s *Server) createComment(ctx context.Context, args map[string]interface{}) (string, error) {
	issueID, ok := args["issueID"].(string)
	if !ok || issueID == "" {
		return "", fmt.Errorf("issueID is required")
	}

	body, ok := args["body"].(string)
	if !ok || body == "" {
		return "", fmt.Errorf("body is required")
	}

	input := linear.CommentCreateInput{
		IssueID: &issueID,
		Body:    &body,
	}

	comment, err := s.client.CommentCreate(ctx, input)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(comment, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// listComments implements linear_list_comments
func (s *Server) listComments(ctx context.Context, args map[string]interface{}) (string, error) {
	var first *int64
	if f, ok := args["first"].(float64); ok {
		val := int64(f)
		first = &val
	}

	var after *string
	if a, ok := args["after"].(string); ok {
		after = &a
	}

	comments, err := s.client.Comments(ctx, first, after)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(comments, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// sendResult sends a successful JSON-RPC response
func (s *Server) sendResult(id, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.send(resp)
}

// sendError sends an error JSON-RPC response
func (s *Server) sendError(id interface{}, code int, message string, data interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	s.send(resp)
}

// send writes a response to stdout
func (s *Server) send(resp JSONRPCResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}

	_, _ = s.writer.Write(data)
	_ = s.writer.WriteByte('\n')
	_ = s.writer.Flush()
}

func main() {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		log.Fatal("LINEAR_API_KEY environment variable not set")
	}

	server, err := NewServer(apiKey)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer server.Close()

	ctx := context.Background()
	err = server.Run(ctx)
	server.Close()
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
