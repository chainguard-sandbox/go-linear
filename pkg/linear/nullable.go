package linear

import "encoding/json"

// Nullable represents an optional GraphQL field that can be unset, null, or have a value.
// Solves the omitempty limitation where we can't send explicit null.
//
// States:
//   - Unset: Field omitted from JSON (IsSet=false)
//   - Null: Field included as null (IsSet=true, Value=nil)
//   - Value: Field included with value (IsSet=true, Value=&T)
//
// Usage in structs:
//
//	Use *Nullable[T] with omitempty:
//	- nil pointer = unset (omitted)
//	- &NewNull() = explicit null
//	- &NewValue(x) = value
type Nullable[T any] struct {
	value *T
	isSet bool
}

// NewValue creates a Nullable with a value.
func NewValue[T any](value T) Nullable[T] {
	return Nullable[T]{value: &value, isSet: true}
}

// NewNull creates a Nullable explicitly set to null.
func NewNull[T any]() Nullable[T] {
	return Nullable[T]{value: nil, isSet: true}
}

// NewUnset creates an unset Nullable (field will be omitted).
func NewUnset[T any]() Nullable[T] {
	return Nullable[T]{isSet: false}
}

// IsSet returns true if the field was explicitly set (even to null).
func (n Nullable[T]) IsSet() bool {
	return n.isSet
}

// Get returns the value and whether it's set.
func (n Nullable[T]) Get() (*T, bool) {
	if !n.isSet {
		return nil, false
	}
	return n.value, true
}

// MarshalJSON implements json.Marshaler.
func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if !n.isSet {
		// This shouldn't happen with pointer approach, but return null as fallback
		return []byte("null"), nil
	}
	if n.value == nil {
		// Explicitly set to null
		return []byte("null"), nil
	}
	// Has a value
	return json.Marshal(*n.value)
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	n.isSet = true
	if string(data) == "null" {
		n.value = nil
		return nil
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	n.value = &value
	return nil
}

// IsZero reports whether Nullable is unset for omitempty behavior.
// Called by encoding/json when omitempty is used.
// Returns true when unset (field omitted), false when set (field included).
func (n Nullable[T]) IsZero() bool {
	return !n.isSet
}
