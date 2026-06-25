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

// New("defaults,...") propagates parseFields errors from the extra-fields segment.
func TestNew_DefaultsWithInvalidExtra(t *testing.T) {
	_, err := New("defaults,id,,title", nil)
	if err == nil {
		t.Error("expected error for empty field in defaults extra segment")
	}
}

// NewForList with a spec that names a wrapper field: "nodes" appears in both
// preserveFields and fs.fields, so filterObject takes the preserveFields branch
// (which applies filterValue recursively) rather than the direct-copy branch.
// Inner objects must still be filtered to contain only the spec fields.
func TestNewForList_OverlapPreservesAndFilters(t *testing.T) {
	fs, err := NewForList("nodes,id", nil)
	if err != nil {
		t.Fatalf("NewForList: %v", err)
	}
	input := `{"nodes":[{"id":"1","extra":"drop"}],"pageInfo":{"hasNextPage":false},"totalCount":1}`
	out, err := fs.Filter([]byte(input))
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(out, &obj); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	nodesRaw, ok := obj["nodes"].([]any)
	if !ok || len(nodesRaw) == 0 {
		t.Fatal("nodes missing or empty")
	}
	inner, ok := nodesRaw[0].(map[string]any)
	if !ok {
		t.Fatal("nodes[0] is not a JSON object")
	}
	if _, ok := inner["id"]; !ok {
		t.Error("inner field 'id' should be present")
	}
	if _, ok := inner["extra"]; ok {
		t.Error("inner field 'extra' should have been filtered out by filterValue on the preserveFields path")
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
	input := `{"nodes":[{"id":"1","title":"T","extra":"e"}],"pageInfo":{"hasNextPage":false},"totalCount":1,"status":"open"}`
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
	if _, ok := obj["status"]; ok {
		t.Error("non-wrapper top-level key 'status' should have been filtered out")
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
	for _, key := range []string{"id", "title"} {
		if _, ok := inner[key]; !ok {
			t.Errorf("inner field %q should be present", key)
		}
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
	want := []string{"a", "b", "c"}
	input, _ := json.Marshal(want)
	out, err := fs.Filter(input)
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	var got []string
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("output is not a valid JSON string array: %v — output: %s", err, out)
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d — output: %s", len(got), len(want), out)
	}
	for i, v := range got {
		if v != want[i] {
			t.Errorf("[%d] = %q, want %q", i, v, want[i])
		}
	}
}

// Property: Filter is idempotent — applying the same spec twice equals applying it once.
// Cases with an excludedKey also assert that the first pass actually removes that key,
// so a no-op Filter implementation cannot pass the test.
func TestFilter_Idempotent(t *testing.T) {
	cases := []struct {
		name        string
		spec        string
		input       string
		excludedKey string // top-level key that must be absent from the first-pass output
	}{
		{"simple", "id", `{"id":"123","title":"Test","priority":1}`, "priority"},
		{"multi", "id,title", `{"id":"1","title":"T","extra":"e"}`, "extra"},
		{"nested", "id,assignee.name", `{"id":"1","assignee":{"name":"Alice","email":"a@b.com"}}`, ""},
		{"array", "id,title", `[{"id":"1","title":"A","extra":"x"},{"id":"2","title":"B","extra":"y"}]`, ""},
		{"no match", "missing", `{"id":"1","title":"T"}`, ""},
		{"empty object", "id", `{}`, ""},
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
			if c.excludedKey != "" {
				var obj map[string]any
				if err := json.Unmarshal(once, &obj); err == nil {
					if _, present := obj[c.excludedKey]; present {
						t.Errorf("Filter() first pass still contains excluded key %q", c.excludedKey)
					}
				}
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
		name       string
		spec       string
		input      string
		wantObjLen int // expected number of top-level keys in filtered output
	}{
		{"simple subset", "id,title", `{"id":"1","title":"T","extra":"e","priority":1}`, 2},
		{"only id", "id", `{"id":"1","title":"T","extra":"e"}`, 1},
		{"nested parent key", "assignee.name", `{"id":"1","assignee":{"name":"Alice","email":"a@b.com"}}`, 1},
		// "nothing matches" produces {}; wantObjLen=0 makes the subset invariant non-vacuous.
		{"nothing matches", "missing", `{"id":"1","title":"T"}`, 0},
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
			if len(obj) != c.wantObjLen {
				t.Errorf("len(obj) = %d, want %d — output: %s", len(obj), c.wantObjLen, out)
			}
			// fs is non-nil for all current cases; this guard is defensive
			// against future rows that use a spec like "none" (nil selector).
			if fs != nil {
				for key := range obj {
					if !keyInSpec(key, fs.fields) {
						t.Errorf("output key %q was not requested in spec %q", key, c.spec)
					}
				}
			}
		})
	}
}

// Property: every key in each object of a filtered array was requested in the spec.
func TestFilter_ArrayOutputIsSubset(t *testing.T) {
	cases := []struct {
		name        string
		spec        string
		input       string
		wantArrLen  int
		wantElemLen int // expected number of top-level keys in each filtered element
	}{
		{"array simple", "id,title", `[{"id":"1","title":"A","extra":"x"},{"id":"2","title":"B","extra":"y"}]`, 2, 2},
		{"array only id", "id", `[{"id":"1","title":"T"},{"id":"2","title":"U"}]`, 2, 1},
		// "nothing matches" produces [{},{}]; wantElemLen=0 makes the subset invariant non-vacuous.
		{"array nothing matches", "missing", `[{"id":"1"},{"id":"2"}]`, 2, 0},
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
			var arr []map[string]any
			if err := json.Unmarshal(out, &arr); err != nil {
				t.Fatalf("output is not a valid JSON array of objects: %v — output: %s", err, out)
			}
			if len(arr) != c.wantArrLen {
				t.Fatalf("len(arr) = %d, want %d — output: %s", len(arr), c.wantArrLen, out)
			}
			for i, obj := range arr {
				if len(obj) != c.wantElemLen {
					t.Errorf("[%d] len(obj) = %d, want %d — obj: %v", i, len(obj), c.wantElemLen, obj)
				}
				for key := range obj {
					if !keyInSpec(key, fs.fields) {
						t.Errorf("[%d] key %q was not requested in spec %q", i, key, c.spec)
					}
				}
			}
		})
	}
}

