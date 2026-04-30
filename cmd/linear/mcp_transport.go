package main

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// fixFlagsTransport wraps an mcp.Transport to normalize double-encoded flags.
//
// Some MCP clients (e.g. Claude Code) intermittently JSON-encode the
// `arguments.flags` value twice, sending a string where the schema expects an
// object. This transport intercepts tools/call requests before schema validation
// and decodes the inner string back into an object.
type fixFlagsTransport struct {
	inner mcp.Transport
}

func (t *fixFlagsTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	conn, err := t.inner.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return &fixFlagsConn{delegate: conn}, nil
}

type fixFlagsConn struct {
	delegate mcp.Connection
}

func (c *fixFlagsConn) SessionID() string                          { return c.delegate.SessionID() }
func (c *fixFlagsConn) Write(ctx context.Context, msg jsonrpc.Message) error {
	return c.delegate.Write(ctx, msg)
}
func (c *fixFlagsConn) Close() error { return c.delegate.Close() }

func (c *fixFlagsConn) Read(ctx context.Context) (jsonrpc.Message, error) {
	msg, err := c.delegate.Read(ctx)
	if err != nil {
		return msg, err
	}

	req, ok := msg.(*jsonrpc.Request)
	if !ok || req.Method != "tools/call" {
		return msg, nil
	}

	fixed, err := fixDoubleEncodedFlags(req.Params)
	if err != nil || fixed == nil {
		return msg, nil
	}
	req.Params = fixed
	return req, nil
}

// fixDoubleEncodedFlags detects and fixes double-encoded flags in tools/call params.
// Returns nil if no fix was needed, or the corrected params if a fix was applied.
func fixDoubleEncodedFlags(raw json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Name      string                     `json:"name"`
		Arguments map[string]json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, err
	}

	flagsRaw, ok := params.Arguments["flags"]
	if !ok {
		return nil, nil
	}

	// Check if flags is a JSON string (double-encoded)
	var flagsStr string
	if err := json.Unmarshal(flagsRaw, &flagsStr); err != nil {
		return nil, nil // flags is already an object, no fix needed
	}

	// Decode the inner JSON string into an object
	var flagsObj json.RawMessage
	if err := json.Unmarshal([]byte(flagsStr), &flagsObj); err != nil {
		return nil, nil // inner value isn't valid JSON, leave it alone
	}

	params.Arguments["flags"] = flagsObj
	return json.Marshal(params)
}
