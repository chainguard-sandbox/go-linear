package issue

import (
	"context"
	"testing"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/dateparser"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// testBuilder creates an IssueFilterBuilder with a nil resolver for unit tests.
// Use this for tests that don't require API resolution.
func testBuilder() *IssueFilterBuilder {
	return &IssueFilterBuilder{
		filter: &intgraphql.IssueFilter{},
		parser: dateparser.New(),
	}
}

// mustSet sets a flag value, panicking on error (acceptable in tests).
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
				mustSet(cmd, name, intToString(v))
			}
		case bool:
			cmd.Flags().Bool(name, false, "")
			if v {
				mustSet(cmd, name, "true")
			}
		case []string:
			cmd.Flags().StringArray(name, nil, "")
			for _, s := range v {
				mustSet(cmd, name, s)
			}
		}
	}

	return cmd
}

func intToString(i int) string {
	return string(rune('0'+i%10)) + func() string {
		if i >= 10 {
			return intToString(i / 10)
		}
		return ""
	}()
}

// Helper to convert int to string without importing strconv
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
// Date Filter Tests
// =============================================================================

func TestDateFilters(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		filterFunc IssueFilterFunc
		afterFlag  string
		beforeFlag string
		afterVal   string
		beforeVal  string
		wantErr    bool
		wantNil    bool
	}{
		// created-at
		{
			name:       "created-at with ISO8601 after",
			filterFunc: ApplyCreatedAt,
			afterFlag:  "created-after",
			afterVal:   "2025-01-01",
			wantNil:    false,
		},
		{
			name:       "created-at with ISO8601 before",
			filterFunc: ApplyCreatedAt,
			beforeFlag: "created-before",
			beforeVal:  "2025-12-31",
			wantNil:    false,
		},
		{
			name:       "created-at with relative date",
			filterFunc: ApplyCreatedAt,
			afterFlag:  "created-after",
			afterVal:   "7d",
			wantNil:    false,
		},
		{
			name:       "created-at with yesterday",
			filterFunc: ApplyCreatedAt,
			afterFlag:  "created-after",
			afterVal:   "yesterday",
			wantNil:    false,
		},
		{
			name:       "created-at empty returns nil",
			filterFunc: ApplyCreatedAt,
			afterFlag:  "created-after",
			beforeFlag: "created-before",
			wantNil:    true,
		},
		{
			name:       "created-at invalid date returns error",
			filterFunc: ApplyCreatedAt,
			afterFlag:  "created-after",
			afterVal:   "not-a-date",
			wantErr:    true,
		},
		// updated-at
		{
			name:       "updated-at with date",
			filterFunc: ApplyUpdatedAt,
			afterFlag:  "updated-after",
			afterVal:   "2025-01-01",
			wantNil:    false,
		},
		// completed-at
		{
			name:       "completed-at with date",
			filterFunc: ApplyCompletedAt,
			afterFlag:  "completed-after",
			afterVal:   "2025-01-01",
			wantNil:    false,
		},
		// started-at
		{
			name:       "started-at with date",
			filterFunc: ApplyStartedAt,
			afterFlag:  "started-after",
			afterVal:   "2025-01-01",
			wantNil:    false,
		},
		// archived-at
		{
			name:       "archived-at with date",
			filterFunc: ApplyArchivedAt,
			afterFlag:  "archived-after",
			afterVal:   "2025-01-01",
			wantNil:    false,
		},
		// canceled-at
		{
			name:       "canceled-at with date",
			filterFunc: ApplyCanceledAt,
			afterFlag:  "canceled-after",
			afterVal:   "2025-01-01",
			wantNil:    false,
		},
		// triaged-at
		{
			name:       "triaged-at with date",
			filterFunc: ApplyTriagedAt,
			afterFlag:  "triaged-after",
			afterVal:   "2025-01-01",
			wantNil:    false,
		},
		// auto-archived-at
		{
			name:       "auto-archived-at with date",
			filterFunc: ApplyAutoArchivedAt,
			afterFlag:  "auto-archived-after",
			afterVal:   "2025-01-01",
			wantNil:    false,
		},
		// auto-closed-at
		{
			name:       "auto-closed-at with date",
			filterFunc: ApplyAutoClosedAt,
			afterFlag:  "auto-closed-after",
			afterVal:   "2025-01-01",
			wantNil:    false,
		},
		// added-to-cycle-at
		{
			name:       "added-to-cycle-at with date",
			filterFunc: ApplyAddedToCycleAt,
			afterFlag:  "added-to-cycle-after",
			afterVal:   "2025-01-01",
			wantNil:    false,
		},
		// snoozed-until-at
		{
			name:       "snoozed-until-at with date",
			filterFunc: ApplySnoozedUntilAt,
			afterFlag:  "snoozed-until-after",
			afterVal:   "2025-01-01",
			wantNil:    false,
		},
		// due-date
		{
			name:       "due-date with date",
			filterFunc: ApplyDueDate,
			afterFlag:  "due-after",
			afterVal:   "2025-01-01",
			wantNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := testBuilder()
			flags := make(map[string]any)
			if tt.afterFlag != "" {
				flags[tt.afterFlag] = tt.afterVal
			}
			if tt.beforeFlag != "" {
				flags[tt.beforeFlag] = tt.beforeVal
			}
			cmd := testCommand(flags)

			err := tt.filterFunc(ctx, cmd, builder)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			filter := builder.Build()
			if tt.wantNil && filter != nil {
				t.Error("expected nil filter, got non-nil")
			}
			if !tt.wantNil && filter == nil {
				t.Error("expected non-nil filter, got nil")
			}
		})
	}
}

