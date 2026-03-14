// Package filter provides utilities to build GraphQL filters from CLI flags.
package filter

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/dateparser"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
)

// DateFormat is the ISO8601 format used for Linear API date fields.
const DateFormat = "2006-01-02T15:04:05.000Z"

// FilterBuilder is the base interface for all entity filter builders.
// Each entity (Issue, Project, Cycle) implements this interface.
type FilterBuilder interface {
	// Resolver returns the resolver for name-to-ID lookups.
	Resolver() *resolver.Resolver

	// Parser returns the date parser for human-readable dates.
	Parser() dateparser.Parser
}

// DateRange holds pointers to Gte and Lt fields for a date comparator.
// This allows generic date filter functions to work with both
// DateComparator and NullableDateComparator.
type DateRange struct {
	Gte *string
	Lt  *string
}

// FilterFunc is a function that applies a filter from command flags.
// Each filter implementation follows this signature.
type FilterFunc[T FilterBuilder] func(ctx context.Context, cmd *cobra.Command, builder T) error

// ApplyDateRange parses date strings and applies them to a DateRange.
// It handles the common pattern of --*-after and --*-before flags.
func ApplyDateRange(parser dateparser.Parser, dr *DateRange, afterFlag, beforeFlag, fieldName string) error {
	if afterFlag != "" {
		t, err := parser.Parse(afterFlag)
		if err != nil {
			return fmt.Errorf("invalid %s-after date: %w", fieldName, err)
		}
		tStr := t.Format(DateFormat)
		dr.Gte = &tStr
	}
	if beforeFlag != "" {
		t, err := parser.Parse(beforeFlag)
		if err != nil {
			return fmt.Errorf("invalid %s-before date: %w", fieldName, err)
		}
		tStr := t.Format(DateFormat)
		dr.Lt = &tStr
	}
	return nil
}

