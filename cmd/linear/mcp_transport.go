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

func (c *fixFlagsConn) SessionID() string { return c.delegate.SessionID() }
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

	fixed, _ := fixDoubleEncodedFlags(req.Params)
	if fixed == nil {
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

	// If flags is not a JSON string, it's already an object — no fix needed.
	flagsStr, ok := asJSONString(flagsRaw)
	if !ok {
		return nil, nil
	}

	// Decode the inner JSON string into an object; leave it alone if invalid.
	flagsObj, ok := asJSONObject([]byte(flagsStr))
	if !ok {
		return nil, nil
	}

	params.Arguments["flags"] = flagsObj
	return json.Marshal(params)
}

// asJSONString returns the string value if raw is a JSON string, else ("", false).
func asJSONString(raw json.RawMessage) (string, bool) {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", false
	}
	return s, true
}

// asJSONObject returns raw as a RawMessage if it is valid JSON object/array, else (nil, false).
func asJSONObject(data []byte) (json.RawMessage, bool) {
	var obj json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, false
	}
	return obj, true
}
