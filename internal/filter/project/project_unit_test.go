package project

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
		filter: &intgraphql.ProjectFilter{},
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
		case bool:
			cmd.Flags().Bool(name, false, "")
			if v {
				mustSet(cmd, name, "true")
			}
		case int:
			cmd.Flags().Int(name, -1, "")
			mustSet(cmd, name, "0") // Mark as changed
		}
	}

	return cmd
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
		{name: "no flags", wantFilter: false},
		{name: "after only", after: "2025-01-01", wantFilter: true},
		{name: "before only", before: "2025-12-31", wantFilter: true},
		{name: "both flags", after: "2025-01-01", before: "2025-12-31", wantFilter: true},
		{name: "invalid after date", after: "invalid", wantErr: true},
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
	tests := []struct {
		name       string
		after      string
		before     string
		wantFilter bool
		wantErr    bool
	}{
		{name: "no flags", wantFilter: false},
		{name: "after only", after: "2025-01-01", wantFilter: true},
		{name: "before only", before: "2025-12-31", wantFilter: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBuilder()
			cmd := testCommand(map[string]any{
				"updated-after":  tt.after,
				"updated-before": tt.before,
			})

			err := ApplyUpdatedAt(context.Background(), cmd, b)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyUpdatedAt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			hasFilter := b.filter.UpdatedAt != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyUpdatedAt() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplyCompletedAt(t *testing.T) {
	tests := []struct {
		name       string
		after      string
		before     string
		wantFilter bool
	}{
		{name: "no flags", wantFilter: false},
		{name: "after only", after: "2025-01-01", wantFilter: true},
		{name: "before only", before: "2025-12-31", wantFilter: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBuilder()
			cmd := testCommand(map[string]any{
				"completed-after":  tt.after,
				"completed-before": tt.before,
			})

			err := ApplyCompletedAt(context.Background(), cmd, b)
			if err != nil {
				t.Fatalf("ApplyCompletedAt() error = %v", err)
			}

			hasFilter := b.filter.CompletedAt != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyCompletedAt() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplyCanceledAt(t *testing.T) {
	tests := []struct {
		name       string
		after      string
		before     string
		wantFilter bool
	}{
		{name: "no flags", wantFilter: false},
		{name: "after only", after: "2025-01-01", wantFilter: true},
		{name: "before only", before: "2025-12-31", wantFilter: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBuilder()
			cmd := testCommand(map[string]any{
				"canceled-after":  tt.after,
				"canceled-before": tt.before,
			})

			err := ApplyCanceledAt(context.Background(), cmd, b)
			if err != nil {
				t.Fatalf("ApplyCanceledAt() error = %v", err)
			}

			hasFilter := b.filter.CanceledAt != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyCanceledAt() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplyStartedAt(t *testing.T) {
	tests := []struct {
		name       string
		after      string
		before     string
		wantFilter bool
	}{
		{name: "no flags", wantFilter: false},
		{name: "after only", after: "2025-01-01", wantFilter: true},
		{name: "before only", before: "2025-12-31", wantFilter: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBuilder()
			cmd := testCommand(map[string]any{
				"started-after":  tt.after,
				"started-before": tt.before,
			})

			err := ApplyStartedAt(context.Background(), cmd, b)
			if err != nil {
				t.Fatalf("ApplyStartedAt() error = %v", err)
			}

			hasFilter := b.filter.StartedAt != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyStartedAt() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplyTargetDate(t *testing.T) {
	b := testBuilder()
	cmd := testCommand(map[string]any{
		"target-after":  "2025-01-01",
		"target-before": "",
	})

	err := ApplyTargetDate(context.Background(), cmd, b)
	if err != nil {
		t.Fatalf("ApplyTargetDate() error = %v", err)
	}

	if b.filter.TargetDate == nil {
		t.Error("TargetDate should not be nil when after flag is set")
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
		})
	}
}

func TestApplyCreator_Empty(t *testing.T) {
	b := testBuilder()
	cmd := testCommand(map[string]any{"creator": ""})

	err := ApplyCreator(context.Background(), cmd, b)
	if err != nil {
		t.Fatalf("ApplyCreator() error = %v", err)
	}

	if b.filter.Creator != nil {
		t.Error("Creator filter should be nil when flag is empty")
	}
}

func TestApplyLead_Empty(t *testing.T) {
	b := testBuilder()
	cmd := testCommand(map[string]any{"lead": ""})

	err := ApplyLead(context.Background(), cmd, b)
	if err != nil {
		t.Fatalf("ApplyLead() error = %v", err)
	}

	if b.filter.Lead != nil {
		t.Error("Lead filter should be nil when flag is empty")
	}
}

// =============================================================================
// Text Filter Tests
// =============================================================================

func TestApplyName(t *testing.T) {
	tests := []struct {
		name       string
		projName   string
		wantFilter bool
	}{
		{name: "empty", projName: "", wantFilter: false},
		{name: "set", projName: "Q1 Launch", wantFilter: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBuilder()
			cmd := testCommand(map[string]any{"name": tt.projName})

			err := ApplyName(context.Background(), cmd, b)
			if err != nil {
				t.Fatalf("ApplyName() error = %v", err)
			}

			hasFilter := b.filter.Name != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyName() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplySlugID(t *testing.T) {
	tests := []struct {
		name       string
		slugID     string
		wantFilter bool
	}{
		{name: "empty", slugID: "", wantFilter: false},
		{name: "set", slugID: "q1-launch", wantFilter: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBuilder()
			cmd := testCommand(map[string]any{"slug-id": tt.slugID})

			err := ApplySlugID(context.Background(), cmd, b)
			if err != nil {
				t.Fatalf("ApplySlugID() error = %v", err)
			}

			hasFilter := b.filter.SlugID != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplySlugID() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

// =============================================================================
// State Filter Tests
// =============================================================================

func TestApplyHealth(t *testing.T) {
	tests := []struct {
		name       string
		health     string
		wantFilter bool
	}{
		{name: "empty", health: "", wantFilter: false},
		{name: "onTrack", health: "onTrack", wantFilter: true},
		{name: "atRisk", health: "atRisk", wantFilter: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBuilder()
			cmd := testCommand(map[string]any{"health": tt.health})

			err := ApplyHealth(context.Background(), cmd, b)
			if err != nil {
				t.Fatalf("ApplyHealth() error = %v", err)
			}

			hasFilter := b.filter.Health != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyHealth() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

// =============================================================================
// Relation Filter Tests
// =============================================================================

func TestApplyHasBlockedBy(t *testing.T) {
	tests := []struct {
		name         string
		hasBlockedBy bool
		wantFilter   bool
	}{
		{name: "false", hasBlockedBy: false, wantFilter: false},
		{name: "true", hasBlockedBy: true, wantFilter: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBuilder()
			cmd := testCommand(map[string]any{"has-blocked-by": tt.hasBlockedBy})

			err := ApplyHasBlockedBy(context.Background(), cmd, b)
			if err != nil {
				t.Fatalf("ApplyHasBlockedBy() error = %v", err)
			}

			hasFilter := b.filter.HasBlockedByRelations != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyHasBlockedBy() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplyHasBlocking(t *testing.T) {
	b := testBuilder()
	cmd := testCommand(map[string]any{"has-blocking": false})

	err := ApplyHasBlocking(context.Background(), cmd, b)
	if err != nil {
		t.Fatalf("ApplyHasBlocking() error = %v", err)
	}

	if b.filter.HasBlockingRelations != nil {
		t.Error("HasBlockingRelations should be nil when flag is false")
	}
}

func TestApplyHasRelated(t *testing.T) {
	b := testBuilder()
	cmd := testCommand(map[string]any{"has-related": false})

	err := ApplyHasRelated(context.Background(), cmd, b)
	if err != nil {
		t.Fatalf("ApplyHasRelated() error = %v", err)
	}

	if b.filter.HasRelatedRelations != nil {
		t.Error("HasRelatedRelations should be nil when flag is false")
	}
}

// =============================================================================
// Registry Tests
// =============================================================================

func TestFilters_Registry(t *testing.T) {
	expectedFilters := []string{
		"created-at", "updated-at", "completed-at", "canceled-at", "started-at", "target-date",
		"id", "creator", "lead",
		"name", "slug-id",
		"health",
		"priority",
		"has-blocked-by", "has-blocking", "has-related",
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

func TestFilterBuilder_CanceledAtComparator(t *testing.T) {
	b := testBuilder()
	comp1 := b.CanceledAtComparator()
	comp2 := b.CanceledAtComparator()

	if comp1 != comp2 {
		t.Error("CanceledAtComparator() should return same instance")
	}
}

func TestFilterBuilder_StartedAtComparator(t *testing.T) {
	b := testBuilder()
	comp1 := b.StartedAtComparator()
	comp2 := b.StartedAtComparator()

	if comp1 != comp2 {
		t.Error("StartedAtComparator() should return same instance")
	}
}

func TestFilterBuilder_TargetDateComparator(t *testing.T) {
	b := testBuilder()
	comp1 := b.TargetDateComparator()
	comp2 := b.TargetDateComparator()

	if comp1 != comp2 {
		t.Error("TargetDateComparator() should return same instance")
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
			name: "with Health",
			setup: func(b *FilterBuilder) {
				health := "onTrack"
				b.filter.Health = &intgraphql.StringComparator{Eq: &health}
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
