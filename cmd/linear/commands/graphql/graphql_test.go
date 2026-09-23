package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// runCommand executes the graphql command in isolation (no root PersistentPreRunE,
// so no API key is needed) with the given args, returning the command error and
// whether the client factory was invoked.
func runCommand(t *testing.T, args ...string) (factoryCalled bool, err error) {
	t.Helper()
	factory := cli.ClientFactory(func() (*linear.Client, error) {
		factoryCalled = true
		return nil, fmt.Errorf("factory should not have been called")
	})
	cmd := NewGraphQLCommand(factory)
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return factoryCalled, err
}

func TestRun_EmptyQueryNeverCallsFactory(t *testing.T) {
	called, err := runCommand(t, "--query", "   ")
	if err == nil {
		t.Fatal("expected error for empty --query")
	}
	if !strings.Contains(err.Error(), "--query is required") {
		t.Errorf("error = %v, want '--query is required'", err)
	}
	if called {
		t.Error("client factory was called despite empty query (should be pre-network)")
	}
}

func TestRun_MutationGateNeverCallsFactory(t *testing.T) {
	called, err := runCommand(t, "--query", `mutation { issueUpdate(id:"x",input:{priority:1}){ success } }`)
	if err == nil {
		t.Fatal("expected error for mutation without --allow-mutation")
	}
	if !strings.Contains(err.Error(), "--allow-mutation") {
		t.Errorf("error = %v, want mutation refusal", err)
	}
	if called {
		t.Error("client factory was called despite blocked mutation (gate must run pre-network)")
	}
}

