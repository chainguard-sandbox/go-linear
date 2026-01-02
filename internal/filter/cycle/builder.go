// Package cycle provides filter building utilities for Linear cycles.
package cycle

import (
	"reflect"

	"github.com/chainguard-sandbox/go-linear/internal/dateparser"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
)

// FilterBuilder builds a CycleFilter from Cobra command flags.
type FilterBuilder struct {
	filter   *intgraphql.CycleFilter
	resolver *resolver.Resolver
	parser   *dateparser.Parser
}

// NewFilterBuilder creates a new cycle filter builder.
func NewFilterBuilder(res *resolver.Resolver) *FilterBuilder {
	return &FilterBuilder{
		filter:   &intgraphql.CycleFilter{},
		resolver: res,
		parser:   dateparser.New(),
	}
}

// Resolver returns the resolver for name-to-ID lookups.
func (b *FilterBuilder) Resolver() *resolver.Resolver {
	return b.resolver
}

// Parser returns the date parser.
func (b *FilterBuilder) Parser() *dateparser.Parser {
	return b.parser
}

// Filter returns the underlying CycleFilter for direct manipulation.
func (b *FilterBuilder) Filter() *intgraphql.CycleFilter {
	return b.filter
}

// CreatedAtComparator returns the DateComparator for CreatedAt, creating if needed.
func (b *FilterBuilder) CreatedAtComparator() *intgraphql.DateComparator {
	if b.filter.CreatedAt == nil {
		b.filter.CreatedAt = &intgraphql.DateComparator{}
	}
	return b.filter.CreatedAt
}

// UpdatedAtComparator returns the DateComparator for UpdatedAt, creating if needed.
func (b *FilterBuilder) UpdatedAtComparator() *intgraphql.DateComparator {
	if b.filter.UpdatedAt == nil {
		b.filter.UpdatedAt = &intgraphql.DateComparator{}
	}
	return b.filter.UpdatedAt
}

// CompletedAtComparator returns the DateComparator for CompletedAt, creating if needed.
func (b *FilterBuilder) CompletedAtComparator() *intgraphql.DateComparator {
	if b.filter.CompletedAt == nil {
		b.filter.CompletedAt = &intgraphql.DateComparator{}
	}
	return b.filter.CompletedAt
}

// StartsAtComparator returns the DateComparator for StartsAt, creating if needed.
func (b *FilterBuilder) StartsAtComparator() *intgraphql.DateComparator {
	if b.filter.StartsAt == nil {
		b.filter.StartsAt = &intgraphql.DateComparator{}
	}
	return b.filter.StartsAt
}

// EndsAtComparator returns the DateComparator for EndsAt, creating if needed.
func (b *FilterBuilder) EndsAtComparator() *intgraphql.DateComparator {
	if b.filter.EndsAt == nil {
		b.filter.EndsAt = &intgraphql.DateComparator{}
	}
	return b.filter.EndsAt
}

// Build returns the constructed filter.
// Returns nil if no filters were set.
func (b *FilterBuilder) Build() *intgraphql.CycleFilter {
	if !hasAnyFilter(
		b.filter.ID,
		b.filter.CreatedAt,
		b.filter.UpdatedAt,
		b.filter.CompletedAt,
		b.filter.StartsAt,
		b.filter.EndsAt,
		b.filter.Name,
		b.filter.Number,
		b.filter.Team,
		b.filter.IsActive,
		b.filter.IsFuture,
		b.filter.IsPast,
		b.filter.IsNext,
		b.filter.IsPrevious,
		b.filter.IsInCooldown,
		b.filter.Issues,
	) {
		return nil
	}

	return b.filter
}

// hasAnyFilter returns true if any of the provided pointer values is non-nil.
func hasAnyFilter(filters ...any) bool {
	for _, f := range filters {
		if f != nil && !reflect.ValueOf(f).IsNil() {
			return true
		}
	}
	return false
}
