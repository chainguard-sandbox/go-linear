package cycle

import (
	"context"
	"testing"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/dateparser"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// testBuilder creates a FilterBuilder with a nil resolver for unit tests.
func testBuilder() *FilterBuilder {
	return &FilterBuilder{
		filter: &intgraphql.CycleFilter{},
		parser: dateparser.New(),
	}
}

// mustSet sets a flag value, panicking on error.
func mustSet(cmd *cobra.Command, name, value string) {
	if err := cmd.Flags().Set(name, value); err != nil {
		panic(err)
	}
}

// testCommand creates a cobra command with the specified flags for testing.
func testCommand(flags map[string]any) *cobra.Command {
	cmd := &cobra.Command{Use: "test"}

	for name, value := range flags {
		switch v := value.(type) {
		case string:
			cmd.Flags().String(name, "", "")
			if v != "" {
				mustSet(cmd, name, v)
			}
		case int:
			cmd.Flags().Int(name, -1, "")
			if v >= 0 {
				mustSet(cmd, name, itoa(v))
			}
		case bool:
			cmd.Flags().Bool(name, false, "")
			if v {
				mustSet(cmd, name, "true")
			}
		}
	}

	return cmd
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + itoa(-i)
	}
	s := ""
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}

// =============================================================================
// Builder Tests
// =============================================================================

func TestFilterBuilder_Build_Empty(t *testing.T) {
	b := testBuilder()
	filter := b.Build()
	if filter != nil {
		t.Errorf("Build() with no filters = %v, want nil", filter)
	}
}

func TestFilterBuilder_Resolver(t *testing.T) {
	b := NewFilterBuilder(nil)
	if b.Resolver() != nil {
		t.Error("Resolver() with nil resolver should return nil")
	}
}

func TestFilterBuilder_Parser(t *testing.T) {
	b := testBuilder()
	if b.Parser() == nil {
		t.Error("Parser() should return non-nil parser")
	}
}

func TestFilterBuilder_Filter(t *testing.T) {
	b := testBuilder()
	if b.Filter() == nil {
		t.Error("Filter() should return non-nil filter")
	}
}

// =============================================================================
// Date Filter Tests
// =============================================================================

