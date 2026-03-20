package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// Attachment retrieves a single file attachment by ID.
//
// Returns:
//   - Attachment.ID: Attachment UUID (always populated)
//   - Attachment.Title: Attachment title (always populated)
//   - Attachment.Subtitle: Attachment subtitle (may be empty)
//   - Attachment.URL: File URL (always populated)
//   - Attachment.Source: External source (may be nil)
//   - Attachment.SourceType: Source type (may be nil)
//   - Attachment.Metadata: Additional metadata (always populated, may be empty object)
//   - Attachment.CreatedAt: Upload timestamp (always populated)
//   - Attachment.UpdatedAt: Last update timestamp (always populated)
//   - error: Non-nil if attachment not found or query fails
//
// Permissions Required: Read
//
// Related: [Attachments]
func (c *Client) Attachment(ctx context.Context, id string) (*intgraphql.GetAttachment_Attachment, error) {
	resp, err := c.gqlClient.GetAttachment(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("attachment query", err)
	}
	return &resp.Attachment, nil
}

// Attachments retrieves a paginated list of file attachments.
//
// Parameters:
//   - first: Number of attachments to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//
// Returns:
//   - Attachments.Nodes: Array of attachments (may be empty)
//   - Attachments.PageInfo.HasNextPage: true if more results available
//   - Attachments.PageInfo.EndCursor: Cursor for next page
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Attachment]
func (c *Client) Attachments(ctx context.Context, first *int64, after *string) (*intgraphql.ListAttachments_Attachments, error) {
	resp, err := c.gqlClient.ListAttachments(ctx, first, after)
	if err != nil {
		return nil, wrapGraphQLError("attachments query", err)
	}
	return &resp.Attachments, nil
}

// AttachmentsFiltered returns attachments matching the specified filter.
func (c *Client) AttachmentsFiltered(ctx context.Context, first *int64, after *string, filter *intgraphql.AttachmentFilter) (*intgraphql.ListAttachmentsFiltered_Attachments, error) {
	resp, err := c.gqlClient.ListAttachmentsFiltered(ctx, first, after, filter)
	if err != nil {
		return nil, wrapGraphQLError("attachments filtered query", err)
	}
	return &resp.Attachments, nil
}

// AttachmentLinkURL attaches an external URL to an issue.
//
// Parameters:
//   - issueID: Issue UUID to attach to (required)
//   - url: URL to link (required)
//   - title: Display title for the attachment (nil = use URL as title)
//
// Returns:
//   - Created attachment with ID, title, and URL
//   - error: Non-nil if operation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	title := "API Documentation"
//	attachment, err := client.AttachmentLinkURL(ctx, issueID,
//	    "https://docs.example.com/api",
//	    &title)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Attached: %s\n", attachment.URL)
//
// Related: [AttachmentLinkGitHubPR], [Attachments]
func (c *Client) AttachmentLinkURL(ctx context.Context, issueID, url string, title *string) (*intgraphql.AttachmentLinkURL_AttachmentLinkURL_Attachment, error) {
	resp, err := c.gqlClient.AttachmentLinkURL(ctx, issueID, url, title)
	if err != nil {
		return nil, wrapGraphQLError("AttachmentLinkURL", err)
	}
	if !resp.AttachmentLinkURL.Success {
		return nil, errMutationFailed("AttachmentLinkURL")
	}
	return &resp.AttachmentLinkURL.Attachment, nil
}

// AttachmentLinkGitHubPR links a GitHub Pull Request to an issue.
//
// The attachment title is automatically extracted from the PR metadata.
// Requires GitHub integration to be configured in Linear workspace.
//
// Parameters:
//   - issueID: Issue UUID to attach to (required)
//   - url: GitHub PR URL (required, e.g., "https://github.com/owner/repo/pull/123")
//
// Returns:
//   - Created attachment with ID, title (from PR), and URL
//   - error: Non-nil if operation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	attachment, err := client.AttachmentLinkGitHubPR(ctx, issueID,
//	    "https://github.com/myorg/myrepo/pull/456")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Linked PR: %s\n", attachment.Title)
//
// Related: [AttachmentLinkURL], [Attachments]
func (c *Client) AttachmentLinkGitHubPR(ctx context.Context, issueID, url string) (*intgraphql.AttachmentLinkGitHubPR_AttachmentLinkGitHubPr_Attachment, error) {
	resp, err := c.gqlClient.AttachmentLinkGitHubPr(ctx, issueID, url)
	if err != nil {
		return nil, wrapGraphQLError("AttachmentLinkGitHubPR", err)
	}
	if !resp.AttachmentLinkGitHubPr.Success {
		return nil, errMutationFailed("AttachmentLinkGitHubPR")
	}
	return &resp.AttachmentLinkGitHubPr.Attachment, nil
}