// =============================================================================
// Numeric Filter Tests
// =============================================================================

func TestNumericFilters(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		filterFunc IssueFilterFunc
		flagName   string
		value      int
		wantNil    bool
		validate   func(t *testing.T, b *IssueFilterBuilder)
	}{
		// priority
		{
			name:       "priority filter with value 0",
			filterFunc: ApplyPriority,
			flagName:   "priority",
			value:      0,
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.Priority == nil || b.filter.Priority.Eq == nil {
					t.Error("Priority filter not set")
					return
				}
				if *b.filter.Priority.Eq != 0.0 {
					t.Errorf("Priority = %f, want 0.0", *b.filter.Priority.Eq)
				}
			},
		},
		{
			name:       "priority filter with value 1 (urgent)",
			filterFunc: ApplyPriority,
			flagName:   "priority",
			value:      1,
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.Priority == nil || b.filter.Priority.Eq == nil {
					t.Error("Priority filter not set")
					return
				}
				if *b.filter.Priority.Eq != 1.0 {
					t.Errorf("Priority = %f, want 1.0", *b.filter.Priority.Eq)
				}
			},
		},
		{
			name:       "priority filter with value 4 (low)",
			filterFunc: ApplyPriority,
			flagName:   "priority",
			value:      4,
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if *b.filter.Priority.Eq != 4.0 {
					t.Errorf("Priority = %f, want 4.0", *b.filter.Priority.Eq)
				}
			},
		},
		{
			name:       "priority filter not set (default -1)",
			filterFunc: ApplyPriority,
			flagName:   "priority",
			value:      -1,
			wantNil:    true,
		},
		// estimate
		{
			name:       "estimate filter with value",
			filterFunc: ApplyEstimate,
			flagName:   "estimate",
			value:      5,
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.Estimate == nil || b.filter.Estimate.Eq == nil {
					t.Error("Estimate filter not set")
					return
				}
				if *b.filter.Estimate.Eq != 5.0 {
					t.Errorf("Estimate = %f, want 5.0", *b.filter.Estimate.Eq)
				}
			},
		},
		{
			name:       "estimate filter not set",
			filterFunc: ApplyEstimate,
			flagName:   "estimate",
			value:      -1,
			wantNil:    true,
		},
		// number
		{
			name:       "number filter with value",
			filterFunc: ApplyNumber,
			flagName:   "number",
			value:      123,
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.Number == nil || b.filter.Number.Eq == nil {
					t.Error("Number filter not set")
					return
				}
				if *b.filter.Number.Eq != 123.0 {
					t.Errorf("Number = %f, want 123.0", *b.filter.Number.Eq)
				}
			},
		},
		{
			name:       "number filter not set",
			filterFunc: ApplyNumber,
			flagName:   "number",
			value:      -1,
			wantNil:    true,
		},
		// customer-count
		{
			name:       "customer-count filter with value",
			filterFunc: ApplyCustomerCount,
			flagName:   "customer-count",
			value:      10,
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.CustomerCount == nil || b.filter.CustomerCount.Eq == nil {
					t.Error("CustomerCount filter not set")
					return
				}
				if *b.filter.CustomerCount.Eq != 10.0 {
					t.Errorf("CustomerCount = %f, want 10.0", *b.filter.CustomerCount.Eq)
				}
			},
		},
		{
			name:       "customer-count filter not set",
			filterFunc: ApplyCustomerCount,
			flagName:   "customer-count",
			value:      -1,
			wantNil:    true,
		},
		// customer-important-count
		{
			name:       "customer-important-count filter with value",
			filterFunc: ApplyCustomerImportantCount,
			flagName:   "customer-important-count",
			value:      5,
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.CustomerImportantCount == nil || b.filter.CustomerImportantCount.Eq == nil {
					t.Error("CustomerImportantCount filter not set")
					return
				}
				if *b.filter.CustomerImportantCount.Eq != 5.0 {
					t.Errorf("CustomerImportantCount = %f, want 5.0", *b.filter.CustomerImportantCount.Eq)
				}
			},
		},
		{
			name:       "customer-important-count filter not set",
			filterFunc: ApplyCustomerImportantCount,
			flagName:   "customer-important-count",
			value:      -1,
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := testBuilder()
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().Int(tt.flagName, -1, "")
			if tt.value >= 0 {
				mustSet(cmd, tt.flagName, itoa(tt.value))
			}

			err := tt.filterFunc(ctx, cmd, builder)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			filter := builder.Build()
			if tt.wantNil {
				if filter != nil {
					t.Error("expected nil filter, got non-nil")
				}
				return
			}
			if filter == nil {
				t.Fatal("expected non-nil filter, got nil")
			}
			if tt.validate != nil {
				tt.validate(t, builder)
			}
		})
	}
}

