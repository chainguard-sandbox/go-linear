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

// NewIssueIterator creates an iterator for automatic pagination through all issues.
//
// Parameters:
//   - client: Linear client instance (required)
//   - pageSize: Number of issues per API request (0 = default 50)
//
// Usage Pattern:
//  1. Create iterator with NewIssueIterator
//  2. Call Next(ctx) in loop - returns true while issues remain
//  3. Call Issue() to get current issue
//  4. Check Err() after loop completes
//
// Iterator automatically:
//   - Fetches pages as needed
//   - Handles cursor management
//   - Stops when no more results
//   - Captures errors for later check
//
// Permissions Required: Read
//
// Related: [Issues], [NewTeamIterator], [NewProjectIterator]
//
// Example:
//
//	iter := linear.NewIssueIterator(client, 100)
//	for iter.Next(ctx) {
//	    issue := iter.Issue()
//	    fmt.Printf("%s: %s\n", issue.ID, issue.Title)
//	}
//	if err := iter.Err(); err != nil {
//	    return fmt.Errorf("iteration failed: %w", err)
//	}
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

// TeamIterator provides automatic pagination over teams.
type TeamIterator struct {
	client  *Client
	first   int64
	buffer  []*intgraphql.ListTeams_Teams_Nodes
	cursor  *string
	index   int
	hasMore bool
	err     error
}

// NewTeamIterator creates an iterator for paginating through all teams.
func NewTeamIterator(client *Client, pageSize int64) *TeamIterator {
	if pageSize <= 0 {
		pageSize = 50
	}
	return &TeamIterator{
		client:  client,
		first:   pageSize,
		hasMore: true,
	}
}

// Next advances to the next team.
func (it *TeamIterator) Next(ctx context.Context) bool {
	if it.err != nil {
		return false
	}

	if it.index < len(it.buffer) {
		it.index++
		return true
	}

	if !it.hasMore {
		return false
	}

	resp, err := it.client.Teams(ctx, &it.first, it.cursor)
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

// Team returns the current team.
func (it *TeamIterator) Team() *intgraphql.ListTeams_Teams_Nodes {
	if it.index == 0 || it.index > len(it.buffer) {
		return nil
	}
	return it.buffer[it.index-1]
}

// Err returns any error encountered during iteration.
func (it *TeamIterator) Err() error {
	return it.err
}

// ProjectIterator provides automatic pagination over projects.
type ProjectIterator struct {
	client  *Client
	first   int64
	buffer  []*intgraphql.ListProjects_Projects_Nodes
	cursor  *string
	index   int
	hasMore bool
	err     error
}

// NewProjectIterator creates an iterator for paginating through all projects.
func NewProjectIterator(client *Client, pageSize int64) *ProjectIterator {
	if pageSize <= 0 {
		pageSize = 50
	}
	return &ProjectIterator{
		client:  client,
		first:   pageSize,
		hasMore: true,
	}
}

// Next advances to the next project.
func (it *ProjectIterator) Next(ctx context.Context) bool {
	if it.err != nil {
		return false
	}

	if it.index < len(it.buffer) {
		it.index++
		return true
	}

	if !it.hasMore {
		return false
	}

	resp, err := it.client.Projects(ctx, &it.first, it.cursor)
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

// Project returns the current project.
func (it *ProjectIterator) Project() *intgraphql.ListProjects_Projects_Nodes {
	if it.index == 0 || it.index > len(it.buffer) {
		return nil
	}
	return it.buffer[it.index-1]
}

// Err returns any error encountered during iteration.
func (it *ProjectIterator) Err() error {
	return it.err
}

// CommentIterator provides automatic pagination over comments.
type CommentIterator struct {
	client  *Client
	first   int64
	buffer  []*intgraphql.ListComments_Comments_Nodes
	cursor  *string
	index   int
	hasMore bool
	err     error
}

// NewCommentIterator creates an iterator for paginating through all comments.
func NewCommentIterator(client *Client, pageSize int64) *CommentIterator {
	if pageSize <= 0 {
		pageSize = 50
	}
	return &CommentIterator{
		client:  client,
		first:   pageSize,
		hasMore: true,
	}
}

// Next advances to the next comment.
func (it *CommentIterator) Next(ctx context.Context) bool {
	if it.err != nil {
		return false
	}

	if it.index < len(it.buffer) {
		it.index++
		return true
	}

	if !it.hasMore {
		return false
	}

	resp, err := it.client.Comments(ctx, &it.first, it.cursor)
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

// Comment returns the current comment.
func (it *CommentIterator) Comment() *intgraphql.ListComments_Comments_Nodes {
	if it.index == 0 || it.index > len(it.buffer) {
		return nil
	}
	return it.buffer[it.index-1]
}

// Err returns any error encountered during iteration.
func (it *CommentIterator) Err() error {
	return it.err
}
