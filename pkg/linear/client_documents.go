package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// Document retrieves a single knowledge base document by ID.
//
// Returns:
//   - Document.ID: Document UUID (always populated)
//   - Document.Title: Document title (always populated)
//   - Document.Content: Document content in markdown (may be empty)
//   - Document.SlugID: URL-friendly identifier (always populated)
//   - Document.Icon: Document icon (may be empty)
//   - Document.Color: Document color hex code (may be empty)
//   - Document.CreatedAt: Creation timestamp (always populated)
//   - Document.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if document not found or query fails
//
// Permissions Required: Read
//
// Related: [Documents]
func (c *Client) Document(ctx context.Context, id string) (*intgraphql.GetDocument_Document, error) {
	resp, err := c.gqlClient.GetDocument(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("document query", err)
	}
	return &resp.Document, nil
}

// Documents retrieves a paginated list of knowledge base documents.
//
// Parameters:
//   - first: Number of documents to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Documents.Nodes: Array of documents (may be empty)
//   - Documents.PageInfo.HasNextPage: true if more results available
//   - Documents.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Document]
func (c *Client) Documents(ctx context.Context, first *int64, after *string) (*intgraphql.ListDocuments_Documents, error) {
	resp, err := c.gqlClient.ListDocuments(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("documents query", err)
	}
	return &resp.Documents, nil
}

// DocumentsFiltered retrieves documents with optional filtering.
//
// Filters:
//   - CreatedAt, UpdatedAt: Date range filters
//   - Creator: Filter by document creator
//   - ID: Filter by document UUID
//   - Initiative, Project, Issue: Filter by associated entities
//   - Title, SlugID: Filter by text fields
//
// Permissions Required: Read
//
// Related: [Documents], [Document]
func (c *Client) DocumentsFiltered(ctx context.Context, first *int64, after *string, filter *intgraphql.DocumentFilter) (*intgraphql.ListDocumentsFiltered_Documents, error) {
	resp, err := c.gqlClient.ListDocumentsFiltered(ctx, first, after, filter)
	if err != nil {
		return nil, wrapGraphQLError("documents filtered query", err)
	}
	return &resp.Documents, nil
}
