package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// Notifications retrieves a paginated list of notifications (inbox).
//
// Parameters:
//   - first: Number of notifications to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//   - filter: Optional filters (includeArchived, etc.)
//
// Returns:
//   - Notifications.Nodes: Array of notifications (may be empty)
//   - Notifications.PageInfo.HasNextPage: true if more results available
//   - error: Non-nil if query fails
//
// Permissions Required: Read
//
// Related: [Notification], [NotificationArchive], [NotificationUpdate]
func (c *Client) Notifications(ctx context.Context, first *int64, after *string, filter *intgraphql.NotificationFilter) (*intgraphql.ListNotifications_Notifications, error) {
	resp, err := c.gqlClient.ListNotifications(ctx, first, after, filter)
	if err != nil {
		return nil, wrapGraphQLError("notifications query", err)
	}
	return &resp.Notifications, nil
}

// Notification retrieves a single notification by ID.
//
// Returns:
//   - Notification with ID, type, dates, and user
//   - error: Non-nil if notification not found or query fails
//
// Permissions Required: Read
//
// Related: [Notifications], [NotificationArchive]
func (c *Client) Notification(ctx context.Context, id string) (*intgraphql.GetNotification_Notification, error) {
	resp, err := c.gqlClient.GetNotification(ctx, id)
	if err != nil {
		return nil, wrapGraphQLError("notification query", err)
	}
	return &resp.Notification, nil
}

// NotificationUpdate marks a notification as read or archives it.
//
// Parameters:
//   - id: Notification UUID to update (required)
//   - input: Fields to update (readAt, archivedAt, snoozedUntilAt)
//
// Returns:
//   - Updated notification with new status
//   - error: Non-nil if update fails or Success is false
//
// Permissions Required: Write
//
// Example (mark as read):
//
//	now := time.Now()
//	notification, err := client.NotificationUpdate(ctx, notificationID, intgraphql.NotificationUpdateInput{
//	    ReadAt: &now,
//	})
//
// Related: [NotificationArchive]
func (c *Client) NotificationUpdate(ctx context.Context, id string, input intgraphql.NotificationUpdateInput) (*intgraphql.NotificationUpdate_NotificationUpdate_Notification, error) {
	resp, err := c.gqlClient.NotificationUpdate(ctx, id, input)
	if err != nil {
		return nil, wrapGraphQLError("NotificationUpdate", err)
	}
	if !resp.NotificationUpdate.Success {
		return nil, errMutationFailed("NotificationUpdate")
	}
	return &resp.NotificationUpdate.Notification, nil
}

// NotificationArchive archives a notification.
//
// Parameters:
//   - id: Notification UUID to archive (required)
//
// Returns:
//   - nil: Notification successfully archived
//   - error: Non-nil if archive fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	err := client.NotificationArchive(ctx, notificationID)
//
// Related: [NotificationUpdate], [NotificationSubscriptionCreate]
func (c *Client) NotificationArchive(ctx context.Context, id string) error {
	resp, err := c.gqlClient.NotificationArchive(ctx, id)
	if err != nil {
		return wrapGraphQLError("NotificationArchive", err)
	}
	if !resp.NotificationArchive.Success {
		return errMutationFailed("NotificationArchive")
	}
	return nil
}

// NotificationSubscriptionCreate subscribes to notifications for an issue or project.
//
// Parameters:
//   - input: Subscription parameters (issueId OR projectId)
//
// Returns:
//   - Created subscription with ID and active status
//   - error: Non-nil if creation fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	subscription, err := client.NotificationSubscriptionCreate(ctx, intgraphql.NotificationSubscriptionCreateInput{
//	    IssueID: &issueID,
//	})
//
// Related: [NotificationSubscriptionDelete]
func (c *Client) NotificationSubscriptionCreate(ctx context.Context, input intgraphql.NotificationSubscriptionCreateInput) (*intgraphql.NotificationSubscriptionCreate_NotificationSubscriptionCreate_NotificationSubscription, error) {
	resp, err := c.gqlClient.NotificationSubscriptionCreate(ctx, input)
	if err != nil {
		return nil, wrapGraphQLError("NotificationSubscriptionCreate", err)
	}
	if !resp.NotificationSubscriptionCreate.Success {
		return nil, errMutationFailed("NotificationSubscriptionCreate")
	}
	return &resp.NotificationSubscriptionCreate.NotificationSubscription, nil
}

// NotificationSubscriptionDelete unsubscribes from notifications.
//
// Parameters:
//   - id: Subscription UUID to delete (required)
//
// Returns:
//   - nil: Subscription successfully deleted
//   - error: Non-nil if delete fails or Success is false
//
// Permissions Required: Write
//
// Example:
//
//	err := client.NotificationSubscriptionDelete(ctx, subscriptionID)
//
// Related: [NotificationSubscriptionCreate]
func (c *Client) NotificationSubscriptionDelete(ctx context.Context, id string) error {
	resp, err := c.gqlClient.NotificationSubscriptionDelete(ctx, id)
	if err != nil {
		return wrapGraphQLError("NotificationSubscriptionDelete", err)
	}
	if !resp.NotificationSubscriptionDelete.Success {
		return errMutationFailed("NotificationSubscriptionDelete")
	}
	return nil
}

// NotificationUnarchive restores an archived notification to inbox.
//
// Parameters:
//   - id: Notification UUID to unarchive (required)
//
// Returns:
//   - nil: Notification successfully unarchived
//   - error: Non-nil if unarchive fails or Success is false
//
// Permissions Required: Write
//
// Related: [NotificationArchive], [Notifications]
func (c *Client) NotificationUnarchive(ctx context.Context, id string) error {
	resp, err := c.gqlClient.UnarchiveNotification(ctx, id)
	if err != nil {
		return wrapGraphQLError("NotificationUnarchive", err)
	}
	if !resp.NotificationUnarchive.Success {
		return errMutationFailed("NotificationUnarchive")
	}
	return nil
}
