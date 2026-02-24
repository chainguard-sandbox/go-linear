package linear

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// IssueUpdateNullable updates an issue with support for explicit null values.
// Use this to remove parent/cycle/project assignments by passing NewNull().
//
// Example removing parent:
//
//	input := linear.IssueUpdateNullableInput{
//	    ParentID: linear.NewNull[string](),
//	}
//	issue, err := client.IssueUpdateNullable(ctx, "issue-id", input)
func (c *Client) IssueUpdateNullable(ctx context.Context, id string, input IssueUpdateNullableInput) (*intgraphql.UpdateIssue_IssueUpdate_Issue, error) {
	// Build GraphQL mutation manually to support explicit null
	mutation := `mutation UpdateIssue($id: String!, $input: IssueUpdateInput!) {
		issueUpdate(id: $id, input: $input) {
			success
			issue {
				id
				identifier
				title
				description
				priority
				parent { id identifier }
				cycle { id name }
				project { id name }
				state { id name }
				updatedAt
			}
		}
	}`

	// Build variables with explicit null handling
	variables := map[string]any{
		"id":    id,
		"input": input.ToMap(),
	}

	// Make request
	reqBody := map[string]any{
		"query":         mutation,
		"operationName": "UpdateIssue",
		"variables":     variables,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.config.APIKey)

	// Use the client's HTTP client (has retry/circuit breaker logic)
	resp, err := c.config.HTTPClient.Do(req) // #nosec G704 - BaseURL from trusted config, not user input
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("nil response from HTTP client")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			IssueUpdate struct {
				Success bool
				Issue   intgraphql.UpdateIssue_IssueUpdate_Issue
			}
		}
		Errors []map[string]any
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %v", result.Errors[0])
	}

	if !result.Data.IssueUpdate.Success {
		return nil, errMutationFailed("IssueUpdateNullable")
	}

	return &result.Data.IssueUpdate.Issue, nil
}

// IssueUpdateNullableInput is the input for IssueUpdateNullable.
type IssueUpdateNullableInput struct {
	Title           *string
	Description     *string
	StateID         *string
	Priority        *int64
	AddedLabelIds   []string
	RemovedLabelIds []string

	// Nullable fields - support explicit null for removal
	AssigneeID         Nullable[string]
	CycleID            Nullable[string]
	ParentID           Nullable[string]
	ProjectID          Nullable[string]
	DueDate            Nullable[string]
	ProjectMilestoneID Nullable[string]
}

// ToMap converts to map for JSON encoding with explicit null support.
func (i *IssueUpdateNullableInput) ToMap() map[string]any {
	m := make(map[string]any)

	if i.Title != nil {
		m["title"] = *i.Title
	}
	if i.Description != nil {
		m["description"] = *i.Description
	}
	if i.AssigneeID.IsSet() {
		if val, ok := i.AssigneeID.Get(); ok {
			if val == nil {
				m["assigneeId"] = nil // Explicit null
			} else {
				m["assigneeId"] = *val
			}
		}
	}
	if i.StateID != nil {
		m["stateId"] = *i.StateID
	}
	if i.Priority != nil {
		m["priority"] = *i.Priority
	}
	if len(i.AddedLabelIds) > 0 {
		m["addedLabelIds"] = i.AddedLabelIds
	}
	if len(i.RemovedLabelIds) > 0 {
		m["removedLabelIds"] = i.RemovedLabelIds
	}

	// Handle nullable fields
	if i.CycleID.IsSet() {
		if val, ok := i.CycleID.Get(); ok {
			if val == nil {
				m["cycleId"] = nil // Explicit null
			} else {
				m["cycleId"] = *val
			}
		}
	}

	if i.ParentID.IsSet() {
		if val, ok := i.ParentID.Get(); ok {
			if val == nil {
				m["parentId"] = nil // Explicit null
			} else {
				m["parentId"] = *val
			}
		}
	}

	if i.ProjectID.IsSet() {
		if val, ok := i.ProjectID.Get(); ok {
			if val == nil {
				m["projectId"] = nil // Explicit null
			} else {
				m["projectId"] = *val
			}
		}
	}

	if i.DueDate.IsSet() {
		if val, ok := i.DueDate.Get(); ok {
			if val == nil {
				m["dueDate"] = nil // Explicit null
			} else {
				m["dueDate"] = *val
			}
		}
	}

	if i.ProjectMilestoneID.IsSet() {
		if val, ok := i.ProjectMilestoneID.Get(); ok {
			if val == nil {
				m["projectMilestoneId"] = nil // Explicit null
			} else {
				m["projectMilestoneId"] = *val
			}
		}
	}

	return m
}
