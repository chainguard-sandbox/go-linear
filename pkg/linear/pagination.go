package linear

import (
	"context"
	"io"
	"sync"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// IssueIterator provides automatic pagination over issues.
// Safe for concurrent use - protected by internal mutex.
type IssueIterator struct {
	mu      sync.Mutex
	client  *Client
	first   int64
	buffer  []*intgraphql.ListIssues_Issues_Nodes
	cursor  *string
	index   int
	hasMore bool
}

// NewIssueIterator creates an iterator for automatic pagination through all issues.
//
// Parameters:
//   - client: Linear client instance (required)
//   - pageSize: Number of issues per API request (0 = default 50)
//
// Usage Pattern:
//
//	iter := linear.NewIssueIterator(client, 100)
//	for {
//	    issue, err := iter.Next(ctx)
//	    if err == io.EOF {
//	        break  // Done
//	    }
//	    if err != nil {
//	        return fmt.Errorf("iteration failed: %w", err)
//	    }
//	    fmt.Printf("%s: %s\n", issue.ID, issue.Title)
//	}
//
// Thread Safety: Safe for concurrent use. Returns values instead of mutating state.
//
// Permissions Required: Read
//
// Related: [Issues], [NewTeamIterator], [NewProjectIterator]
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

// Next returns the next issue or an error.
// Returns io.EOF when no more issues are available.
//
// Safe for concurrent use - protected by internal mutex.
func (it *IssueIterator) Next(ctx context.Context) (*intgraphql.ListIssues_Issues_Nodes, error) {
	it.mu.Lock()
	defer it.mu.Unlock()

	// Return buffered item
	if it.index < len(it.buffer) {
		issue := it.buffer[it.index]
		it.index++
		return issue, nil
	}

	// No more pages
	if !it.hasMore {
		return nil, io.EOF
	}

	// Fetch next page
	resp, err := it.client.Issues(ctx, &it.first, it.cursor)
	if err != nil {
		return nil, err
	}

	it.buffer = resp.Nodes
	it.index = 0
	it.hasMore = resp.PageInfo.HasNextPage
	it.cursor = resp.PageInfo.EndCursor

	// Return first item from new page
	if len(it.buffer) == 0 {
		return nil, io.EOF
	}

	it.index = 1
	return it.buffer[0], nil
}

// TeamIterator provides automatic pagination over teams.
// Safe for concurrent use - protected by internal mutex.
type TeamIterator struct {
	mu      sync.Mutex
	client  *Client
	first   int64
	buffer  []*intgraphql.ListTeams_Teams_Nodes
	cursor  *string
	index   int
	hasMore bool
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

// Next returns the next team or an error.
// Returns io.EOF when no more teams are available.
//
// Safe for concurrent use - protected by internal mutex.
func (it *TeamIterator) Next(ctx context.Context) (*intgraphql.ListTeams_Teams_Nodes, error) {
	it.mu.Lock()
	defer it.mu.Unlock()

	if it.index < len(it.buffer) {
		team := it.buffer[it.index]
		it.index++
		return team, nil
	}

	if !it.hasMore {
		return nil, io.EOF
	}

	resp, err := it.client.Teams(ctx, &it.first, it.cursor)
	if err != nil {
		return nil, err
	}

	it.buffer = resp.Nodes
	it.index = 0
	it.hasMore = resp.PageInfo.HasNextPage
	it.cursor = resp.PageInfo.EndCursor

	if len(it.buffer) == 0 {
		return nil, io.EOF
	}

	it.index = 1
	return it.buffer[0], nil
}

// ProjectIterator provides automatic pagination over projects.
// Safe for concurrent use - protected by internal mutex.
type ProjectIterator struct {
	mu      sync.Mutex
	client  *Client
	first   int64
	buffer  []*intgraphql.ListProjects_Projects_Nodes
	cursor  *string
	index   int
	hasMore bool
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

// Next returns the next project or an error.
// Returns io.EOF when no more projects are available.
//
// Safe for concurrent use - protected by internal mutex.
func (it *ProjectIterator) Next(ctx context.Context) (*intgraphql.ListProjects_Projects_Nodes, error) {
	it.mu.Lock()
	defer it.mu.Unlock()

	if it.index < len(it.buffer) {
		project := it.buffer[it.index]
		it.index++
		return project, nil
	}

	if !it.hasMore {
		return nil, io.EOF
	}

	resp, err := it.client.Projects(ctx, &it.first, it.cursor)
	if err != nil {
		return nil, err
	}

	it.buffer = resp.Nodes
	it.index = 0
	it.hasMore = resp.PageInfo.HasNextPage
	it.cursor = resp.PageInfo.EndCursor

	if len(it.buffer) == 0 {
		return nil, io.EOF
	}

	it.index = 1
	return it.buffer[0], nil
}

// CommentIterator provides automatic pagination over comments.
// Safe for concurrent use - protected by internal mutex.
type CommentIterator struct {
	mu      sync.Mutex
	client  *Client
	first   int64
	buffer  []*intgraphql.ListComments_Comments_Nodes
	cursor  *string
	index   int
	hasMore bool
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

// Next returns the next comment or an error.
// Returns io.EOF when no more comments are available.
//
// Safe for concurrent use - protected by internal mutex.
func (it *CommentIterator) Next(ctx context.Context) (*intgraphql.ListComments_Comments_Nodes, error) {
	it.mu.Lock()
	defer it.mu.Unlock()

	if it.index < len(it.buffer) {
		comment := it.buffer[it.index]
		it.index++
		return comment, nil
	}

	if !it.hasMore {
		return nil, io.EOF
	}

	resp, err := it.client.Comments(ctx, &it.first, it.cursor)
	if err != nil {
		return nil, err
	}

	it.buffer = resp.Nodes
	it.index = 0
	it.hasMore = resp.PageInfo.HasNextPage
	it.cursor = resp.PageInfo.EndCursor

	if len(it.buffer) == 0 {
		return nil, io.EOF
	}

	it.index = 1
	return it.buffer[0], nil
}
