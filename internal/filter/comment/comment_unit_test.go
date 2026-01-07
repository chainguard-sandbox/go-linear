package comment

import (
	"context"
	"testing"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/dateparser"
	"github.com/chainguard-sandbox/go-linear/internal/filter/common"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// testBuilder creates a FilterBuilder with a nil resolver for unit tests.
func testBuilder() *FilterBuilder {
	return &FilterBuilder{
		filter: &intgraphql.CommentFilter{},
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
	// Parser is now a value type, always valid - test it works
	b := testBuilder()
	_, err := b.Parser().Parse("today")
	if err != nil {
		t.Errorf("Parser() should return working parser, got error: %v", err)
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

			err := common.ApplyCreatedAt(context.Background(), cmd, b)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBuilder()
			cmd := testCommand(map[string]any{
				"updated-after":  tt.after,
				"updated-before": tt.before,
			})

			err := common.ApplyUpdatedAt(context.Background(), cmd, b)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyUpdatedAt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			hasFilter := b.filter.UpdatedAt != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyUpdatedAt() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
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

			err := common.ApplyID(context.Background(), cmd, b)
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

	if b.filter.User != nil {
		t.Error("User filter should be nil when flag is empty")
	}
}

func TestApplyIssue_Empty(t *testing.T) {
	b := testBuilder()
	cmd := testCommand(map[string]any{"issue": ""})

	err := ApplyIssue(context.Background(), cmd, b)
	if err != nil {
		t.Fatalf("ApplyIssue() error = %v", err)
	}

	if b.filter.Issue != nil {
		t.Error("Issue filter should be nil when flag is empty")
	}
}

// =============================================================================
// Text Filter Tests
// =============================================================================

func TestApplyBody(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantFilter bool
	}{
		{name: "empty", body: "", wantFilter: false},
		{name: "set", body: "search text", wantFilter: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBuilder()
			cmd := testCommand(map[string]any{"body": tt.body})

			err := ApplyBody(context.Background(), cmd, b)
			if err != nil {
				t.Fatalf("ApplyBody() error = %v", err)
			}

			hasFilter := b.filter.Body != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyBody() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

// =============================================================================
// Registry Tests
// =============================================================================

func TestFilters_Registry(t *testing.T) {
	expectedFilters := []string{
		"created-at", "updated-at",
		"id", "creator", "issue",
		"body",
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
			name: "with Body",
			setup: func(b *FilterBuilder) {
				body := "search"
				b.filter.Body = &intgraphql.StringComparator{ContainsIgnoreCase: &body}
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
