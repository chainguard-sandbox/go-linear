package main

import (
	"bytes"
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

	fixed, err := fixDoubleEncodedFlags(req.Params)
	if err == nil && fixed != nil {
		req.Params = fixed
		return req, nil
	}
	// No fix applied — either nothing needed changing or the params were
	// malformed. Forward the original message unchanged; the downstream MCP
	// layer rejects anything invalid. (Behavior locked by
	// TestFixFlagsConn_Read_FixErrorFallsThrough.)
	return msg, nil
}

// fixDoubleEncodedFlags detects and fixes double-encoded flags in tools/call params.
// Returns nil if no fix was needed, or the corrected params if a fix was applied.
// Uses map-based unmarshaling to preserve all top-level fields (e.g. _meta/progressToken).
func fixDoubleEncodedFlags(raw json.RawMessage) (json.RawMessage, error) {
	var params map[string]json.RawMessage
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, err
	}

	// JSON null unmarshals to a nil map without error; the "arguments" lookup
	// below misses on a nil map, so null params fall through to (nil, nil).
	argsRaw, ok := params["arguments"]
	if !ok {
		return nil, nil
	}

	args, ok := asJSONObjectMap(argsRaw)
	if !ok {
		return nil, nil
	}

	flagsRaw, ok := args["flags"]
	if !ok {
		return nil, nil
	}

	// If flags is not a JSON string, it's already an object — no fix needed.
	flagsStr, ok := asJSONString(flagsRaw)
	if !ok {
		return nil, nil
	}

	// Decode the inner JSON string into an object; leave it alone if non-object or invalid.
	flagsObj, ok := asJSONObject([]byte(flagsStr))
	if !ok {
		return nil, nil
	}

	args["flags"] = flagsObj
	patchedArgs, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	params["arguments"] = patchedArgs
	return json.Marshal(params)
}

// asJSONObjectMap unmarshals raw into a JSON object map. ok is false if raw is
// not a JSON object, in which case the caller treats it as "nothing to fix".
func asJSONObjectMap(raw json.RawMessage) (map[string]json.RawMessage, bool) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, false
	}
	if m == nil { // JSON null unmarshals without error into a nil map
		return nil, false
	}
	return m, true
}

// asJSONString returns the string value if raw is a JSON string, else ("", false).
func asJSONString(raw json.RawMessage) (string, bool) {
	var s *string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", false
	}
	if s == nil { // JSON null unmarshals without error into a nil pointer
		return "", false
	}
	return *s, true
}

// asJSONObject returns data as a RawMessage if it is a valid JSON object, else (nil, false).
func asJSONObject(data []byte) (json.RawMessage, bool) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return nil, false
	}
	if !json.Valid(trimmed) {
		return nil, false
	}
	return json.RawMessage(trimmed), true
}
