package linear

import (
	"context"

	intgraphql "github.com/eslerm/go-linear/internal/graphql"
)

// IssueIterator provides automatic pagination over issues.
type IssueIterator struct {
	client  *Client
	first   int64
	buffer  []*intgraphql.ListIssues_Issues_Nodes
	cursor  *string
	index   int
	hasMore bool
	err     error
}

// NewIssueIterator creates an iterator for paginating through all issues.
func NewIssueIterator(client *Client, pageSize int64) *IssueIterator {
	if pageSize <= 0 {
		pageSize = 50
	}
	return &IssueIterator{
		client:  client,
		first:   pageSize,
		hasMore: true,
	}
}

// Next advances to the next issue.
func (it *IssueIterator) Next(ctx context.Context) bool {
	if it.err != nil {
		return false
	}

	// Return buffered item
	if it.index < len(it.buffer) {
		it.index++
		return true
	}

	// No more pages
	if !it.hasMore {
		return false
	}

	// Fetch next page
	resp, err := it.client.Issues(ctx, &it.first, it.cursor)
	if err != nil {
		it.err = err
		return false
	}

	it.buffer = resp.Nodes
	it.index = 1
	it.hasMore = resp.PageInfo.HasNextPage
	it.cursor = resp.PageInfo.EndCursor

	return len(it.buffer) > 0
}

// Issue returns the current issue.
func (it *IssueIterator) Issue() *intgraphql.ListIssues_Issues_Nodes {
	if it.index == 0 || it.index > len(it.buffer) {
		return nil
	}
	return it.buffer[it.index-1]
}

// Err returns any error encountered during iteration.
func (it *IssueIterator) Err() error {
	return it.err
}
