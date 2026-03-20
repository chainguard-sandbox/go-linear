package linear_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// ExampleClient_IssueCreate demonstrates creating an issue with common fields.
func ExampleClient_IssueCreate() {
	client, _ := linear.NewClient("lin_api_xxx")
	ctx := context.Background()

	// Get team ID
	teams, _ := client.Teams(ctx, nil, nil)
	teamID := teams.Nodes[0].ID

	// Create issue
	title := "Fix login bug"
	desc := "Users can't log in on Safari"
	priority := int64(linear.PriorityUrgent)

	issue, err := client.IssueCreate(ctx, linear.IssueCreateInput{
		TeamID:      teamID,
		Title:       &title,
		Description: &desc,
		Priority:    &priority,
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created issue: %s\n", issue.ID)
}

// ExampleClient_IssueUpdate_content demonstrates updating title and description.
func ExampleClient_IssueUpdate_content() {
	client, _ := linear.NewClient("lin_api_xxx")

	updatedTitle := "Updated: Fix critical login bug"
	updatedDesc := "Added reproduction steps"

	_, err := client.IssueUpdate(context.Background(), "issue-uuid", linear.IssueUpdateInput{
		Title:       &updatedTitle,
		Description: &updatedDesc,
	})

	if err != nil {
		log.Fatal(err)
	}
}

// ExampleClient_IssueUpdate_assignToCycle demonstrates sprint planning.
func ExampleClient_IssueUpdate_assignToCycle() {
	client, _ := linear.NewClient("lin_api_xxx")
	ctx := context.Background()

	// Get current cycle
	cycles, _ := client.Cycles(ctx, nil, nil)
	cycleID := cycles.Nodes[0].ID

	// Assign issue to current sprint
	_, err := client.IssueUpdate(ctx, "issue-uuid", linear.IssueUpdateInput{
		CycleID: &cycleID,
	})

	if err != nil {
		log.Fatal(err)
	}
}

// ExampleClient_IssueUpdate_priority demonstrates changing issue urgency.
func ExampleClient_IssueUpdate_priority() {
	client, _ := linear.NewClient("lin_api_xxx")

	urgent := int64(1)
	_, err := client.IssueUpdate(context.Background(), "issue-uuid", linear.IssueUpdateInput{
		Priority: &urgent,
	})

	if err != nil {
		log.Fatal(err)
	}
}

// ExampleClient_IssueUpdate_assignee demonstrates task assignment and unassignment.
func ExampleClient_IssueUpdate_assignee() {
	client, _ := linear.NewClient("lin_api_xxx")
	ctx := context.Background()

	// Assign to yourself
	viewer, _ := client.Viewer(ctx)
	_, err := client.IssueUpdate(ctx, "issue-uuid", linear.IssueUpdateInput{
		AssigneeID: &viewer.ID,
	})

	if err != nil {
		log.Fatal(err)
	}

	// Unassign (empty string)
	emptyAssignee := ""
	_, err = client.IssueUpdate(ctx, "issue-uuid", linear.IssueUpdateInput{
		AssigneeID: &emptyAssignee,
	})

	if err != nil {
		log.Fatal(err)
	}
}

// ExampleClient_IssueUpdate_state demonstrates workflow state transitions.
func ExampleClient_IssueUpdate_state() {
	client, _ := linear.NewClient("lin_api_xxx")
	ctx := context.Background()

	// Find "In Progress" state
	states, _ := client.WorkflowStates(ctx, nil, nil)
	var stateID string
	for _, state := range states.Nodes {
		if state.Type == "started" {
			stateID = state.ID
			break
		}
	}

	// Move issue to In Progress
	_, err := client.IssueUpdate(ctx, "issue-uuid", linear.IssueUpdateInput{
		StateID: &stateID,
	})

	if err != nil {
		log.Fatal(err)
	}
}

// ExampleClient_IssueUpdate_labels demonstrates adding and removing labels.
func ExampleClient_IssueUpdate_labels() {
	client, _ := linear.NewClient("lin_api_xxx")
	ctx := context.Background()

	// Find bug label
	labels, _ := client.IssueLabels(ctx, nil, nil)
	var bugLabelID string
	for _, label := range labels.Nodes {
		if label.Name == "bug" {
			bugLabelID = label.ID
			break
		}
	}

	// Add label
	labelIDsToAdd := []string{bugLabelID}
	_, err := client.IssueUpdate(ctx, "issue-uuid", linear.IssueUpdateInput{
		AddedLabelIds: labelIDsToAdd,
	})

	if err != nil {
		log.Fatal(err)
	}
}

// ExampleClient_CommentCreate demonstrates adding comments to issues.
func ExampleClient_CommentCreate() {
	client, _ := linear.NewClient("lin_api_xxx")

	issueID := "issue-uuid"
	body := "This also affects mobile users"

	_, err := client.CommentCreate(context.Background(), linear.CommentCreateInput{
		IssueID: &issueID,
		Body:    &body,
	})

	if err != nil {
		log.Fatal(err)
	}
}

// ExampleClient_IssueCreate_fullContext demonstrates gathering all IDs for complete issue creation.
func ExampleClient_IssueCreate_fullContext() {
	client, _ := linear.NewClient("lin_api_xxx")
	ctx := context.Background()

	// Gather IDs from various sources
	teams, _ := client.Teams(ctx, nil, nil)
	teamID := teams.Nodes[0].ID

	viewer, _ := client.Viewer(ctx)
	assigneeID := viewer.ID

	cycles, _ := client.Cycles(ctx, nil, nil)
	cycleID := cycles.Nodes[0].ID

	// Create with all relationships
	title := "Comprehensive issue"
	priority := int64(linear.PriorityUrgent)

	issue, err := client.IssueCreate(ctx, linear.IssueCreateInput{
		TeamID:     teamID,
		Title:      &title,
		Priority:   &priority,
		AssigneeID: &assigneeID,
		CycleID:    &cycleID,
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created: %s\n", issue.ID)
}

// ExampleNewIssueIterator_conditionalUpdate demonstrates bulk conditional updates.
func ExampleNewIssueIterator_conditionalUpdate() {
	client, _ := linear.NewClient("lin_api_xxx")
	ctx := context.Background()

	viewer, _ := client.Viewer(ctx)
	myID := viewer.ID

	// Bulk operation: Assign all unassigned in-progress issues to me
	iter := linear.NewIssueIterator(client, 50)
	for {
		issue, err := iter.Next(ctx)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// Condition: in-progress and unassigned
		if issue.State.Name == "Todo" {
			_, err := client.IssueUpdate(ctx, issue.ID, linear.IssueUpdateInput{
				AssigneeID: &myID,
			})
			if err != nil {
				log.Printf("Update failed: %v", err)
			}
		}
	}
}

// ExampleClient_IssueCreate_subIssue demonstrates creating sub-issues with safety checks.
func ExampleClient_IssueCreate_subIssue() {
	client, _ := linear.NewClient("lin_api_xxx")
	ctx := context.Background()

	// Get parent issue
	parent, _ := client.Issue(ctx, "parent-uuid")

	// Create sub-issue
	subTitle := "Subtask of parent"
	_, err := client.IssueCreate(ctx, linear.IssueCreateInput{
		TeamID:   parent.Team.ID,
		ParentID: &parent.ID,
		Title:    &subTitle,
	})

	if err != nil {
		log.Fatal(err)
	}
}
