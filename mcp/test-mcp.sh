#!/bin/bash
# Test script for the Linear MCP server
#
# This script sends test JSON-RPC requests to the MCP server via stdio
# and verifies it responds correctly.
#
# Usage:
#   export LINEAR_API_KEY=lin_api_xxx
#   chmod +x test-mcp.sh
#   ./test-mcp.sh

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

MCP_SERVER="../cmd/linear-mcp/linear-mcp"

# Check if server binary exists
if [ ! -f "$MCP_SERVER" ]; then
    echo -e "${RED}✗ MCP server binary not found at $MCP_SERVER${NC}"
    echo "Build it first: cd ../cmd/linear-mcp && go build"
    exit 1
fi

# Check if LINEAR_API_KEY is set
if [ -z "$LINEAR_API_KEY" ]; then
    echo -e "${RED}✗ LINEAR_API_KEY environment variable not set${NC}"
    exit 1
fi

echo -e "${YELLOW}Testing Linear MCP Server${NC}\n"

# Test 1: Initialize
echo -e "${YELLOW}Test 1: Initialize${NC}"
RESPONSE=$(echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | $MCP_SERVER 2>/dev/null | head -1)
if echo "$RESPONSE" | grep -q '"protocolVersion"'; then
    echo -e "${GREEN}✓ Initialize successful${NC}"
    echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
else
    echo -e "${RED}✗ Initialize failed${NC}"
    echo "$RESPONSE"
    exit 1
fi

echo ""

# Test 2: List tools
echo -e "${YELLOW}Test 2: List Tools${NC}"
RESPONSE=$(echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | $MCP_SERVER 2>/dev/null | head -1)
if echo "$RESPONSE" | grep -q '"tools"'; then
    echo -e "${GREEN}✓ List tools successful${NC}"
    # Extract just the tool names
    echo "$RESPONSE" | python3 -c "import sys, json; data = json.load(sys.stdin); print('Available tools:', ', '.join([t['name'] for t in data['result']['tools']]))" 2>/dev/null || echo "$RESPONSE"
else
    echo -e "${RED}✗ List tools failed${NC}"
    echo "$RESPONSE"
    exit 1
fi

echo ""

# Test 3: Call linear_get_viewer
echo -e "${YELLOW}Test 3: Get Viewer (Authentication Test)${NC}"
REQUEST='{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"linear_get_viewer","arguments":{}}}'
RESPONSE=$(echo "$REQUEST" | $MCP_SERVER 2>/dev/null | head -1)
if echo "$RESPONSE" | grep -q '"content"'; then
    echo -e "${GREEN}✓ Get viewer successful${NC}"
    echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
else
    echo -e "${RED}✗ Get viewer failed${NC}"
    echo "$RESPONSE"
    exit 1
fi

echo ""

# Test 4: List teams
echo -e "${YELLOW}Test 4: List Teams${NC}"
REQUEST='{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"linear_list_teams","arguments":{"first":3}}}'
RESPONSE=$(echo "$REQUEST" | $MCP_SERVER 2>/dev/null | head -1)
if echo "$RESPONSE" | grep -q '"content"'; then
    echo -e "${GREEN}✓ List teams successful${NC}"
    # Show just the team names
    echo "$RESPONSE" | python3 -c "import sys, json; data = json.load(sys.stdin); teams = json.loads(data['result']['content'][0]['text']); print('Teams:', ', '.join([t['name'] for t in teams.get('nodes', [])]))" 2>/dev/null || echo "$RESPONSE"
else
    echo -e "${RED}✗ List teams failed${NC}"
    echo "$RESPONSE"
    exit 1
fi

echo ""
echo -e "${GREEN}✓ All tests passed!${NC}"
echo ""
echo "The MCP server is working correctly and can:"
echo "  - Initialize MCP protocol connection"
echo "  - List available tools"
echo "  - Authenticate with Linear API"
echo "  - Execute Linear operations (get viewer, list teams)"
echo ""
echo "Next steps:"
echo "  1. Configure Claude Desktop (see docs/MCP.md)"
echo "  2. Try the full example: cd ../examples/mcp-client && go run main.go"