func TestApplyCreatedAt(t *testing.T) {
	tests := []struct {
		name       string
		after      string
		before     string
		wantFilter bool
		wantErr    bool
	}{
		{
			name:       "no flags",
			wantFilter: false,
		},
		{
			name:       "after only",
			after:      "2025-01-01",
			wantFilter: true,
		},
		{
			name:       "before only",
			before:     "2025-12-31",
			wantFilter: true,
		},
		{
			name:       "both flags",
			after:      "2025-01-01",
			before:     "2025-12-31",
			wantFilter: true,
		},
		{
			name:    "invalid after date",
			after:   "invalid",
			wantErr: true,
		},
		{
			name:    "invalid before date",
			before:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBuilder()
			cmd := testCommand(map[string]any{
				"created-after":  tt.after,
				"created-before": tt.before,
			})

			err := ApplyCreatedAt(context.Background(), cmd, b)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyCreatedAt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Skip filter check on error cases
			if tt.wantErr {
				return
			}

			hasFilter := b.filter.CreatedAt != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyCreatedAt() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplyUpdatedAt(t *testing.T) {
	b := testBuilder()
	cmd := testCommand(map[string]any{
		"updated-after":  "2025-01-01",
		"updated-before": "2025-12-31",
	})

	err := ApplyUpdatedAt(context.Background(), cmd, b)
	if err != nil {
		t.Fatalf("ApplyUpdatedAt() error = %v", err)
	}

	if b.filter.UpdatedAt == nil {
		t.Error("UpdatedAt filter not set")
	}
}

func TestApplyCompletedAt(t *testing.T) {
	b := testBuilder()
	cmd := testCommand(map[string]any{
		"completed-after":  "2025-01-01",
		"completed-before": "2025-12-31",
	})

	err := ApplyCompletedAt(context.Background(), cmd, b)
	if err != nil {
		t.Fatalf("ApplyCompletedAt() error = %v", err)
	}

	if b.filter.CompletedAt == nil {
		t.Error("CompletedAt filter not set")
	}
}

func TestApplyStartsAt(t *testing.T) {
	b := testBuilder()
	cmd := testCommand(map[string]any{
		"starts-after":  "2025-01-01",
		"starts-before": "2025-03-31",
	})

	err := ApplyStartsAt(context.Background(), cmd, b)
	if err != nil {
		t.Fatalf("ApplyStartsAt() error = %v", err)
	}

	if b.filter.StartsAt == nil {
		t.Error("StartsAt filter not set")
	}
}

func TestApplyEndsAt(t *testing.T) {
	b := testBuilder()
	cmd := testCommand(map[string]any{
		"ends-after":  "2025-01-01",
		"ends-before": "2025-03-31",
	})

	err := ApplyEndsAt(context.Background(), cmd, b)
	if err != nil {
		t.Fatalf("ApplyEndsAt() error = %v", err)
	}

	if b.filter.EndsAt == nil {
		t.Error("EndsAt filter not set")
	}
}

// =============================================================================
// Boolean Filter Tests
// =============================================================================

func TestApplyIsActive(t *testing.T) {
	tests := []struct {
		name       string
		active     bool
		wantFilter bool
	}{
		{name: "not set", active: false, wantFilter: false},
		{name: "set", active: true, wantFilter: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBuilder()
			cmd := testCommand(map[string]any{"active": tt.active})

			err := ApplyIsActive(context.Background(), cmd, b)
			if err != nil {
				t.Fatalf("ApplyIsActive() error = %v", err)
			}

			hasFilter := b.filter.IsActive != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyIsActive() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplyIsFuture(t *testing.T) {
	b := testBuilder()
	cmd := testCommand(map[string]any{"future": true})

	err := ApplyIsFuture(context.Background(), cmd, b)
	if err != nil {
		t.Fatalf("ApplyIsFuture() error = %v", err)
	}

	if b.filter.IsFuture == nil {
		t.Error("IsFuture filter not set")
	}
}

func TestApplyIsPast(t *testing.T) {
	b := testBuilder()
	cmd := testCommand(map[string]any{"past": true})

	err := ApplyIsPast(context.Background(), cmd, b)
	if err != nil {
		t.Fatalf("ApplyIsPast() error = %v", err)
	}

	if b.filter.IsPast == nil {
		t.Error("IsPast filter not set")
	}
}

func TestApplyIsNext(t *testing.T) {
	b := testBuilder()
	cmd := testCommand(map[string]any{"next": true})

	err := ApplyIsNext(context.Background(), cmd, b)
	if err != nil {
		t.Fatalf("ApplyIsNext() error = %v", err)
	}

	if b.filter.IsNext == nil {
		t.Error("IsNext filter not set")
	}
}

func TestApplyIsPrevious(t *testing.T) {
	b := testBuilder()
	cmd := testCommand(map[string]any{"previous": true})

	err := ApplyIsPrevious(context.Background(), cmd, b)
	if err != nil {
		t.Fatalf("ApplyIsPrevious() error = %v", err)
	}

	if b.filter.IsPrevious == nil {
		t.Error("IsPrevious filter not set")
	}
}

func TestApplyIsInCooldown(t *testing.T) {
	b := testBuilder()
	cmd := testCommand(map[string]any{"in-cooldown": true})

	err := ApplyIsInCooldown(context.Background(), cmd, b)
	if err != nil {
		t.Fatalf("ApplyIsInCooldown() error = %v", err)
	}

	if b.filter.IsInCooldown == nil {
		t.Error("IsInCooldown filter not set")
	}
}

// =============================================================================
// Entity Filter Tests
// =============================================================================

func TestApplyID(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		wantFilter bool
	}{
		{name: "empty", id: "", wantFilter: false},
		{name: "set", id: "test-uuid", wantFilter: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBuilder()
			cmd := testCommand(map[string]any{"id": tt.id})

			err := ApplyID(context.Background(), cmd, b)
			if err != nil {
				t.Fatalf("ApplyID() error = %v", err)
			}

			hasFilter := b.filter.ID != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyID() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}

			if tt.wantFilter && b.filter.ID.Eq != nil && *b.filter.ID.Eq != tt.id {
				t.Errorf("ApplyID() ID = %v, want %v", *b.filter.ID.Eq, tt.id)
			}
		})
	}
}

func TestApplyName(t *testing.T) {
	tests := []struct {
		name       string
		cycleName  string
		wantFilter bool
	}{
		{name: "empty", cycleName: "", wantFilter: false},
		{name: "set", cycleName: "Sprint 1", wantFilter: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBuilder()
			cmd := testCommand(map[string]any{"name": tt.cycleName})

			err := ApplyName(context.Background(), cmd, b)
			if err != nil {
				t.Fatalf("ApplyName() error = %v", err)
			}

			hasFilter := b.filter.Name != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyName() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}

			if tt.wantFilter && b.filter.Name.ContainsIgnoreCase != nil && *b.filter.Name.ContainsIgnoreCase != tt.cycleName {
				t.Errorf("ApplyName() Name = %v, want %v", *b.filter.Name.ContainsIgnoreCase, tt.cycleName)
			}
		})
	}
}

func TestApplyNumber(t *testing.T) {
	tests := []struct {
		name       string
		number     int
		wantFilter bool
	}{
		{name: "not set", number: -1, wantFilter: false},
		{name: "zero", number: 0, wantFilter: true},
		{name: "positive", number: 42, wantFilter: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBuilder()
			cmd := testCommand(map[string]any{"number": tt.number})

			err := ApplyNumber(context.Background(), cmd, b)
			if err != nil {
				t.Fatalf("ApplyNumber() error = %v", err)
			}

			hasFilter := b.filter.Number != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyNumber() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}

			if tt.wantFilter && b.filter.Number.Eq != nil && *b.filter.Number.Eq != float64(tt.number) {
				t.Errorf("ApplyNumber() Number = %v, want %v", *b.filter.Number.Eq, float64(tt.number))
			}
		})
	}
}

