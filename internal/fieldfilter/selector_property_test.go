package fieldfilter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// fieldError.Error() is exercised when an error is returned and its message is read.
func TestNew_ErrorMessage(t *testing.T) {
	_, err := New("id,,title", nil)
	if err == nil {
		t.Fatal("expected error for empty field in spec")
	}
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

// New("defaults", nil) with no command defaults resolves to an empty field list
// and should return a nil selector (no filtering).
func TestNew_DefaultsNoCommandDefaults(t *testing.T) {
	fs, err := New("defaults", nil)
	if err != nil {
		t.Fatalf("New(\"defaults\", nil): %v", err)
	}
	if fs != nil {
		t.Error("expected nil selector when defaults spec has no configured defaults")
	}
}

// NewForList returns nil when New returns nil (no filtering requested).
func TestNewForList_NilPassthrough(t *testing.T) {
	fs, err := NewForList("none", nil)
	if err != nil {
		t.Fatalf("NewForList(\"none\", nil): %v", err)
	}
	if fs != nil {
		t.Error("expected nil selector for 'none' spec")
	}
}

// NewForList propagates errors from New (e.g. empty field in spec).
func TestNewForList_PropagatesError(t *testing.T) {
	_, err := NewForList("id,,title", nil)
	if err == nil {
		t.Error("expected error for empty field in spec")
	}
}

// NewForList preserves pagination wrapper fields (nodes, pageInfo, totalCount)
// while filtering inner object fields — exercises filterObject's preserveFields branch.
func TestNewForList_PreservesWrapperFields(t *testing.T) {
	fs, err := NewForList("id,title", nil)
	if err != nil {
		t.Fatalf("NewForList: %v", err)
	}
	input := `{"nodes":[{"id":"1","title":"T","extra":"e"}],"pageInfo":{"hasNextPage":false},"totalCount":1}`
	out, err := fs.Filter([]byte(input))
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(out, &obj); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	for _, key := range []string{"nodes", "pageInfo", "totalCount"} {
		if _, ok := obj[key]; !ok {
			t.Errorf("wrapper field %q missing from output", key)
		}
	}
	nodesRaw, ok := obj["nodes"].([]any)
	if !ok || len(nodesRaw) == 0 {
		t.Fatal("nodes is missing or empty")
	}
	inner, ok := nodesRaw[0].(map[string]any)
	if !ok {
		t.Fatal("nodes[0] is not a JSON object")
	}
	if _, ok := inner["extra"]; ok {
		t.Error("inner field 'extra' should have been filtered out")
	}
}

// Filter returns an error on invalid JSON input.
func TestFilter_InvalidJSON(t *testing.T) {
	fs, _ := New("id", nil)
	_, err := fs.Filter([]byte(`{invalid json`))
	if err == nil {
		t.Error("expected error for invalid JSON input")
	}
}

// filterValue default branch: scalar values inside an array pass through unchanged.
func TestFilter_ScalarArrayElements(t *testing.T) {
	fs, _ := New("id", nil)
	out, err := fs.Filter([]byte(`["a", "b", "c"]`))
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	if !json.Valid(out) {
		t.Errorf("output is not valid JSON: %s", out)
	}
}

// Property: Filter is idempotent — applying the same spec twice equals applying it once.
func TestFilter_Idempotent(t *testing.T) {
	cases := []struct {
		name  string
		spec  string
		input string
	}{
		{"simple", "id", `{"id":"123","title":"Test","priority":1}`},
		{"multi", "id,title", `{"id":"1","title":"T","extra":"e"}`},
		{"nested", "id,assignee.name", `{"id":"1","assignee":{"name":"Alice","email":"a@b.com"}}`},
		{"array", "id,title", `[{"id":"1","title":"A","extra":"x"},{"id":"2","title":"B","extra":"y"}]`},
		{"no match", "missing", `{"id":"1","title":"T"}`},
		{"empty object", "id", `{}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fs, err := New(c.spec, nil)
			if err != nil {
				t.Fatalf("New(%q): %v", c.spec, err)
			}
			once, err := fs.Filter([]byte(c.input))
			if err != nil {
				t.Fatalf("Filter() first pass: %v", err)
			}
			twice, err := fs.Filter(once)
			if err != nil {
				t.Fatalf("Filter() second pass: %v", err)
			}
			if !jsonEqual(once, twice) {
				t.Errorf("Idempotent: first=%s second=%s", once, twice)
			}
		})
	}
}

// Property: every key in the filtered output was requested in the spec
// (directly or as the parent of a nested selector).
func TestFilter_OutputIsSubset(t *testing.T) {
	cases := []struct {
		name  string
		spec  string
		input string
	}{
		{"simple subset", "id,title", `{"id":"1","title":"T","extra":"e","priority":1}`},
		{"only id", "id", `{"id":"1","title":"T","extra":"e"}`},
		{"nested parent key", "assignee.name", `{"id":"1","assignee":{"name":"Alice","email":"a@b.com"}}`},
		{"nothing matches", "missing", `{"id":"1","title":"T"}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fs, err := New(c.spec, nil)
			if err != nil {
				t.Fatalf("New(%q): %v", c.spec, err)
			}
			out, err := fs.Filter([]byte(c.input))
			if err != nil {
				t.Fatalf("Filter(): %v", err)
			}

			var obj map[string]any
			if err := json.Unmarshal(out, &obj); err != nil {
				t.Fatalf("output is not a valid JSON object: %v — output: %s", err, out)
			}
			for key := range obj {
				if !keyInSpec(key, fs.fields) {
					t.Errorf("output key %q was not requested in spec %q", key, c.spec)
				}
			}
		})
	}
}

