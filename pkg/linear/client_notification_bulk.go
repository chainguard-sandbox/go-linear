package linear

import (
	"context"
	"time"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// NotificationArchiveAll archives all notifications for a given entity.
//
// Parameters:
//   - input: Entity to archive notifications for (issueId, projectId, etc.)
//
// Returns:
//   - nil: Notifications successfully archived
//   - error: Non-nil if operation fails
//
// Permissions Required: Write
//
// Related: [NotificationArchive], [NotificationMarkReadAll]
func (c *Client) NotificationArchiveAll(ctx context.Context, input intgraphql.NotificationEntityInput) error {
	resp, err := c.gqlClient.NotificationArchiveAll(ctx, input)
	if err != nil {
		return wrapGraphQLError("NotificationArchiveAll", err)
	}
	if !resp.NotificationArchiveAll.Success {
		return errMutationFailed("NotificationArchiveAll")
	}
	return nil
}

// NotificationMarkReadAll marks all notifications as read for a given entity.
//
// Parameters:
//   - input: Entity to mark notifications for
//   - readAt: Time to set as read timestamp
//
// Returns:
//   - nil: Notifications successfully marked as read
//   - error: Non-nil if operation fails
//
// Permissions Required: Write
//
// Related: [NotificationMarkUnreadAll], [NotificationUpdate]
func (c *Client) NotificationMarkReadAll(ctx context.Context, input intgraphql.NotificationEntityInput, readAt time.Time) error {
	resp, err := c.gqlClient.NotificationMarkReadAll(ctx, input, readAt)
	if err != nil {
		return wrapGraphQLError("NotificationMarkReadAll", err)
	}
	if !resp.NotificationMarkReadAll.Success {
		return errMutationFailed("NotificationMarkReadAll")
	}
	return nil
}

// NotificationMarkUnreadAll marks all notifications as unread for a given entity.
//
// Parameters:
//   - input: Entity to mark notifications for
//
// Returns:
//   - nil: Notifications successfully marked as unread
//   - error: Non-nil if operation fails
//
// Permissions Required: Write
//
// Related: [NotificationMarkReadAll], [NotificationUpdate]
func (c *Client) NotificationMarkUnreadAll(ctx context.Context, input intgraphql.NotificationEntityInput) error {
	resp, err := c.gqlClient.NotificationMarkUnreadAll(ctx, input)
	if err != nil {
		return wrapGraphQLError("NotificationMarkUnreadAll", err)
	}
	if !resp.NotificationMarkUnreadAll.Success {
		return errMutationFailed("NotificationMarkUnreadAll")
	}
	return nil
}

// NotificationSnoozeAll snoozes all notifications for a given entity.
//
// Parameters:
//   - input: Entity to snooze notifications for
//   - snoozedUntilAt: Time until notifications are snoozed
//
// Returns:
//   - nil: Notifications successfully snoozed
//   - error: Non-nil if operation fails
//
// Permissions Required: Write
//
// Related: [NotificationUnsnoozeAll], [NotificationUpdate]
func (c *Client) NotificationSnoozeAll(ctx context.Context, input intgraphql.NotificationEntityInput, snoozedUntilAt time.Time) error {
	resp, err := c.gqlClient.NotificationSnoozeAll(ctx, input, snoozedUntilAt)
	if err != nil {
		return wrapGraphQLError("NotificationSnoozeAll", err)
	}
	if !resp.NotificationSnoozeAll.Success {
		return errMutationFailed("NotificationSnoozeAll")
	}
	return nil
}

// NotificationUnsnoozeAll unsnoozes all notifications for a given entity.
//
// Parameters:
//   - input: Entity to unsnooze notifications for
//   - unsnoozedAt: Time when the notification was unsnoozed
//
// Returns:
//   - nil: Notifications successfully unsnoozed
//   - error: Non-nil if operation fails
//
// Permissions Required: Write
//
// Related: [NotificationSnoozeAll], [NotificationUpdate]
func (c *Client) NotificationUnsnoozeAll(ctx context.Context, input intgraphql.NotificationEntityInput, unsnoozedAt time.Time) error {
	resp, err := c.gqlClient.NotificationUnsnoozeAll(ctx, input, unsnoozedAt)
	if err != nil {
		return wrapGraphQLError("NotificationUnsnoozeAll", err)
	}
	if !resp.NotificationUnsnoozeAll.Success {
		return errMutationFailed("NotificationUnsnoozeAll")
	}
	return nil
}