// =============================================================================
// Registry Tests
// =============================================================================

func TestFilters_Registry(t *testing.T) {
	expectedFilters := []string{
		"created-at", "updated-at", "completed-at", "starts-at", "ends-at",
		"is-active", "is-future", "is-past", "is-next", "is-previous", "is-in-cooldown",
		"id", "name", "number", "team",
	}

	for _, name := range expectedFilters {
		if _, ok := Filters[name]; !ok {
			t.Errorf("Filters missing expected filter: %s", name)
		}
	}

	if len(Filters) != len(expectedFilters) {
		t.Errorf("Filters count = %d, want %d", len(Filters), len(expectedFilters))
	}
}

// =============================================================================
// Comparator Accessor Tests
// =============================================================================

func TestFilterBuilder_CreatedAtComparator(t *testing.T) {
	b := testBuilder()
	comp1 := b.CreatedAtComparator()
	comp2 := b.CreatedAtComparator()

	if comp1 != comp2 {
		t.Error("CreatedAtComparator() should return same instance")
	}
}

func TestFilterBuilder_UpdatedAtComparator(t *testing.T) {
	b := testBuilder()
	comp1 := b.UpdatedAtComparator()
	comp2 := b.UpdatedAtComparator()

	if comp1 != comp2 {
		t.Error("UpdatedAtComparator() should return same instance")
	}
}

func TestFilterBuilder_CompletedAtComparator(t *testing.T) {
	b := testBuilder()
	comp1 := b.CompletedAtComparator()
	comp2 := b.CompletedAtComparator()

	if comp1 != comp2 {
		t.Error("CompletedAtComparator() should return same instance")
	}
}

func TestFilterBuilder_StartsAtComparator(t *testing.T) {
	b := testBuilder()
	comp1 := b.StartsAtComparator()
	comp2 := b.StartsAtComparator()

	if comp1 != comp2 {
		t.Error("StartsAtComparator() should return same instance")
	}
}

func TestFilterBuilder_EndsAtComparator(t *testing.T) {
	b := testBuilder()
	comp1 := b.EndsAtComparator()
	comp2 := b.EndsAtComparator()

	if comp1 != comp2 {
		t.Error("EndsAtComparator() should return same instance")
	}
}

// =============================================================================
// Build Tests with Filters
// =============================================================================

func TestFilterBuilder_Build_WithFilters(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*FilterBuilder)
		wantNil bool
	}{
		{
			name:    "empty",
			setup:   func(b *FilterBuilder) {},
			wantNil: true,
		},
		{
			name: "with ID",
			setup: func(b *FilterBuilder) {
				id := "test"
				b.filter.ID = &intgraphql.IDComparator{Eq: &id}
			},
			wantNil: false,
		},
		{
			name: "with IsActive",
			setup: func(b *FilterBuilder) {
				active := true
				b.filter.IsActive = &intgraphql.BooleanComparator{Eq: &active}
			},
			wantNil: false,
		},
		{
			name: "with CreatedAt",
			setup: func(b *FilterBuilder) {
				b.CreatedAtComparator()
				date := "2025-01-01T00:00:00.000Z"
				b.filter.CreatedAt.Gte = &date
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBuilder()
			tt.setup(b)

			filter := b.Build()
			gotNil := filter == nil
			if gotNil != tt.wantNil {
				t.Errorf("Build() gotNil = %v, wantNil %v", gotNil, tt.wantNil)
			}
		})
	}
}
