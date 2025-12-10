// Package main demonstrates how to interact with the linear-mcp server.
//
// This example shows how to:
//  1. Start the MCP server as a subprocess
//  2. Send JSON-RPC requests via stdio
//  3. Parse responses
//  4. Execute Linear operations through MCP tools
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/mcp-client/main.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCPClient wraps communication with the MCP server
type MCPClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
	nextID int
}

// NewMCPClient starts the MCP server and returns a client
func NewMCPClient(serverPath string) (*MCPClient, error) {
	cmd := exec.Command(serverPath)
	cmd.Env = append(os.Environ(), "LINEAR_API_KEY="+os.Getenv("LINEAR_API_KEY"))

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Capture stderr for debugging
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start MCP server: %w", err)
	}

	return &MCPClient{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewScanner(stdout),
		nextID: 1,
	}, nil
}

// Close shuts down the MCP server
func (c *MCPClient) Close() error {
	c.stdin.Close()
	return c.cmd.Wait()
}

// Call sends a JSON-RPC request and returns the response
func (c *MCPClient) Call(method string, params interface{}) (*JSONRPCResponse, error) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      c.nextID,
		Method:  method,
		Params:  params,
	}
	c.nextID++

	// Send request
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if _, err := c.stdin.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Read response
	if !c.stdout.Scan() {
		if err := c.stdout.Err(); err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}
		return nil, fmt.Errorf("no response from server")
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(c.stdout.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s - %v", resp.Error.Code, resp.Error.Message, resp.Error.Data)
	}

	return &resp, nil
}

// CallTool calls a Linear MCP tool
func (c *MCPClient) CallTool(name string, arguments map[string]interface{}) (string, error) {
	params := map[string]interface{}{
		"name":      name,
		"arguments": arguments,
	}

	resp, err := c.Call("tools/call", params)
	if err != nil {
		return "", err
	}

	// Parse tool response
	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return "", fmt.Errorf("failed to parse tool result: %w", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return result.Content[0].Text, nil
}

func main() {
	// Path to the MCP server binary (adjust if needed)
	serverPath := "../../cmd/linear-mcp/linear-mcp"

	// Check if server binary exists
	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		log.Fatal("MCP server not found. Build it first: cd cmd/linear-mcp && go build")
	}

	// Check API key
	if os.Getenv("LINEAR_API_KEY") == "" {
		log.Fatal("LINEAR_API_KEY environment variable not set")
	}

	fmt.Println("Starting Linear MCP client example...")

	// Create MCP client
	client, err := NewMCPClient(serverPath)
	if err != nil {
		log.Fatalf("Failed to create MCP client: %v", err)
	}
	defer client.Close()

	// Initialize the MCP connection
	fmt.Println("\n1. Initializing MCP connection...")
	initResp, err := client.Call("initialize", map[string]interface{}{})
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	fmt.Printf("✓ Initialized: %s\n", string(initResp.Result))

	// Get authenticated user
	fmt.Println("\n2. Getting authenticated user...")
	viewer, err := client.CallTool("linear_get_viewer", map[string]interface{}{})
	if err != nil {
		log.Fatalf("Failed to get viewer: %v", err)
	}
	fmt.Printf("✓ Authenticated user:\n%s\n", viewer)

	// List teams
	fmt.Println("\n3. Listing teams...")
	teams, err := client.CallTool("linear_list_teams", map[string]interface{}{
		"first": 5,
	})
	if err != nil {
		log.Fatalf("Failed to list teams: %v", err)
	}
	fmt.Printf("✓ Teams:\n%s\n", teams)

	// List issues
	fmt.Println("\n4. Listing recent issues...")
	issues, err := client.CallTool("linear_list_issues", map[string]interface{}{
		"first": 5,
	})
	if err != nil {
		log.Fatalf("Failed to list issues: %v", err)
	}
	fmt.Printf("✓ Recent issues:\n%s\n", issues)

	// Example: Create an issue (commented out to avoid creating test issues)
	/*
		fmt.Println("\n5. Creating a test issue...")
		newIssue, err := client.CallTool("linear_create_issue", map[string]interface{}{
			"teamID":      "YOUR_TEAM_ID_HERE",
			"title":       "Test issue from MCP client",
			"description": "This issue was created via the MCP protocol",
			"priority":    3, // Normal priority
		})
		if err != nil {
			log.Fatalf("Failed to create issue: %v", err)
		}
		fmt.Printf("✓ Created issue:\n%s\n", newIssue)
	*/

	fmt.Println("\n✓ MCP client example completed successfully!")
}
