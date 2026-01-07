package issue

import (
	"reflect"

	"github.com/chainguard-sandbox/go-linear/internal/dateparser"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
)

// IssueFilterBuilder builds an IssueFilter from Cobra command flags.
// Implements FilterBuilder and various filter-specific interfaces.
type IssueFilterBuilder struct {
	filter   *intgraphql.IssueFilter
	resolver *resolver.Resolver
	parser   dateparser.Parser
}

// NewIssueFilterBuilder creates a new filter builder.
func NewIssueFilterBuilder(res *resolver.Resolver) *IssueFilterBuilder {
	return &IssueFilterBuilder{
		filter:   &intgraphql.IssueFilter{},
		resolver: res,
		parser:   dateparser.New(),
	}
}

// Resolver returns the resolver for name-to-ID lookups.
func (b *IssueFilterBuilder) Resolver() *resolver.Resolver {
	return b.resolver
}

// Parser returns the date parser.
func (b *IssueFilterBuilder) Parser() dateparser.Parser {
	return b.parser
}

// Filter returns the underlying IssueFilter for direct manipulation.
// Used by filter functions that need entity-specific access.
func (b *IssueFilterBuilder) Filter() *intgraphql.IssueFilter {
	return b.filter
}

// CreatedAtComparator returns the DateComparator for CreatedAt, creating if needed.
// Implements CreatedAtFilterable interface.
func (b *IssueFilterBuilder) CreatedAtComparator() *intgraphql.DateComparator {
	if b.filter.CreatedAt == nil {
		b.filter.CreatedAt = &intgraphql.DateComparator{}
	}
	return b.filter.CreatedAt
}

// UpdatedAtComparator returns the DateComparator for UpdatedAt, creating if needed.
func (b *IssueFilterBuilder) UpdatedAtComparator() *intgraphql.DateComparator {
	if b.filter.UpdatedAt == nil {
		b.filter.UpdatedAt = &intgraphql.DateComparator{}
	}
	return b.filter.UpdatedAt
}

// CompletedAtComparator returns the NullableDateComparator for CompletedAt, creating if needed.
func (b *IssueFilterBuilder) CompletedAtComparator() *intgraphql.NullableDateComparator {
	if b.filter.CompletedAt == nil {
		b.filter.CompletedAt = &intgraphql.NullableDateComparator{}
	}
	return b.filter.CompletedAt
}

// StartedAtComparator returns the NullableDateComparator for StartedAt, creating if needed.
func (b *IssueFilterBuilder) StartedAtComparator() *intgraphql.NullableDateComparator {
	if b.filter.StartedAt == nil {
		b.filter.StartedAt = &intgraphql.NullableDateComparator{}
	}
	return b.filter.StartedAt
}

// ArchivedAtComparator returns the NullableDateComparator for ArchivedAt, creating if needed.
func (b *IssueFilterBuilder) ArchivedAtComparator() *intgraphql.NullableDateComparator {
	if b.filter.ArchivedAt == nil {
		b.filter.ArchivedAt = &intgraphql.NullableDateComparator{}
	}
	return b.filter.ArchivedAt
}

// Build returns the constructed filter.
// Returns nil if no filters were set.
func (b *IssueFilterBuilder) Build() *intgraphql.IssueFilter {
	// Check if any filters were actually set using helper
	if !hasAnyFilter(
		b.filter.Team,
		b.filter.Assignee,
		b.filter.State,
		b.filter.Priority,
		b.filter.CreatedAt,
		b.filter.CompletedAt,
		b.filter.UpdatedAt,
		b.filter.Labels,
		b.filter.Creator,
		b.filter.Cycle,
		b.filter.Project,
		b.filter.Parent,
		b.filter.ArchivedAt,
		b.filter.CanceledAt,
		b.filter.StartedAt,
		b.filter.TriagedAt,
		b.filter.HasSuggestedTeams,
		b.filter.HasSuggestedAssignees,
		b.filter.HasSuggestedProjects,
		b.filter.HasSuggestedLabels,
		b.filter.Comments,
		b.filter.Attachments,
		b.filter.Subscribers,
		b.filter.Children,
		b.filter.Needs,
		b.filter.Reactions,
		b.filter.DueDate,
		b.filter.Estimate,
		b.filter.Title,
		b.filter.Description,
		b.filter.Number,
		b.filter.ID,
		b.filter.Delegate,
		b.filter.SnoozedBy,
		b.filter.SnoozedUntilAt,
		b.filter.AutoArchivedAt,
		b.filter.AutoClosedAt,
		b.filter.AddedToCycleAt,
		b.filter.AddedToCyclePeriod,
		b.filter.ProjectMilestone,
		b.filter.LastAppliedTemplate,
		b.filter.CustomerCount,
		b.filter.CustomerImportantCount,
		b.filter.HasBlockedByRelations,
		b.filter.HasBlockingRelations,
		b.filter.HasDuplicateRelations,
		b.filter.HasRelatedRelations,
		b.filter.SLAStatus,
	) {
		return nil
	}

	return b.filter
}

// hasAnyFilter returns true if any of the provided pointer values is non-nil.
// Uses reflection to correctly handle nil interface values.
func hasAnyFilter(filters ...any) bool {
	for _, f := range filters {
		if f != nil && !reflect.ValueOf(f).IsNil() {
			return true
		}
	}
	return false
}
