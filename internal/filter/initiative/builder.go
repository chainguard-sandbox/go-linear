// Package initiative provides filtering for Linear initiatives.
package initiative

import (
	"context"
	"reflect"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/dateparser"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
)

// FilterFunc applies a filter based on command flags.
type FilterFunc func(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error

// FilterBuilder constructs InitiativeFilter from command flags.
type FilterBuilder struct {
	filter   *intgraphql.InitiativeFilter
	resolver *resolver.Resolver
	parser   *dateparser.Parser
}

// NewFilterBuilder creates a new FilterBuilder.
func NewFilterBuilder(res *resolver.Resolver) *FilterBuilder {
	return &FilterBuilder{
		filter:   &intgraphql.InitiativeFilter{},
		resolver: res,
		parser:   dateparser.New(),
	}
}

// Build returns the filter if any fields are set, otherwise nil.
func (b *FilterBuilder) Build() *intgraphql.InitiativeFilter {
	if !hasAnyFilter(
		b.filter.ID,
		b.filter.CreatedAt,
		b.filter.TargetDate,
		b.filter.Creator,
		b.filter.Owner,
		b.filter.Name,
		b.filter.SlugID,
		b.filter.Health,
		b.filter.Status,
	) {
		return nil
	}
	return b.filter
}

// Resolver returns the resolver for entity lookups.
func (b *FilterBuilder) Resolver() *resolver.Resolver {
	return b.resolver
}

// Parser returns the date parser.
func (b *FilterBuilder) Parser() *dateparser.Parser {
	return b.parser
}

// Filter returns the underlying filter for direct access.
func (b *FilterBuilder) Filter() *intgraphql.InitiativeFilter {
	return b.filter
}

// FromFlags applies all registered filters from command flags.
func (b *FilterBuilder) FromFlags(ctx context.Context, cmd *cobra.Command) error {
	for _, fn := range Filters {
		if err := fn(ctx, cmd, b); err != nil {
			return err
		}
	}
	return nil
}

// CreatedAtComparator returns the CreatedAt comparator, creating if needed.
func (b *FilterBuilder) CreatedAtComparator() *intgraphql.DateComparator {
	if b.filter.CreatedAt == nil {
		b.filter.CreatedAt = &intgraphql.DateComparator{}
	}
	return b.filter.CreatedAt
}

// TargetDateComparator returns the TargetDate comparator, creating if needed.
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

// hasAnyFilter checks if any of the provided values are non-nil.
func hasAnyFilter(values ...any) bool {
	for _, v := range values {
		if v != nil && !reflect.ValueOf(v).IsNil() {
			return true
		}
	}
	return false
}
