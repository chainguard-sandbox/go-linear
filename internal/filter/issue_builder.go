package filter

import (
	"github.com/chainguard-sandbox/go-linear/internal/dateparser"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
)

// IssueFilterBuilder builds an IssueFilter from Cobra command flags.
// Implements FilterBuilder and various filter-specific interfaces.
type IssueFilterBuilder struct {
	filter   *intgraphql.IssueFilter
	resolver *resolver.Resolver
	parser   *dateparser.Parser
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
func (b *IssueFilterBuilder) Parser() *dateparser.Parser {
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
	// Check if any filters were actually set
	if b.filter.Team == nil &&
		b.filter.Assignee == nil &&
		b.filter.State == nil &&
		b.filter.Priority == nil &&
		b.filter.CreatedAt == nil &&
		b.filter.CompletedAt == nil &&
		b.filter.UpdatedAt == nil &&
		b.filter.Labels == nil &&
		b.filter.Creator == nil &&
		b.filter.Cycle == nil &&
		b.filter.Project == nil &&
		b.filter.Parent == nil &&
		b.filter.ArchivedAt == nil &&
		b.filter.CanceledAt == nil &&
		b.filter.StartedAt == nil &&
		b.filter.TriagedAt == nil &&
		b.filter.HasSuggestedTeams == nil &&
		b.filter.HasSuggestedAssignees == nil &&
		b.filter.HasSuggestedProjects == nil &&
		b.filter.HasSuggestedLabels == nil &&
		b.filter.Comments == nil &&
		b.filter.Attachments == nil &&
		b.filter.Subscribers == nil &&
		b.filter.Children == nil &&
		b.filter.Needs == nil &&
		b.filter.Reactions == nil &&
		b.filter.DueDate == nil &&
		b.filter.Estimate == nil &&
		b.filter.Title == nil &&
		b.filter.Description == nil &&
		b.filter.Number == nil &&
		b.filter.ID == nil &&
		b.filter.Delegate == nil &&
		b.filter.SnoozedBy == nil &&
		b.filter.SnoozedUntilAt == nil &&
		b.filter.AutoArchivedAt == nil &&
		b.filter.AutoClosedAt == nil &&
		b.filter.AddedToCycleAt == nil &&
		b.filter.AddedToCyclePeriod == nil &&
		b.filter.ProjectMilestone == nil &&
		b.filter.LastAppliedTemplate == nil &&
		b.filter.CustomerCount == nil &&
		b.filter.CustomerImportantCount == nil &&
		b.filter.HasBlockedByRelations == nil &&
		b.filter.HasBlockingRelations == nil &&
		b.filter.HasDuplicateRelations == nil &&
		b.filter.HasRelatedRelations == nil &&
		b.filter.SLAStatus == nil {
		return nil
	}

	return b.filter
}
