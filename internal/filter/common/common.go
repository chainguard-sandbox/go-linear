// Package common provides shared filter implementations using Go generics.
package common

import (
	"github.com/chainguard-sandbox/go-linear/v2/internal/dateparser"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// DateFilterable is implemented by filter builders that support CreatedAt filtering.
type DateFilterable interface {
	CreatedAtComparator() *intgraphql.DateComparator
	Parser() dateparser.Parser
}

// UpdateDateFilterable extends DateFilterable with UpdatedAt support.
type UpdateDateFilterable interface {
	DateFilterable
	UpdatedAtComparator() *intgraphql.DateComparator
}

// IDFilterable is implemented by filter builders that support ID filtering.
type IDFilterable interface {
	SetID(*intgraphql.IDComparator)
}
