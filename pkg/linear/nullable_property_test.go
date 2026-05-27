package linear

import (
	"encoding/json"
	"testing"
	"unicode/utf8"
)

// Property: marshal → unmarshal is the identity for all three states.
func TestNullable_RoundTrip(t *testing.T) {
	t.Run("string values", func(t *testing.T) {
		for _, s := range []string{"", "hello", `"quoted"`, "with\nnewline", "unicode: 日本語"} {
			original := NewValue(s)
			data, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("Marshal(%q): %v", s, err)
			}
			var decoded Nullable[string]
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal(%q): %v", s, err)
			}
			v, ok := decoded.Get()
			if !ok || v == nil || *v != s {
				t.Errorf("RoundTrip(%q): got (%v, %v)", s, v, ok)
			}
		}
	})

	t.Run("null", func(t *testing.T) {
		original := NewNull[string]()
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		var decoded Nullable[string]
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		v, ok := decoded.Get()
		if !ok || v != nil {
			t.Errorf("RoundTrip(null): Get() = (%v, %v), want (nil, true)", v, ok)
		}
	})

	t.Run("int values", func(t *testing.T) {
		for _, n := range []int{0, 1, -1, 42, 1<<31 - 1, -(1 << 31)} {
			original := NewValue(n)
			data, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("RoundTrip(%d) Marshal: %v", n, err)
			}
			var decoded Nullable[int]
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("RoundTrip(%d) Unmarshal: %v", n, err)
			}
			v, ok := decoded.Get()
			if !ok || v == nil || *v != n {
				t.Errorf("RoundTrip(%d): got (%v, %v)", n, v, ok)
			}
		}
	})
}

// UnmarshalJSON returns an error when JSON type doesn't match T.
func TestNullable_UnmarshalTypeError(t *testing.T) {
	var n Nullable[int]
	if err := json.Unmarshal([]byte(`"not-a-number"`), &n); err == nil {
		t.Error("expected error unmarshaling string into Nullable[int]")
	}
}

// Property: the three states are mutually exclusive.
// For each state: IsSet, IsZero, and Get() must all agree.
func TestNullable_StateInvariant(t *testing.T) {
	cases := []struct {
		name  string
		n     Nullable[string]
		isSet bool
		isNil bool // whether Get() returns a nil pointer
	}{
		{"unset", NewUnset[string](), false, true},
		{"null", NewNull[string](), true, true},
		{"value", NewValue("x"), true, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.n.IsSet(); got != c.isSet {
				t.Errorf("IsSet() = %v, want %v", got, c.isSet)
			}
			// IsZero is the inverse of IsSet for omitempty
			if got := c.n.IsZero(); got != !c.isSet {
				t.Errorf("IsZero() = %v, want %v", got, !c.isSet)
			}
			v, ok := c.n.Get()
			if ok != c.isSet {
				t.Errorf("Get() ok = %v, want %v", ok, c.isSet)
			}
			if (v == nil) != c.isNil {
				t.Errorf("Get() v==nil is %v, want %v", v == nil, c.isNil)
			}
		})
	}
}

// Fuzz: marshal→unmarshal round-trip for valid UTF-8 strings never panics and preserves value.
// JSON requires valid UTF-8; encoding/json replaces invalid bytes with U+FFFD, so the
// round-trip invariant is only guaranteed for valid UTF-8 strings.
func FuzzNullable_RoundTrip(f *testing.F) {
	f.Add("hello")
	f.Add("")
	f.Add(`"quoted"`)
	f.Add("null")
	f.Add("unicode: 日本語")

	f.Fuzz(func(t *testing.T, s string) {
		if !utf8.ValidString(s) {
			return // JSON round-trip only preserves valid UTF-8
		}
		original := NewValue(s)
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var decoded Nullable[string]
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		v, ok := decoded.Get()
		if !ok || v == nil || *v != s {
			t.Errorf("RoundTrip(%q): got (%v, %v)", s, v, ok)
		}
	})
}
