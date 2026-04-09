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

func TestNullable_UnmarshalJSON(t *testing.T) {
	// Unmarshal value
	t.Run("unmarshal value", func(t *testing.T) {
		var n Nullable[string]
		if err := json.Unmarshal([]byte(`"test"`), &n); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if !n.IsSet() {
			t.Error("Expected IsSet() to be true after unmarshal")
		}
		val, ok := n.Get()
		if !ok || val == nil || *val != "test" {
			t.Error("Get() should return test value")
		}
	})

	// Unmarshal null
	t.Run("unmarshal null", func(t *testing.T) {
		var n Nullable[string]
		if err := json.Unmarshal([]byte(`null`), &n); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if !n.IsSet() {
			t.Error("Expected IsSet() to be true after unmarshal null")
		}
		val, ok := n.Get()
		if !ok {
			t.Error("Get() should return ok=true for null")
		}
		if val != nil {
			t.Error("Get() should return nil value for null")
		}
	})

	// Unmarshal int
	t.Run("unmarshal int", func(t *testing.T) {
		var n Nullable[int]
		if err := json.Unmarshal([]byte(`42`), &n); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		val, ok := n.Get()
		if !ok || val == nil || *val != 42 {
			t.Error("Get() should return 42")
		}
	})
}

func TestNullable_IsZero(t *testing.T) {
	// Unset should be zero
	t.Run("unset is zero", func(t *testing.T) {
		n := NewUnset[string]()
		if !n.IsZero() {
			t.Error("Unset should be zero")
		}
	})

	// Value should not be zero
	t.Run("value is not zero", func(t *testing.T) {
		n := NewValue("test")
		if n.IsZero() {
			t.Error("Value should not be zero")
		}
	})

	// Null should not be zero (it's explicitly set)
	t.Run("null is not zero", func(t *testing.T) {
		n := NewNull[string]()
		if n.IsZero() {
			t.Error("Null should not be zero (it's explicitly set)")
		}
	})
}

