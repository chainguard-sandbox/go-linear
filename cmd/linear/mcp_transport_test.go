package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
)

func TestFixDoubleEncodedFlags(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNil bool // true = no fix needed
		wantOut string
	}{
		{
			name:    "already an object - no fix",
			input:   `{"name":"go-linear_issue_update","arguments":{"flags":{"state":"Done"}}}`,
			wantNil: true,
		},
		{
			name:    "double-encoded flags - fixed",
			input:   `{"name":"go-linear_issue_update","arguments":{"flags":"{\"state\":\"Done\"}"}}`,
			wantOut: `{"name":"go-linear_issue_update","arguments":{"flags":{"state":"Done"}}}`,
		},
		{
			name:    "no flags key - no fix",
			input:   `{"name":"go-linear_issue_list","arguments":{}}`,
			wantNil: true,
		},
		{
			name:    "double-encoded with multiple fields",
			input:   `{"name":"go-linear_issue_update","arguments":{"flags":"{\"issue\":\"ENG-1\",\"body\":\"test\"}"}}`,
			wantOut: `{"name":"go-linear_issue_update","arguments":{"flags":{"issue":"ENG-1","body":"test"}}}`,
		},

		// _meta / progressToken preservation (the headline fix in this PR).
		{
			name:    "_meta is preserved through fix",
			input:   `{"name":"go-linear_issue_update","arguments":{"flags":"{\"state\":\"Done\"}"},"_meta":{"progressToken":"abc-123"}}`,
			wantOut: `{"name":"go-linear_issue_update","arguments":{"flags":{"state":"Done"}},"_meta":{"progressToken":"abc-123"}}`,
		},
		{
			name:    "unknown top-level field is preserved",
			input:   `{"name":"x","arguments":{"flags":"{\"a\":1}"},"customField":[1,2,3]}`,
			wantOut: `{"name":"x","arguments":{"flags":{"a":1}},"customField":[1,2,3]}`,
		},

		// asJSONObject hardening: inner non-object JSON must be rejected.
		{
			name:    "flags string decodes to null - no fix",
			input:   `{"name":"x","arguments":{"flags":"null"}}`,
			wantNil: true,
		},
		{
			name:    "flags string decodes to array - no fix",
			input:   `{"name":"x","arguments":{"flags":"[1,2,3]"}}`,
			wantNil: true,
		},
		{
			name:    "flags string decodes to scalar number - no fix",
			input:   `{"name":"x","arguments":{"flags":"42"}}`,
			wantNil: true,
		},
		{
			name:    "flags string decodes to scalar string - no fix",
			input:   `{"name":"x","arguments":{"flags":"\"hello\""}}`,
			wantNil: true,
		},
		{
			name:    "flags string is malformed JSON - no fix",
			input:   `{"name":"x","arguments":{"flags":"{not valid"}}`,
			wantNil: true,
		},
		{
			name:    "flags object with leading whitespace is accepted",
			input:   `{"name":"x","arguments":{"flags":"   {\"a\":1}"}}`,
			wantOut: `{"name":"x","arguments":{"flags":{"a":1}}}`,
		},

		// Shape variations on arguments / flags.
		{
			name:    "arguments missing entirely - no fix",
			input:   `{"name":"x"}`,
			wantNil: true,
		},
		{
			name:    "arguments is null - no fix",
			input:   `{"name":"x","arguments":null}`,
			wantNil: true,
		},
		{
			name:    "arguments is not an object - no fix",
			input:   `{"name":"x","arguments":[1,2]}`,
			wantNil: true,
		},
		{
			name:    "flags is already a non-string scalar - no fix",
			input:   `{"name":"x","arguments":{"flags":42}}`,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fixDoubleEncodedFlags(json.RawMessage(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %s", result)
				}
				return
			}
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			// Normalize by round-tripping through map to avoid key-order differences
			var got, want map[string]any
			if err := json.Unmarshal(result, &got); err != nil {
				t.Fatalf("result is invalid JSON: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.wantOut), &want); err != nil {
				t.Fatalf("wantOut is invalid JSON: %v", err)
			}
			gotB, _ := json.Marshal(got)
			wantB, _ := json.Marshal(want)
			if !bytes.Equal(gotB, wantB) {
				t.Errorf("got %s, want %s", gotB, wantB)
			}
		})
	}
}