func TestParseVar(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		wantName string
		want     any
		wantErr  bool
	}{
		{name: "string default", in: "id=abc-123", wantName: "id", want: "abc-123"},
		{name: "explicit string", in: "title:string=Hello World", wantName: "title", want: "Hello World"},
		{name: "int", in: "first:int=50", wantName: "first", want: int64(50)},
		{name: "float", in: "estimate:float=1.5", wantName: "estimate", want: 1.5},
		{name: "bool", in: "active:bool=true", wantName: "active", want: true},
		{name: "json object", in: `filter:json={"state":"open"}`, wantName: "filter", want: map[string]any{"state": "open"}},
		{name: "json array", in: "ids:json=[1,2,3]", wantName: "ids", want: []any{float64(1), float64(2), float64(3)}},
		{name: "value with equals", in: "q:string=a=b", wantName: "q", want: "a=b"},
		{name: "empty value", in: "note=", wantName: "note", want: ""},
		{name: "no equals", in: "broken", wantErr: true},
		{name: "empty name", in: "=v", wantErr: true},
		{name: "bad int", in: "n:int=x", wantErr: true},
		{name: "bad bool", in: "b:bool=maybe", wantErr: true},
		{name: "bad json", in: "j:json={", wantErr: true},
		{name: "unknown type", in: "x:date=2024", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, got, err := parseVar(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseVar(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if gotName != tt.wantName {
				t.Errorf("name = %q, want %q", gotName, tt.wantName)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("value = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestLoadQuery(t *testing.T) {
	t.Run("inline", func(t *testing.T) {
		got, err := loadQuery("query { viewer { id } }", nil)
		if err != nil {
			t.Fatal(err)
		}
		if got != "query { viewer { id } }" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("stdin", func(t *testing.T) {
		got, err := loadQuery("-", strings.NewReader("query { me }"))
		if err != nil {
			t.Fatal(err)
		}
		if got != "query { me }" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "q.graphql")
		if err := os.WriteFile(path, []byte("query { fromfile }"), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := loadQuery("@"+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got != "query { fromfile }" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("@- stdin", func(t *testing.T) {
		got, err := loadQuery("@-", strings.NewReader("query { atdash }"))
		if err != nil {
			t.Fatal(err)
		}
		if got != "query { atdash }" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		if _, err := loadQuery("@/no/such/file.graphql", nil); err == nil {
			t.Error("expected error for missing file")
		}
	})
}

func TestBuildVars_MergeFileAndFlags(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vars.json")
	if err := os.WriteFile(path, []byte(`{"id":"from-file","first":10}`), 0o600); err != nil {
		t.Fatal(err)
	}

	// --var id overrides the file; new key added; first stays from file.
	vars, err := buildVars("@"+path, []string{"id=from-flag", "active:bool=true"})
	if err != nil {
		t.Fatal(err)
	}
	if vars["id"] != "from-flag" {
		t.Errorf("id = %v, want from-flag (flag should override file)", vars["id"])
	}
	if vars["active"] != true {
		t.Errorf("active = %v, want true", vars["active"])
	}
	if vars["first"] != float64(10) {
		t.Errorf("first = %v, want 10 (from file)", vars["first"])
	}
}

func TestBuildVars_Errors(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		if _, err := buildVars("@/no/such/vars.json", nil); err == nil || !strings.Contains(err.Error(), "read vars file") {
			t.Errorf("err = %v, want 'read vars file'", err)
		}
	})
	t.Run("non-object json", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "arr.json")
		if err := os.WriteFile(path, []byte(`[1,2,3]`), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := buildVars("@"+path, nil); err == nil || !strings.Contains(err.Error(), "JSON object") {
			t.Errorf("err = %v, want 'JSON object'", err)
		}
	})

	// JSON null must not nil the map and panic on a later --var assignment.
	t.Run("null does not panic", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "null.json")
		if err := os.WriteFile(path, []byte(`null`), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := buildVars("@"+path, []string{"x=y"}); err == nil || !strings.Contains(err.Error(), "must be a JSON object") {
			t.Errorf("err = %v, want 'must be a JSON object'", err)
		}
	})
}

func TestBuildVars_Empty(t *testing.T) {
	vars, err := buildVars("", nil)
	if err != nil {
		t.Fatal(err)
	}
	if vars != nil {
		t.Errorf("expected nil vars, got %#v", vars)
	}
}

func TestCheckOperation(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		operationName string
		allowMutation bool
		wantErr       bool
	}{
		{name: "query allowed", query: `query { viewer { id } }`},
		{name: "anonymous query allowed", query: `{ viewer { id } }`},
		{name: "mutation blocked by default", query: `mutation { issueUpdate(id:"x",input:{}){ success } }`, wantErr: true},
		{name: "mutation allowed with flag", query: `mutation { issueUpdate(id:"x",input:{}){ success } }`, allowMutation: true},
		{name: "subscription rejected", query: `subscription { issues { id } }`, wantErr: true},
		{name: "multi-op needs name", query: `query A { viewer { id } } query B { viewer { name } }`, wantErr: true},
		{name: "multi-op select query", query: `query A { viewer { id } } mutation B { x { success } }`, operationName: "A"},
		{name: "multi-op select mutation blocked", query: `query A { viewer { id } } mutation B { x { success } }`, operationName: "B", wantErr: true},
		{name: "multi-op select mutation allowed", query: `query A { viewer { id } } mutation B { x { success } }`, operationName: "B", allowMutation: true},
		{name: "unknown operation name", query: `query A { viewer { id } }`, operationName: "Z", wantErr: true},
		{name: "syntax error", query: `query { viewer {`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkOperation(tt.query, tt.operationName, tt.allowMutation)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkOperation() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrintJSON(t *testing.T) {
	data := json.RawMessage(`{"viewer":{"id":"u1"}}`)

	t.Run("pretty", func(t *testing.T) {
		var buf bytes.Buffer
		if err := printJSON(&buf, data, false); err != nil {
			t.Fatal(err)
		}
		out := buf.String()
		if !strings.Contains(out, "\n  \"viewer\"") {
			t.Errorf("expected indented output, got %q", out)
		}
	})

	t.Run("compact", func(t *testing.T) {
		var buf bytes.Buffer
		if err := printJSON(&buf, data, true); err != nil {
			t.Fatal(err)
		}
		out := strings.TrimSpace(buf.String())
		if out != `{"viewer":{"id":"u1"}}` {
			t.Errorf("compact output = %q", out)
		}
	})

	t.Run("empty", func(t *testing.T) {
		var buf bytes.Buffer
		if err := printJSON(&buf, nil, false); err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(buf.String()) != "null" {
			t.Errorf("empty output = %q", buf.String())
		}
	})
}