// =============================================================================
// Boolean Filter Tests
// =============================================================================

func TestBooleanFilters(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		filterFunc IssueFilterFunc
		flagName   string
		value      bool
		wantNil    bool
		validate   func(t *testing.T, b *IssueFilterBuilder)
	}{
		// has-blocked-by
		{
			name:       "has-blocked-by true",
			filterFunc: ApplyHasBlockedByRelations,
			flagName:   "has-blocked-by",
			value:      true,
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.HasBlockedByRelations == nil {
					t.Error("HasBlockedByRelations not set")
					return
				}
				if b.filter.HasBlockedByRelations.Eq == nil || !*b.filter.HasBlockedByRelations.Eq {
					t.Error("HasBlockedByRelations.Eq should be true")
				}
			},
		},
		{
			name:       "has-blocked-by false (not set)",
			filterFunc: ApplyHasBlockedByRelations,
			flagName:   "has-blocked-by",
			value:      false,
			wantNil:    true,
		},
		// has-blocking
		{
			name:       "has-blocking true",
			filterFunc: ApplyHasBlockingRelations,
			flagName:   "has-blocking",
			value:      true,
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.HasBlockingRelations == nil {
					t.Error("HasBlockingRelations not set")
				}
			},
		},
		{
			name:       "has-blocking false",
			filterFunc: ApplyHasBlockingRelations,
			flagName:   "has-blocking",
			value:      false,
			wantNil:    true,
		},
		// has-duplicate
		{
			name:       "has-duplicate true",
			filterFunc: ApplyHasDuplicateRelations,
			flagName:   "has-duplicate",
			value:      true,
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.HasDuplicateRelations == nil {
					t.Error("HasDuplicateRelations not set")
				}
			},
		},
		{
			name:       "has-duplicate false",
			filterFunc: ApplyHasDuplicateRelations,
			flagName:   "has-duplicate",
			value:      false,
			wantNil:    true,
		},
		// has-related
		{
			name:       "has-related true",
			filterFunc: ApplyHasRelatedRelations,
			flagName:   "has-related",
			value:      true,
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.HasRelatedRelations == nil {
					t.Error("HasRelatedRelations not set")
				}
			},
		},
		{
			name:       "has-related false",
			filterFunc: ApplyHasRelatedRelations,
			flagName:   "has-related",
			value:      false,
			wantNil:    true,
		},
		// has-suggested-assignees
		{
			name:       "has-suggested-assignees true",
			filterFunc: ApplyHasSuggestedAssignees,
			flagName:   "has-suggested-assignees",
			value:      true,
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.HasSuggestedAssignees == nil {
					t.Error("HasSuggestedAssignees not set")
				}
			},
		},
		{
			name:       "has-suggested-assignees false",
			filterFunc: ApplyHasSuggestedAssignees,
			flagName:   "has-suggested-assignees",
			value:      false,
			wantNil:    true,
		},
		// has-suggested-labels
		{
			name:       "has-suggested-labels true",
			filterFunc: ApplyHasSuggestedLabels,
			flagName:   "has-suggested-labels",
			value:      true,
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.HasSuggestedLabels == nil {
					t.Error("HasSuggestedLabels not set")
				}
			},
		},
		{
			name:       "has-suggested-labels false",
			filterFunc: ApplyHasSuggestedLabels,
			flagName:   "has-suggested-labels",
			value:      false,
			wantNil:    true,
		},
		// has-suggested-projects
		{
			name:       "has-suggested-projects true",
			filterFunc: ApplyHasSuggestedProjects,
			flagName:   "has-suggested-projects",
			value:      true,
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.HasSuggestedProjects == nil {
					t.Error("HasSuggestedProjects not set")
				}
			},
		},
		{
			name:       "has-suggested-projects false",
			filterFunc: ApplyHasSuggestedProjects,
			flagName:   "has-suggested-projects",
			value:      false,
			wantNil:    true,
		},
		// has-suggested-teams
		{
			name:       "has-suggested-teams true",
			filterFunc: ApplyHasSuggestedTeams,
			flagName:   "has-suggested-teams",
			value:      true,
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.HasSuggestedTeams == nil {
					t.Error("HasSuggestedTeams not set")
				}
			},
		},
		{
			name:       "has-suggested-teams false",
			filterFunc: ApplyHasSuggestedTeams,
			flagName:   "has-suggested-teams",
			value:      false,
			wantNil:    true,
		},
		// children (has-children)
		{
			name:       "has-children true",
			filterFunc: ApplyChildren,
			flagName:   "has-children",
			value:      true,
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.Children == nil {
					t.Error("Children not set")
				}
			},
		},
		{
			name:       "has-children false",
			filterFunc: ApplyChildren,
			flagName:   "has-children",
			value:      false,
			wantNil:    true,
		},
		// reactions (has-reactions)
		{
			name:       "has-reactions true",
			filterFunc: ApplyReactions,
			flagName:   "has-reactions",
			value:      true,
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.Reactions == nil {
					t.Error("Reactions not set")
				}
			},
		},
		{
			name:       "has-reactions false",
			filterFunc: ApplyReactions,
			flagName:   "has-reactions",
			value:      false,
			wantNil:    true,
		},
		// needs (has-needs)
		{
			name:       "has-needs true",
			filterFunc: ApplyNeeds,
			flagName:   "has-needs",
			value:      true,
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.Needs == nil {
					t.Error("Needs not set")
				}
			},
		},
		{
			name:       "has-needs false",
			filterFunc: ApplyNeeds,
			flagName:   "has-needs",
			value:      false,
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := testBuilder()
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().Bool(tt.flagName, false, "")
			if tt.value {
				mustSet(cmd, tt.flagName, "true")
			}

			err := tt.filterFunc(ctx, cmd, builder)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			filter := builder.Build()
			if tt.wantNil {
				if filter != nil {
					t.Error("expected nil filter, got non-nil")
				}
				return
			}
			if filter == nil {
				t.Fatal("expected non-nil filter, got nil")
			}
			if tt.validate != nil {
				tt.validate(t, builder)
			}
		})
	}
}