// asJSONObjectMap returns ok=false for anything that is not a JSON object.
// JSON null is the subtle case: json.Unmarshal accepts null into a map with a
// nil error and a nil map, so it must be rejected explicitly.
func TestAsJSONObjectMap(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOK  bool
		wantKey string // if set, key must be present in the returned map
	}{
		{
			name:   "null is rejected",
			input:  `null`,
			wantOK: false,
		},
		{
			// encoding/json accepts surrounding whitespace, so this is
			// caught by the nil-map guard, not the parse-error path.
			name:   "whitespace-padded null is rejected",
			input:  ` null `,
			wantOK: false,
		},
		{
			name:   "empty object is accepted",
			input:  `{}`,
			wantOK: true,
		},
		{
			name:    "object with key is accepted",
			input:   `{"a":1}`,
			wantOK:  true,
			wantKey: "a",
		},
		{
			name:    "object with nested object value is accepted",
			input:   `{"a":{"b":1}}`,
			wantOK:  true,
			wantKey: "a",
		},
		{
			name:    "object with multiple keys is accepted",
			input:   `{"a":1,"b":2}`,
			wantOK:  true,
			wantKey: "b",
		},
		{
			name:   "empty input is rejected",
			input:  ``,
			wantOK: false,
		},
		{
			name:   "array is rejected",
			input:  `[1]`,
			wantOK: false,
		},
		{
			name:   "scalar number is rejected",
			input:  `42`,
			wantOK: false,
		},
		{
			name:   "scalar string is rejected",
			input:  `"s"`,
			wantOK: false,
		},
		{
			name:   "malformed JSON is rejected",
			input:  `{malformed`,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, ok := asJSONObjectMap(json.RawMessage(tt.input))
			if ok != tt.wantOK {
				t.Fatalf("asJSONObjectMap(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if !ok {
				if m != nil {
					t.Errorf("asJSONObjectMap(%q) returned non-nil map %v with ok=false", tt.input, m)
				}
				return
			}
			if m == nil {
				t.Errorf("asJSONObjectMap(%q) returned nil map with ok=true", tt.input)
			}
			if tt.wantKey != "" {
				if _, present := m[tt.wantKey]; !present {
					t.Errorf("asJSONObjectMap(%q) missing expected key %q: %v", tt.input, tt.wantKey, m)
				}
			}
		})
	}
}

func TestAsJSONString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantOK bool
		want   string // expected value when wantOK
	}{
		{
			name:   "null is rejected",
			input:  `null`,
			wantOK: false,
		},
		{
			// encoding/json accepts surrounding whitespace, so this is
			// caught by the nil-pointer guard, not the parse-error path.
			name:   "whitespace-padded null is rejected",
			input:  ` null `,
			wantOK: false,
		},
		{
			name:   "string is accepted",
			input:  `"s"`,
			wantOK: true,
			want:   "s",
		},
		{
			name:   "empty string is accepted",
			input:  `""`,
			wantOK: true,
			want:   "",
		},
		{
			name:   "object is rejected",
			input:  `{}`,
			wantOK: false,
		},
		{
			name:   "array is rejected",
			input:  `["s"]`,
			wantOK: false,
		},
		{
			name:   "scalar number is rejected",
			input:  `42`,
			wantOK: false,
		},
		{
			name:   "empty input is rejected",
			input:  ``,
			wantOK: false,
		},
		{
			name:   "malformed JSON is rejected",
			input:  `"unterminated`,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, ok := asJSONString(json.RawMessage(tt.input))
			if ok != tt.wantOK {
				t.Fatalf("asJSONString(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if !ok {
				if s != "" {
					t.Errorf("asJSONString(%q) returned %q with ok=false, want empty string", tt.input, s)
				}
				return
			}
			if s != tt.want {
				t.Errorf("asJSONString(%q) = %q, want %q", tt.input, s, tt.want)
			}
		})
	}
}

