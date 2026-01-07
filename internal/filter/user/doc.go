// Package user provides filtering for users.
//
// Supported filters (10):
//   - created-at: Filter by creation date (--created-after, --created-before)
//   - updated-at: Filter by update date (--updated-after, --updated-before)
//   - id: Filter by user UUID (--id)
//   - name: Filter by name contains (--name)
//   - email: Filter by email contains (--email)
//   - display-name: Filter by display name contains (--display-name)
//   - active: Filter by active status (--active)
//   - admin: Filter by admin status (--admin)
//   - is-me: Filter for current user (--is-me)
package user
