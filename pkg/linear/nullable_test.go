package linear

import (
	"encoding/json"
	"testing"
)

func TestNullable_MarshalJSON(t *testing.T) {
	// Test with value
	t.Run("with value", func(t *testing.T) {
		n := NewValue("test")
		data, err := json.Marshal(n)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		if string(data) != `"test"` {
			t.Errorf("Expected \"test\", got %s", string(data))
		}
	})

	// Test with null
	t.Run("with null", func(t *testing.T) {
		n := NewNull[string]()
		data, err := json.Marshal(n)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		if string(data) != "null" {
			t.Errorf("Expected null, got %s", string(data))
		}
	})

	// Test unset
	t.Run("unset", func(t *testing.T) {
		n := NewUnset[string]()
		if n.IsSet() {
			t.Error("Unset should return false for IsSet()")
		}
	})
}

func TestNullable_InStruct(t *testing.T) {
	type TestStruct struct {
		Field *Nullable[string] `json:"field,omitempty"`
	}

	// With value
	t.Run("value included", func(t *testing.T) {
		v := NewValue("test")
		s := TestStruct{Field: &v}
		data, _ := json.Marshal(s)
		if string(data) != `{"field":"test"}` {
			t.Errorf("Expected {\"field\":\"test\"}, got %s", string(data))
		}
	})

	// With null - should be included
	t.Run("null included", func(t *testing.T) {
		v := NewNull[string]()
		s := TestStruct{Field: &v}
		data, _ := json.Marshal(s)
		if string(data) != `{"field":null}` {
			t.Errorf("Expected {\"field\":null}, got %s", string(data))
		}
	})

	// Unset - use nil pointer, omitted by omitempty
	t.Run("unset omitted", func(t *testing.T) {
		s := TestStruct{Field: nil}
		data, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		expected := "{}"
		if string(data) != expected {
			t.Errorf("Expected %s, got %s", expected, string(data))
		}
	})
}

func TestNullable_Get(t *testing.T) {
	// With value
	n := NewValue("test")
	val, ok := n.Get()
	if !ok || val == nil || *val != "test" {
		t.Error("Get() should return value and true")
	}

	// With null
	n2 := NewNull[string]()
	val, ok = n2.Get()
	if !ok || val != nil {
		t.Error("Get() for null should return nil and true")
	}

	// Unset
	n3 := NewUnset[string]()
	_, ok = n3.Get()
	if ok {
		t.Error("Get() for unset should return false")
	}
}
