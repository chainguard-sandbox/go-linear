package linear

import (
	"context"
)

// IssueSubscribe subscribes a user to an issue's notifications.
//
// Parameters:
//   - id: Issue UUID to subscribe to (required)
//   - userID: User UUID to subscribe (nil = current user)
//
// Returns:
//   - nil: Successfully subscribed
//   - error: Non-nil if subscription fails or Success is false
//
// Permissions Required: Write
//
// Related: [IssueUnsubscribe]
func (c *Client) IssueSubscribe(ctx context.Context, id string, userID *string) error {
	resp, err := c.gqlClient.IssueSubscribe(ctx, id, userID)
	if err != nil {
		return wrapGraphQLError("IssueSubscribe", err)
	}
	if !resp.IssueSubscribe.Success {
		return errMutationFailed("IssueSubscribe")
	}
	return nil
}

// IssueUnsubscribe unsubscribes a user from an issue's notifications.
//
// Parameters:
//   - id: Issue UUID to unsubscribe from (required)
//   - userID: User UUID to unsubscribe (nil = current user)
//
// Returns:
//   - nil: Successfully unsubscribed
//   - error: Non-nil if unsubscription fails or Success is false
//
// Permissions Required: Write
//
// Related: [IssueSubscribe]
func (c *Client) IssueUnsubscribe(ctx context.Context, id string, userID *string) error {
	resp, err := c.gqlClient.IssueUnsubscribe(ctx, id, userID)
	if err != nil {
		return wrapGraphQLError("IssueUnsubscribe", err)
	}
	if !resp.IssueUnsubscribe.Success {
		return errMutationFailed("IssueUnsubscribe")
	}
	return nil
}
