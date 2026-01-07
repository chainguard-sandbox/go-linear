// Package state provides filtering for workflow states.
//
// Supported filters (7):
//   - created-at: Filter by creation date (--created-after, --created-before)
//   - updated-at: Filter by update date (--updated-after, --updated-before)
//   - id: Filter by workflow state UUID (--id)
//   - name: Filter by state name contains (--name)
//   - description: Filter by description contains (--description)
//   - type: Filter by state type (--type): triage, backlog, unstarted, started, completed, canceled
//   - team: Filter by team name/key (--team)
package state