// Property: output is always valid JSON, even when no fields match.
func TestFilter_AlwaysValidJSON(t *testing.T) {
	cases := []struct{ spec, input string }{
		{"id", `{"id":"1","title":"T"}`},
		{"missing", `{"id":"1","title":"T"}`},
		{"id,title", `[{"id":"1"},{"id":"2"}]`},
		{"x", `{}`},
		{"x", `[]`},
	}
	for _, c := range cases {
		fs, err := New(c.spec, nil)
		if err != nil {
			t.Fatalf("New(%q): %v", c.spec, err)
		}
		out, err := fs.Filter([]byte(c.input))
		if err != nil {
			t.Errorf("Filter(%q, %q): %v", c.spec, c.input, err)
			continue
		}
		if !json.Valid(out) {
			t.Errorf("Filter(%q, %q) output is not valid JSON: %s", c.spec, c.input, out)
		}
	}
}

// Fuzz: Filter never panics on arbitrary JSON input with a valid spec.
func FuzzFilter_NoPanic(f *testing.F) {
	f.Add("id", `{"id":"123","title":"Test"}`)
	f.Add("id,title", `[{"id":"1","title":"A"},{"id":"2"}]`)
	f.Add("assignee.name", `{"assignee":{"name":"Alice"}}`)
	f.Add("x", `{}`)
	f.Add("id", `[]`)

	f.Fuzz(func(t *testing.T, spec string, input string) {
		fs, err := New(spec, nil)
		if err != nil {
			return // invalid spec is expected
		}
		if fs == nil {
			return // nil means no filtering
		}
		out, err := fs.Filter([]byte(input))
		if err != nil {
			return // invalid JSON input is expected
		}
		if !json.Valid(out) {
			t.Errorf("Filter(%q, %q) produced invalid JSON: %s", spec, input, out)
		}
	})
}

// keyInSpec returns true if key is directly in fields or is the parent of a nested field.
func keyInSpec(key string, fields map[string]bool) bool {
	if fields[key] {
		return true
	}
	prefix := key + "."
	for f := range fields {
		if strings.HasPrefix(f, prefix) {
			return true
		}
	}
	return false
}

func jsonEqual(a, b []byte) bool {
	var va, vb any
	if json.Unmarshal(a, &va) != nil || json.Unmarshal(b, &vb) != nil {
		return false
	}
	na, _ := json.Marshal(va)
	nb, _ := json.Marshal(vb)
	return bytes.Equal(na, nb)
}
