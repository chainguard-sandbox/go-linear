package document

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
		filter: &intgraphql.DocumentFilter{},
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

// =============================================================================
// Text Filter Tests
// =============================================================================

func TestApplyTitle(t *testing.T) {
	tests := []struct {
		name       string
		title      string
		wantFilter bool
	}{
		{name: "empty", title: "", wantFilter: false},
		{name: "set", title: "Design Doc", wantFilter: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBuilder()
			cmd := testCommand(map[string]any{"title": tt.title})

			err := ApplyTitle(context.Background(), cmd, b)
			if err != nil {
				t.Fatalf("ApplyTitle() error = %v", err)
			}

			hasFilter := b.filter.Title != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyTitle() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}

			if tt.wantFilter && b.filter.Title.ContainsIgnoreCase != nil && *b.filter.Title.ContainsIgnoreCase != tt.title {
				t.Errorf("ApplyTitle() Title = %v, want %v", *b.filter.Title.ContainsIgnoreCase, tt.title)
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
		{name: "set", slugID: "design-doc-123", wantFilter: true},
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

			if tt.wantFilter && b.filter.SlugID.Eq != nil && *b.filter.SlugID.Eq != tt.slugID {
				t.Errorf("ApplySlugID() SlugID = %v, want %v", *b.filter.SlugID.Eq, tt.slugID)
			}
		})
	}
}

// =============================================================================
// Entity Filter Tests (Empty Flag Path)
// =============================================================================

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

func TestApplyInitiative_Empty(t *testing.T) {
	b := testBuilder()
	cmd := testCommand(map[string]any{"initiative": ""})

	err := ApplyInitiative(context.Background(), cmd, b)
	if err != nil {
		t.Fatalf("ApplyInitiative() error = %v", err)
	}

	if b.filter.Initiative != nil {
		t.Error("Initiative filter should be nil when flag is empty")
	}
}

func TestApplyProject_Empty(t *testing.T) {
	b := testBuilder()
	cmd := testCommand(map[string]any{"project": ""})

	err := ApplyProject(context.Background(), cmd, b)
	if err != nil {
		t.Fatalf("ApplyProject() error = %v", err)
	}

	if b.filter.Project != nil {
		t.Error("Project filter should be nil when flag is empty")
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
// Updated At No Flags Test
// =============================================================================

func TestApplyUpdatedAt_Empty(t *testing.T) {
	b := testBuilder()
	cmd := testCommand(map[string]any{
		"updated-after":  "",
		"updated-before": "",
	})

	err := ApplyUpdatedAt(context.Background(), cmd, b)
	if err != nil {
		t.Fatalf("ApplyUpdatedAt() error = %v", err)
	}

	if b.filter.UpdatedAt != nil {
		t.Error("UpdatedAt filter should be nil when flags are empty")
	}
}

// =============================================================================
// Registry Tests
// =============================================================================

func TestFilters_Registry(t *testing.T) {
	expectedFilters := []string{
		"created-at", "updated-at",
		"id", "creator", "initiative", "project", "issue",
		"title", "slug-id",
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
			name: "with Title",
			setup: func(b *FilterBuilder) {
				title := "test"
				b.filter.Title = &intgraphql.StringComparator{ContainsIgnoreCase: &title}
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