// =============================================================================
// String Filter Tests
// =============================================================================

func TestStringFilters(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		filterFunc IssueFilterFunc
		flagName   string
		value      string
		wantNil    bool
		validate   func(t *testing.T, b *IssueFilterBuilder)
	}{
		// title
		{
			name:       "title filter with value",
			filterFunc: ApplyTitle,
			flagName:   "title",
			value:      "bug fix",
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.Title == nil || b.filter.Title.Contains == nil {
					t.Error("Title filter not set")
					return
				}
				if *b.filter.Title.Contains != "bug fix" {
					t.Errorf("Title = %s, want 'bug fix'", *b.filter.Title.Contains)
				}
			},
		},
		{
			name:       "title filter empty",
			filterFunc: ApplyTitle,
			flagName:   "title",
			value:      "",
			wantNil:    true,
		},
		// description
		{
			name:       "description filter with value",
			filterFunc: ApplyDescription,
			flagName:   "description",
			value:      "TODO",
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.Description == nil || b.filter.Description.Contains == nil {
					t.Error("Description filter not set")
					return
				}
				if *b.filter.Description.Contains != "TODO" {
					t.Errorf("Description = %s, want 'TODO'", *b.filter.Description.Contains)
				}
			},
		},
		{
			name:       "description filter empty",
			filterFunc: ApplyDescription,
			flagName:   "description",
			value:      "",
			wantNil:    true,
		},
		// id
		{
			name:       "id filter with UUID",
			filterFunc: ApplyID,
			flagName:   "id",
			value:      "550e8400-e29b-41d4-a716-446655440000",
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.ID == nil || b.filter.ID.Eq == nil {
					t.Error("ID filter not set")
					return
				}
				if *b.filter.ID.Eq != "550e8400-e29b-41d4-a716-446655440000" {
					t.Errorf("ID = %s, want UUID", *b.filter.ID.Eq)
				}
			},
		},
		{
			name:       "id filter empty",
			filterFunc: ApplyID,
			flagName:   "id",
			value:      "",
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := testBuilder()
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String(tt.flagName, "", "")
			if tt.value != "" {
				mustSet(cmd, tt.flagName, tt.value)
			}

			err := tt.filterFunc(ctx, cmd, builder)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			filter := builder.Build()
			if tt.wantNil {
				if filter != nil {
					t.Error("expected nil filter, got non-nil")
				}
				return
			}
			if filter == nil {
				t.Fatal("expected non-nil filter, got nil")
			}
			if tt.validate != nil {
				tt.validate(t, builder)
			}
		})
	}
}

