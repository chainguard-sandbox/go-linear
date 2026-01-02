// Package document provides filtering for Linear documents.
package document

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

// FilterBuilder constructs DocumentFilter from command flags.
type FilterBuilder struct {
	filter   *intgraphql.DocumentFilter
	resolver *resolver.Resolver
	parser   *dateparser.Parser
}

// NewFilterBuilder creates a new FilterBuilder.
func NewFilterBuilder(res *resolver.Resolver) *FilterBuilder {
	return &FilterBuilder{
		filter:   &intgraphql.DocumentFilter{},
		resolver: res,
		parser:   dateparser.New(),
	}
}

// Build returns the filter if any fields are set, otherwise nil.
func (b *FilterBuilder) Build() *intgraphql.DocumentFilter {
	if !hasAnyFilter(
		b.filter.ID,
		b.filter.CreatedAt,
		b.filter.UpdatedAt,
		b.filter.Creator,
		b.filter.Initiative,
		b.filter.Project,
		b.filter.Issue,
		b.filter.Title,
		b.filter.SlugID,
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
func (b *FilterBuilder) Filter() *intgraphql.DocumentFilter {
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

// UpdatedAtComparator returns the UpdatedAt comparator, creating if needed.
func (b *FilterBuilder) UpdatedAtComparator() *intgraphql.DateComparator {
	if b.filter.UpdatedAt == nil {
		b.filter.UpdatedAt = &intgraphql.DateComparator{}
	}
	return b.filter.UpdatedAt
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
