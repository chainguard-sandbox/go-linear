package project

import (
	"context"
	"reflect"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/dateparser"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
)

// FilterBuilder constructs ProjectFilter from CLI flags.
type FilterBuilder struct {
	filter   *intgraphql.ProjectFilter
	resolver *resolver.Resolver
	parser   *dateparser.Parser
}

// NewFilterBuilder creates a FilterBuilder with the given resolver.
func NewFilterBuilder(res *resolver.Resolver) *FilterBuilder {
	return &FilterBuilder{
		filter:   &intgraphql.ProjectFilter{},
		resolver: res,
		parser:   dateparser.New(),
	}
}

// Resolver returns the resolver for entity lookups.
func (b *FilterBuilder) Resolver() *resolver.Resolver {
	return b.resolver
}

// Parser returns the date parser.
func (b *FilterBuilder) Parser() *dateparser.Parser {
	return b.parser
}

// Filter returns the underlying ProjectFilter.
func (b *FilterBuilder) Filter() *intgraphql.ProjectFilter {
	return b.filter
}

// CreatedAtComparator returns or creates the CreatedAt comparator.
func (b *FilterBuilder) CreatedAtComparator() *intgraphql.DateComparator {
	if b.filter.CreatedAt == nil {
		b.filter.CreatedAt = &intgraphql.DateComparator{}
	}
	return b.filter.CreatedAt
}

// UpdatedAtComparator returns or creates the UpdatedAt comparator.
func (b *FilterBuilder) UpdatedAtComparator() *intgraphql.DateComparator {
	if b.filter.UpdatedAt == nil {
		b.filter.UpdatedAt = &intgraphql.DateComparator{}
	}
	return b.filter.UpdatedAt
}

// CompletedAtComparator returns or creates the CompletedAt comparator.
func (b *FilterBuilder) CompletedAtComparator() *intgraphql.NullableDateComparator {
	if b.filter.CompletedAt == nil {
		b.filter.CompletedAt = &intgraphql.NullableDateComparator{}
	}
	return b.filter.CompletedAt
}

// CanceledAtComparator returns or creates the CanceledAt comparator.
func (b *FilterBuilder) CanceledAtComparator() *intgraphql.NullableDateComparator {
	if b.filter.CanceledAt == nil {
		b.filter.CanceledAt = &intgraphql.NullableDateComparator{}
	}
	return b.filter.CanceledAt
}

// StartedAtComparator returns or creates the StartedAt comparator.
func (b *FilterBuilder) StartedAtComparator() *intgraphql.NullableDateComparator {
	if b.filter.StartedAt == nil {
		b.filter.StartedAt = &intgraphql.NullableDateComparator{}
	}
	return b.filter.StartedAt
}

// TargetDateComparator returns or creates the TargetDate comparator.
func (b *FilterBuilder) TargetDateComparator() *intgraphql.NullableDateComparator {
	if b.filter.TargetDate == nil {
		b.filter.TargetDate = &intgraphql.NullableDateComparator{}
	}
	return b.filter.TargetDate
}

// SetID sets the ID comparator. Implements common.IDFilterable.
func (b *FilterBuilder) SetID(comp *intgraphql.IDComparator) {
	b.filter.ID = comp
}

// Build returns the filter if any fields are set, nil otherwise.
func (b *FilterBuilder) Build() *intgraphql.ProjectFilter {
	if !hasAnyFilter(
		b.filter.ID,
		b.filter.Name,
		b.filter.SlugID,
		b.filter.CreatedAt,
		b.filter.UpdatedAt,
		b.filter.CompletedAt,
		b.filter.CanceledAt,
		b.filter.StartedAt,
		b.filter.TargetDate,
		b.filter.Creator,
		b.filter.Lead,
		b.filter.Health,
		b.filter.Priority,
		b.filter.HasBlockedByRelations,
		b.filter.HasBlockingRelations,
		b.filter.HasRelatedRelations,
	) {
		return nil
	}
	return b.filter
}

// hasAnyFilter returns true if any of the values is non-nil and non-zero.
func hasAnyFilter(values ...any) bool {
	for _, v := range values {
		if v != nil && !reflect.ValueOf(v).IsNil() {
			return true
		}
	}
	return false
}

// FromFlags populates the filter from cobra command flags.
func (b *FilterBuilder) FromFlags(ctx context.Context, cmd *cobra.Command) error {
	for _, applyFn := range Filters {
		if err := applyFn(ctx, cmd, b); err != nil {
			return err
		}
	}
	return nil
}