// =============================================================================
// Enum Filter Tests
// =============================================================================

func TestEnumFilters(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		filterFunc IssueFilterFunc
		flagName   string
		value      string
		wantNil    bool
		validate   func(t *testing.T, b *IssueFilterBuilder)
	}{
		// sla-status
		{
			name:       "sla-status Breached",
			filterFunc: ApplySLAStatus,
			flagName:   "sla-status",
			value:      "Breached",
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.SLAStatus == nil || b.filter.SLAStatus.Eq == nil {
					t.Error("SLAStatus filter not set")
					return
				}
				if string(*b.filter.SLAStatus.Eq) != "Breached" {
					t.Errorf("SLAStatus = %s, want 'Breached'", string(*b.filter.SLAStatus.Eq))
				}
			},
		},
		{
			name:       "sla-status HighRisk",
			filterFunc: ApplySLAStatus,
			flagName:   "sla-status",
			value:      "HighRisk",
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if string(*b.filter.SLAStatus.Eq) != "HighRisk" {
					t.Errorf("SLAStatus = %s, want 'HighRisk'", string(*b.filter.SLAStatus.Eq))
				}
			},
		},
		{
			name:       "sla-status empty",
			filterFunc: ApplySLAStatus,
			flagName:   "sla-status",
			value:      "",
			wantNil:    true,
		},
		// added-to-cycle-period
		{
			name:       "added-to-cycle-period before",
			filterFunc: ApplyAddedToCyclePeriod,
			flagName:   "added-to-cycle-period",
			value:      "before",
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.AddedToCyclePeriod == nil || b.filter.AddedToCyclePeriod.Eq == nil {
					t.Error("AddedToCyclePeriod filter not set")
					return
				}
				if string(*b.filter.AddedToCyclePeriod.Eq) != "before" {
					t.Errorf("AddedToCyclePeriod = %s, want 'before'", string(*b.filter.AddedToCyclePeriod.Eq))
				}
			},
		},
		{
			name:       "added-to-cycle-period during",
			filterFunc: ApplyAddedToCyclePeriod,
			flagName:   "added-to-cycle-period",
			value:      "during",
			wantNil:    false,
		},
		{
			name:       "added-to-cycle-period after",
			filterFunc: ApplyAddedToCyclePeriod,
			flagName:   "added-to-cycle-period",
			value:      "after",
			wantNil:    false,
		},
		{
			name:       "added-to-cycle-period empty",
			filterFunc: ApplyAddedToCyclePeriod,
			flagName:   "added-to-cycle-period",
			value:      "",
			wantNil:    true,
		},
		// last-applied-template
		{
			name:       "last-applied-template with UUID",
			filterFunc: ApplyLastAppliedTemplate,
			flagName:   "last-applied-template",
			value:      "550e8400-e29b-41d4-a716-446655440000",
			wantNil:    false,
			validate: func(t *testing.T, b *IssueFilterBuilder) {
				if b.filter.LastAppliedTemplate == nil || b.filter.LastAppliedTemplate.ID == nil {
					t.Error("LastAppliedTemplate filter not set")
					return
				}
			},
		},
		{
			name:       "last-applied-template empty",
			filterFunc: ApplyLastAppliedTemplate,
			flagName:   "last-applied-template",
			value:      "",
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := testBuilder()
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String(tt.flagName, "", "")
			if tt.value != "" {
				mustSet(cmd, tt.flagName, tt.value)
			}

			err := tt.filterFunc(ctx, cmd, builder)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			filter := builder.Build()
			if tt.wantNil {
				if filter != nil {
					t.Error("expected nil filter, got non-nil")
				}
				return
			}
			if filter == nil {
				t.Fatal("expected non-nil filter, got nil")
			}
			if tt.validate != nil {
				tt.validate(t, builder)
			}
		})
	}
}

