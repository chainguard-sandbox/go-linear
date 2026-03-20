package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
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

// DocumentCreate creates a new knowledge base document.
//
// Parameters:
//   - input: Document creation parameters
//
// Required fields:
//   - Title: Document title (required)
//
// Optional fields:
//   - Content: Document content in markdown
//   - ProjectID: Link to project
//   - InitiativeID: Link to initiative
//   - Icon: Document icon
//   - Color: Icon color
//
// Returns:
//   - Created document with ID, title, content, and metadata
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Related: [DocumentUpdate], [DocumentDelete]
func (c *Client) DocumentCreate(ctx context.Context, input intgraphql.DocumentCreateInput) (*intgraphql.CreateDocument_DocumentCreate_Document, error) {
	resp, err := c.gqlClient.CreateDocument(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("DocumentCreate", err)
	}
	if !resp.DocumentCreate.Success {
		return nil, errMutationFailed("DocumentCreate")
	}
	return &resp.DocumentCreate.Document, nil
}

// DocumentUpdate updates an existing knowledge base document.
//
// Parameters:
//   - id: Document UUID to update (required)
//   - input: Fields to update (all optional)
//
// Optional fields:
//   - Title: Update document title
//   - Content: Update document content
//   - ProjectID: Update project link
//   - InitiativeID: Update initiative link
//
// Returns:
//   - Updated document with new values
//   - error: Non-nil if update fails or Success is false
//
// Permissions Required: Write
//
// Related: [DocumentCreate], [DocumentDelete]
func (c *Client) DocumentUpdate(ctx context.Context, id string, input intgraphql.DocumentUpdateInput) (*intgraphql.UpdateDocument_DocumentUpdate_Document, error) {
	resp, err := c.gqlClient.UpdateDocument(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("DocumentUpdate", err)
	}
	if !resp.DocumentUpdate.Success {
		return nil, errMutationFailed("DocumentUpdate")
	}
	return &resp.DocumentUpdate.Document, nil
}

// DocumentDelete deletes a knowledge base document.
//
// Parameters:
//   - id: Document UUID to delete (required)
//
// Returns:
//   - nil: Document successfully deleted
//   - error: Non-nil if delete fails or Success is false
//
// Permissions Required: Write
//
// Related: [DocumentCreate], [DocumentUpdate]
func (c *Client) DocumentDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.DeleteDocument(ctx, id)
	if err != nil {
		return wrapGraphQLError("DocumentDelete", err)
	}
	if !resp.DocumentDelete.Success {
		return errMutationFailed("DocumentDelete")
	}
	return nil
}

// DocumentUnarchive restores a deleted document.
//
// Parameters:
//   - id: Document UUID to restore (required)
//
// Returns:
//   - nil: Document successfully restored
//   - error: Non-nil if unarchive fails or Success is false
//
// Permissions Required: Write
//
// Related: [DocumentDelete], [DocumentCreate]
func (c *Client) DocumentUnarchive(ctx context.Context, id string) error {
	resp, err := c.gqlClient.UnarchiveDocument(ctx, id)
	if err != nil {
		return wrapGraphQLError("DocumentUnarchive", err)
	}
	if !resp.DocumentUnarchive.Success {
		return errMutationFailed("DocumentUnarchive")
	}
	return nil
}