func TestIssueUpdateNullableInput_ToMap(t *testing.T) {
	// Empty input
	t.Run("empty input", func(t *testing.T) {
		input := IssueUpdateNullableInput{}
		m := input.ToMap()
		if len(m) != 0 {
			t.Errorf("Empty input should produce empty map, got %v", m)
		}
	})

	// With title
	t.Run("with title", func(t *testing.T) {
		title := "Test Title"
		input := IssueUpdateNullableInput{Title: &title}
		m := input.ToMap()
		if m["title"] != title {
			t.Errorf("Expected title=%s, got %v", title, m["title"])
		}
	})

	// With null parent (explicit removal)
	t.Run("with null parent", func(t *testing.T) {
		input := IssueUpdateNullableInput{
			ParentID: NewNull[string](),
		}
		m := input.ToMap()
		if _, exists := m["parentId"]; !exists {
			t.Error("Expected parentId to be in map")
		}
		if m["parentId"] != nil {
			t.Errorf("Expected parentId=nil, got %v", m["parentId"])
		}
	})

	// With value parent
	t.Run("with value parent", func(t *testing.T) {
		input := IssueUpdateNullableInput{
			ParentID: NewValue("parent-123"),
		}
		m := input.ToMap()
		if m["parentId"] != "parent-123" {
			t.Errorf("Expected parentId=parent-123, got %v", m["parentId"])
		}
	})

	// With unset parent (should not be in map)
	t.Run("with unset parent", func(t *testing.T) {
		input := IssueUpdateNullableInput{
			ParentID: NewUnset[string](),
		}
		m := input.ToMap()
		if _, exists := m["parentId"]; exists {
			t.Error("Expected parentId to NOT be in map for unset")
		}
	})

	// With null cycle
	t.Run("with null cycle", func(t *testing.T) {
		input := IssueUpdateNullableInput{
			CycleID: NewNull[string](),
		}
		m := input.ToMap()
		if _, exists := m["cycleId"]; !exists {
			t.Error("Expected cycleId to be in map")
		}
		if m["cycleId"] != nil {
			t.Errorf("Expected cycleId=nil, got %v", m["cycleId"])
		}
	})

	// With null project
	t.Run("with null project", func(t *testing.T) {
		input := IssueUpdateNullableInput{
			ProjectID: NewNull[string](),
		}
		m := input.ToMap()
		if _, exists := m["projectId"]; !exists {
			t.Error("Expected projectId to be in map")
		}
		if m["projectId"] != nil {
			t.Errorf("Expected projectId=nil, got %v", m["projectId"])
		}
	})

	// With null assignee (explicit unassign)
	t.Run("with null assignee", func(t *testing.T) {
		input := IssueUpdateNullableInput{
			AssigneeID: NewNull[string](),
		}
		m := input.ToMap()
		if _, exists := m["assigneeId"]; !exists {
			t.Error("Expected assigneeId to be in map")
		}
		if m["assigneeId"] != nil {
			t.Errorf("Expected assigneeId=nil, got %v", m["assigneeId"])
		}
	})

	// With value assignee
	t.Run("with value assignee", func(t *testing.T) {
		input := IssueUpdateNullableInput{
			AssigneeID: NewValue("user-123"),
		}
		m := input.ToMap()
		if m["assigneeId"] != "user-123" {
			t.Errorf("Expected assigneeId=user-123, got %v", m["assigneeId"])
		}
	})

	// With value estimate
	t.Run("with value estimate", func(t *testing.T) {
		input := IssueUpdateNullableInput{
			Estimate: NewValue(int64(5)),
		}
		m := input.ToMap()
		if m["estimate"] != int64(5) {
			t.Errorf("Expected estimate=5, got %v", m["estimate"])
		}
	})

	// With null estimate (explicit clear)
	t.Run("with null estimate", func(t *testing.T) {
		input := IssueUpdateNullableInput{
			Estimate: NewNull[int64](),
		}
		m := input.ToMap()
		if _, exists := m["estimate"]; !exists {
			t.Error("Expected estimate to be in map")
		}
		if m["estimate"] != nil {
			t.Errorf("Expected estimate=nil, got %v", m["estimate"])
		}
	})

	// With unset estimate (should not appear in map)
	t.Run("with unset estimate", func(t *testing.T) {
		input := IssueUpdateNullableInput{
			Estimate: NewUnset[int64](),
		}
		m := input.ToMap()
		if _, exists := m["estimate"]; exists {
			t.Error("Expected estimate to NOT be in map for unset")
		}
	})

	// All fields
	t.Run("all fields", func(t *testing.T) {
		title := "Title"
		desc := "Desc"
		state := "state-id"
		priority := int64(1)
		estimate := int64(3)
		input := IssueUpdateNullableInput{
			Title:           &title,
			Description:     &desc,
			AssigneeID:      NewValue("assignee-id"),
			StateID:         &state,
			Priority:        &priority,
			Estimate:        NewValue(estimate),
			AddedLabelIds:   []string{"label-1"},
			RemovedLabelIds: []string{"label-2"},
			CycleID:         NewValue("cycle-id"),
			ParentID:        NewNull[string](),
			ProjectID:       NewUnset[string](),
		}
		m := input.ToMap()

		if m["title"] != title {
			t.Errorf("Expected title=%s", title)
		}
		if m["description"] != desc {
			t.Errorf("Expected description=%s", desc)
		}
		if m["assigneeId"] != "assignee-id" {
			t.Errorf("Expected assigneeId=assignee-id, got %v", m["assigneeId"])
		}
		if m["stateId"] != state {
			t.Errorf("Expected stateId=%s", state)
		}
		if m["priority"] != priority {
			t.Errorf("Expected priority=%d", priority)
		}
		if m["estimate"] != estimate {
			t.Errorf("Expected estimate=%d, got %v", estimate, m["estimate"])
		}
		if m["cycleId"] != "cycle-id" {
			t.Errorf("Expected cycleId=cycle-id")
		}
		if m["parentId"] != nil {
			t.Errorf("Expected parentId=nil (explicit null)")
		}
		if _, exists := m["projectId"]; exists {
			t.Error("projectId should not be in map (unset)")
		}
	})
}