// =============================================================================
// Edge Cases and Regression Tests
// =============================================================================

func TestFilterEdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("flag not defined returns nil without error", func(t *testing.T) {
		builder := testBuilder()
		cmd := &cobra.Command{Use: "test"}
		// Don't define any flags

		// All filters should handle missing flags gracefully
		filters := []IssueFilterFunc{
			ApplyPriority,
			ApplyEstimate,
			ApplyNumber,
			ApplyCustomerCount,
			ApplyCustomerImportantCount,
		}

		for _, f := range filters {
			if err := f(ctx, cmd, builder); err != nil {
				t.Errorf("filter returned error for undefined flag: %v", err)
			}
		}

		if builder.Build() != nil {
			t.Error("expected nil filter when no flags defined")
		}
	})

	t.Run("multiple filters combine correctly", func(t *testing.T) {
		builder := testBuilder()
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Int("priority", -1, "")
		cmd.Flags().String("title", "", "")
		cmd.Flags().Bool("has-blocked-by", false, "")

		mustSet(cmd, "priority", "1")
		mustSet(cmd, "title", "urgent")
		mustSet(cmd, "has-blocked-by", "true")

		_ = ApplyPriority(ctx, cmd, builder)
		_ = ApplyTitle(ctx, cmd, builder)
		_ = ApplyHasBlockedByRelations(ctx, cmd, builder)

		filter := builder.Build()
		if filter == nil {
			t.Fatal("expected non-nil filter")
		}
		if filter.Priority == nil {
			t.Error("Priority not set")
		}
		if filter.Title == nil {
			t.Error("Title not set")
		}
		if filter.HasBlockedByRelations == nil {
			t.Error("HasBlockedByRelations not set")
		}
	})

	t.Run("date range with both after and before", func(t *testing.T) {
		builder := testBuilder()
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("created-after", "", "")
		cmd.Flags().String("created-before", "", "")
		mustSet(cmd, "created-after", "2025-01-01")
		mustSet(cmd, "created-before", "2025-12-31")

		err := ApplyCreatedAt(ctx, cmd, builder)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.CreatedAt == nil {
			t.Fatal("CreatedAt filter not set")
		}
		if filter.CreatedAt.Gte == nil {
			t.Error("CreatedAt.Gte not set")
		}
		if filter.CreatedAt.Lt == nil {
			t.Error("CreatedAt.Lt not set")
		}
	})
}

// =============================================================================
// Entity Resolution Filter Tests (require mock resolver)
// =============================================================================

// mockResolver implements resolver methods for testing without API calls.
type mockResolver struct {
	teams    map[string]string
	states   map[string]string
	users    map[string]string
	labels   map[string]string
	projects map[string]string
	cycles   map[string]string
}