// AttachmentCreate creates a custom attachment for an issue.
//
// This method creates attachments with custom metadata, useful for integrating
// external tools or displaying custom data cards in Linear.
//
// Parameters:
//   - input: Attachment creation parameters (title, issueId, url, metadata, etc.)
//
// Required fields:
//   - IssueID: UUID of the issue to attach to
//   - Title: Attachment title
//
// Optional fields:
//   - URL: External URL for the attachment
//   - Subtitle: Additional context
//   - IconURL: Custom icon (20x20px recommended)
//   - Metadata: JSON object with custom data
//
// Returns:
//   - Created attachment with ID, title, URL, and metadata
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	attachment, err := client.AttachmentCreate(ctx, intgraphql.AttachmentCreateInput{
//	    IssueID: issueID,
//	    Title:   "Build Status",
//	    URL:     stringPtr("https://ci.example.com/builds/123"),
//	    Subtitle: stringPtr("Build #123 - Passed"),
//	})
//
// Related: [AttachmentDelete], [AttachmentLinkURL]
func (c *Client) AttachmentCreate(ctx context.Context, input intgraphql.AttachmentCreateInput) (*intgraphql.AttachmentCreate_AttachmentCreate_Attachment, error) {
	resp, err := c.gqlClient.AttachmentCreate(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("AttachmentCreate", err)
	}
	if !resp.AttachmentCreate.Success {
		return nil, errMutationFailed("AttachmentCreate")
	}
	return &resp.AttachmentCreate.Attachment, nil
}

// AttachmentLinkSlack links a Slack thread to an issue.
//
// Requires Slack integration to be configured in Linear workspace.
//
// Parameters:
//   - issueID: Issue UUID to attach to (required)
//   - url: Slack message URL (required, e.g., "https://workspace.slack.com/archives/C123/p1234567890")
//
// Returns:
//   - Created attachment with ID, title, and URL
//   - error: Non-nil if operation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	attachment, err := client.AttachmentLinkSlack(ctx, issueID,
//	    "https://myworkspace.slack.com/archives/C123ABC/p1234567890123456")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Linked Slack thread: %s\n", attachment.URL)
//
// Related: [AttachmentLinkURL], [Attachments]
func (c *Client) AttachmentLinkSlack(ctx context.Context, issueID, url string) (*intgraphql.AttachmentLinkSlack_AttachmentLinkSlack_Attachment, error) {
	resp, err := c.gqlClient.AttachmentLinkSlack(ctx, issueID, url)
	if err != nil {
		return nil, wrapGraphQLError("AttachmentLinkSlack", err)
	}
	if !resp.AttachmentLinkSlack.Success {
		return nil, errMutationFailed("AttachmentLinkSlack")
	}
	return &resp.AttachmentLinkSlack.Attachment, nil
}

// AttachmentDelete permanently deletes an attachment.
//
// Parameters:
//   - id: Attachment UUID to delete (required)
//
// Returns:
//   - nil: Attachment successfully deleted
//   - error: Non-nil if delete fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	err := client.AttachmentDelete(ctx, attachmentID)
//	if err != nil {
//	    return fmt.Errorf("failed to delete attachment: %w", err)
//	}
//
// Related: [AttachmentCreate], [AttachmentLinkURL]
func (c *Client) AttachmentDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.AttachmentDelete(ctx, id)
	if err != nil {
		return wrapGraphQLError("AttachmentDelete", err)
	}
	if !resp.AttachmentDelete.Success {
		return errMutationFailed("AttachmentDelete")
	}
	return nil
}
