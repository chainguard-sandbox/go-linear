package label

import (
	"context"
	"reflect"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/dateparser"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
)

// FilterBuilder constructs IssueLabelFilter from CLI flags.
type FilterBuilder struct {
	filter   *intgraphql.IssueLabelFilter
	resolver *resolver.Resolver
	parser   *dateparser.Parser
}

// NewFilterBuilder creates a FilterBuilder with the given resolver.
func NewFilterBuilder(res *resolver.Resolver) *FilterBuilder {
	return &FilterBuilder{
		filter:   &intgraphql.IssueLabelFilter{},
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

// Filter returns the underlying IssueLabelFilter.
func (b *FilterBuilder) Filter() *intgraphql.IssueLabelFilter {
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

// Build returns the filter if any fields are set, nil otherwise.
func (b *FilterBuilder) Build() *intgraphql.IssueLabelFilter {
	if !hasAnyFilter(
		b.filter.ID,
		b.filter.Name,
		b.filter.CreatedAt,
		b.filter.UpdatedAt,
		b.filter.Creator,
		b.filter.Team,
		b.filter.IsGroup,
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