func newMockResolver() *mockResolver {
	return &mockResolver{
		teams:    map[string]string{"Engineering": "team-uuid-1", "Design": "team-uuid-2"},
		states:   map[string]string{"Todo": "state-uuid-1", "In Progress": "state-uuid-2", "Done": "state-uuid-3"},
		users:    map[string]string{"me": "user-uuid-1", "alice@example.com": "user-uuid-2"},
		labels:   map[string]string{"bug": "label-uuid-1", "feature": "label-uuid-2"},
		projects: map[string]string{"Q1 Goals": "project-uuid-1"},
		cycles:   map[string]string{"Sprint 1": "cycle-uuid-1"},
	}
}

// TestEntityResolutionFiltersWithMock tests entity filters that require resolution
// but don't need a live API connection.
func TestEntityResolutionFiltersWithMock(t *testing.T) {
	// These tests would require injecting a mock resolver into the builder.
	// The current implementation uses a real resolver, so these are placeholder
	// tests that document expected behavior.

	t.Run("entity filters require resolver", func(t *testing.T) {
		// Document that entity resolution filters (team, state, assignee, etc.)
		// require a resolver with API access and are tested in issue_test.go
		// with the 'read' build tag.
		entityFilters := []string{
			"team", "state", "assignee", "creator", "delegate",
			"cycle", "project", "project-milestone", "parent",
			"labels", "subscribers", "snoozed-by", "attachments", "comments",
		}
		t.Logf("Entity resolution filters requiring live API: %v", entityFilters)
		t.Log("These are tested in issue_test.go with //go:build read")
	})
}

// TestFromFlagsRegistry tests that FromFlags iterates through all registered filters.
func TestFromFlagsRegistry(t *testing.T) {
	t.Run("registry contains all expected filters", func(t *testing.T) {
		expectedFilters := []string{
			"created-at", "updated-at", "completed-at", "started-at", "archived-at",
			"added-to-cycle-at", "added-to-cycle-period", "assignee", "attachments",
			"auto-archived-at", "auto-closed-at", "canceled-at", "children", "comments",
			"creator", "customer-count", "customer-important-count", "cycle", "delegate",
			"description", "due-date", "estimate", "has-blocked-by", "has-blocking",
			"has-duplicate", "has-related", "has-suggested-assignees", "has-suggested-labels",
			"has-suggested-projects", "has-suggested-teams", "id", "labels",
			"last-applied-template", "needs", "number", "parent", "priority", "project",
			"project-milestone", "reactions", "sla-status", "snoozed-by", "snoozed-until-at",
			"state", "subscribers", "team", "title", "triaged-at",
		}

		if len(IssueFilters) != len(expectedFilters) {
			t.Errorf("IssueFilters has %d entries, expected %d", len(IssueFilters), len(expectedFilters))
		}

		for _, name := range expectedFilters {
			if _, ok := IssueFilters[name]; !ok {
				t.Errorf("IssueFilters missing filter: %s", name)
			}
		}
	})

	t.Run("individual filters apply correctly", func(t *testing.T) {
		// Test that individual filters work when called directly.
		// FromFlags requires all filter flags to be defined, so we test
		// individual filters instead.
		builder := testBuilder()

		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Int("priority", -1, "")
		cmd.Flags().String("title", "", "")
		cmd.Flags().Bool("has-blocked-by", false, "")

		mustSet(cmd, "priority", "2")
		mustSet(cmd, "title", "test")
		mustSet(cmd, "has-blocked-by", "true")

		ctx := context.Background()

		// Apply filters individually
		if err := ApplyPriority(ctx, cmd, builder); err != nil {
			t.Fatalf("ApplyPriority error: %v", err)
		}
		if err := ApplyTitle(ctx, cmd, builder); err != nil {
			t.Fatalf("ApplyTitle error: %v", err)
		}
		if err := ApplyHasBlockedByRelations(ctx, cmd, builder); err != nil {
			t.Fatalf("ApplyHasBlockedByRelations error: %v", err)
		}

		filter := builder.Build()
		if filter == nil {
			t.Fatal("expected non-nil filter")
		}
		if filter.Priority == nil {
			t.Error("Priority not applied")
		}
		if filter.Title == nil {
			t.Error("Title not applied")
		}
		if filter.HasBlockedByRelations == nil {
			t.Error("HasBlockedByRelations not applied")
		}
	})
}