// Outer-params parse error must be returned to the caller, not swallowed.
func TestFixDoubleEncodedFlags_OuterParseError(t *testing.T) {
	cases := []string{
		`{not valid json`,
		`[1,2,3]`, // valid JSON but not an object — cannot unmarshal into map
		``,
	}
	for _, in := range cases {
		result, err := fixDoubleEncodedFlags(json.RawMessage(in))
		if err == nil {
			t.Errorf("input %q: expected error, got result=%s", in, result)
		}
		if result != nil {
			t.Errorf("input %q: expected nil result on error, got %s", in, result)
		}
	}
}

// fakeConn is a minimal mcp.Connection for exercising fixFlagsConn.Read.
type fakeConn struct {
	msg     jsonrpc.Message
	readErr error
}

func (f *fakeConn) SessionID() string                             { return "fake" }
func (f *fakeConn) Read(context.Context) (jsonrpc.Message, error) { return f.msg, f.readErr }
func (f *fakeConn) Write(context.Context, jsonrpc.Message) error  { return nil }
func (f *fakeConn) Close() error                                  { return nil }

// Read propagates the delegate's read error untouched.
func TestFixFlagsConn_Read_DelegateError(t *testing.T) {
	wantErr := errors.New("boom")
	c := &fixFlagsConn{delegate: &fakeConn{readErr: wantErr}}
	_, err := c.Read(context.Background())
	if !errors.Is(err, wantErr) {
		t.Errorf("expected delegate error %v, got %v", wantErr, err)
	}
}

// Read passes non-tools/call requests through unchanged.
func TestFixFlagsConn_Read_NonToolsCall(t *testing.T) {
	req := &jsonrpc.Request{Method: "initialize", Params: json.RawMessage(`{}`)}
	c := &fixFlagsConn{delegate: &fakeConn{msg: req}}
	got, err := c.Read(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != req {
		t.Errorf("expected the same message reference, got %#v", got)
	}
}

// Read on a tools/call with double-encoded flags rewrites params in-place.
func TestFixFlagsConn_Read_FixesParams(t *testing.T) {
	req := &jsonrpc.Request{
		Method: "tools/call",
		Params: json.RawMessage(`{"name":"x","arguments":{"flags":"{\"a\":1}"}}`),
	}
	c := &fixFlagsConn{delegate: &fakeConn{msg: req}}
	got, err := c.Read(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	gotReq, ok := got.(*jsonrpc.Request)
	if !ok {
		t.Fatalf("expected *jsonrpc.Request, got %T", got)
	}
	var params map[string]any
	if err := json.Unmarshal(gotReq.Params, &params); err != nil {
		t.Fatalf("rewritten params not valid JSON: %v", err)
	}
	args, _ := params["arguments"].(map[string]any)
	flags, _ := args["flags"].(map[string]any)
	if flags["a"] != float64(1) {
		t.Errorf("expected rewritten flags.a == 1, got %v (full params: %s)", flags["a"], gotReq.Params)
	}
}

// Read on a tools/call where the fix function errors falls through silently.
// This test locks in current behavior; if a future change wants fail-closed
// semantics, it must update this test deliberately.
func TestFixFlagsConn_Read_FixErrorFallsThrough(t *testing.T) {
	req := &jsonrpc.Request{
		Method: "tools/call",
		Params: json.RawMessage(`{not valid json`),
	}
	c := &fixFlagsConn{delegate: &fakeConn{msg: req}}
	got, err := c.Read(context.Background())
	if err != nil {
		t.Errorf("Read should swallow fix-function errors, got %v", err)
	}
	if got != req {
		t.Errorf("expected original message reference on fix error, got %#v", got)
	}
}