// Property: nested sub-objects inside array elements are also filtered — unrequested keys
// within a nested object must not leak through.
func TestFilter_ArrayNestedObjectSubset(t *testing.T) {
	fs, err := New("assignee.name", nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	input := `[{"id":"1","assignee":{"name":"Alice","email":"x@example.com"},"extra":"drop"}]`
	out, err := fs.Filter([]byte(input))
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	var arr []map[string]any
	if err := json.Unmarshal(out, &arr); err != nil {
		t.Fatalf("output not valid JSON array: %v — output: %s", err, out)
	}
	if len(arr) != 1 {
		t.Fatalf("len(arr) = %d, want 1 — output: %s", len(arr), out)
	}
	elem := arr[0]
	if _, ok := elem["id"]; ok {
		t.Error("top-level 'id' should have been filtered out")
	}
	if _, ok := elem["extra"]; ok {
		t.Error("top-level 'extra' should have been filtered out")
	}
	assignee, ok := elem["assignee"].(map[string]any)
	if !ok {
		t.Fatalf("'assignee' missing or not an object — output: %s", out)
	}
	if _, ok := assignee["name"]; !ok {
		t.Error("assignee.name should be present")
	}
	if _, ok := assignee["email"]; ok {
		t.Error("assignee.email should have been filtered out")
	}
}

// Property: deeply nested spec (a.b.c) filters recursively through all levels.
func TestFilter_DeepNested(t *testing.T) {
	fs, err := New("a.b.c", nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	input := `{"a":{"b":{"c":"keep","d":"drop"},"e":"drop"},"f":"drop"}`
	out, err := fs.Filter([]byte(input))
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	a, ok := result["a"].(map[string]any)
	if !ok {
		t.Fatalf("output missing key 'a' or not an object — output: %s", out)
	}
	b, ok := a["b"].(map[string]any)
	if !ok {
		t.Fatalf("output missing key 'a.b' or not an object — output: %s", out)
	}
	if _, ok := b["c"]; !ok {
		t.Errorf("output missing 'a.b.c' — output: %s", out)
	}
	if _, ok := b["d"]; ok {
		t.Error("output contains 'a.b.d', should have been filtered out")
	}
	if _, ok := a["e"]; ok {
		t.Error("output contains 'a.e', should have been filtered out")
	}
	if _, ok := result["f"]; ok {
		t.Error("output contains 'f', should have been filtered out")
	}
}

// Property: deeply nested spec (a.b.c) filters recursively even when an intermediate
// key holds an array — filterObject recurses into each array element via filterValue.
func TestFilter_DeepNestedArray(t *testing.T) {
	fs, err := New("a.b.c", nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// "a" is an array; each element has "b.c" to keep plus "b.d" and "e" to drop.
	input := `{"a":[{"b":{"c":"keep","d":"drop"},"e":"drop"},{"b":{"c":"also-keep","d":"also-drop"},"e":"also-drop"}]}`
	out, err := fs.Filter([]byte(input))
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("output not valid JSON: %v — output: %s", err, out)
	}
	aRaw, ok := result["a"].([]any)
	if !ok {
		t.Fatalf("'a' missing or not an array — output: %s", out)
	}
	if len(aRaw) != 2 {
		t.Fatalf("len(a) = %d, want 2 — output: %s", len(aRaw), out)
	}
	for i, elem := range aRaw {
		obj, ok := elem.(map[string]any)
		if !ok {
			t.Fatalf("a[%d] is not an object", i)
		}
		if _, ok := obj["e"]; ok {
			t.Errorf("a[%d].e should have been filtered out", i)
		}
		b, ok := obj["b"].(map[string]any)
		if !ok {
			t.Fatalf("a[%d].b missing or not an object", i)
		}
		if _, ok := b["c"]; !ok {
			t.Errorf("a[%d].b.c should be present", i)
		}
		if _, ok := b["d"]; ok {
			t.Errorf("a[%d].b.d should have been filtered out", i)
		}
	}
}

// Property: output is always valid JSON, even when no fields match.
func TestFilter_AlwaysValidJSON(t *testing.T) {
	cases := []struct {
		spec       string
		input      string
		wantOutput string // if non-empty, exact expected output bytes
	}{
		{"id", `{"id":"1","title":"T"}`, ""},
		{"missing", `{"id":"1","title":"T"}`, ""},
		{"id,title", `[{"id":"1"},{"id":"2"}]`, ""},
		{"x", `{}`, ""},
		{"x", `[]`, ""},
		// JSON null input: passes through unchanged and re-encodes as "null".
		{"x", `null`, "null"},
		// Top-level JSON number and boolean: filterValue default branch passes through.
		{"x", `42`, "42"},
		{"x", `true`, "true"},
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
		if c.wantOutput != "" && string(out) != c.wantOutput {
			t.Errorf("Filter(%q, %q) = %s, want %s", c.spec, c.input, out, c.wantOutput)
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
	f.Add("a.b.c", `{"a":{"b":{"c":"v","d":"drop"},"e":"drop"}}`) // two-level recursive filterObject path

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
// Assumes JSON keys do not contain literal dots; a key like "a.b" would match a spec
// "a.b.c" via the prefix check, mirroring the same assumption in filterObject.
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
