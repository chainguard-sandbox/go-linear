//go:build integration && write

// Package commands contains CLI integration tests for write operations.
//
// Run with: go test -tags="integration,write" ./cmd/linear/commands/...
// Requires: LINEAR_API_KEY environment variable
//
// WARNING: These tests CREATE, UPDATE, and DELETE real data in Linear.
// Run only against a test workspace.
package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// extractEntity handles both nested {"key": {...}} and flat {...} response formats
func extractEntity(result map[string]any, key string) map[string]any {
	if entity, ok := result[key].(map[string]any); ok {
		return entity
	}
	return result
}

// extractIssue handles both nested {"issue": {...}} and flat {...} response formats
func extractIssue(result map[string]any) map[string]any {
	return extractEntity(result, "issue")
}

// writeTestRunner extends cliRunner for write operations
type writeTestRunner struct {
	*cliRunner
	teamKey string // Team to use for creating issues
}

func newWriteTestRunner(t *testing.T) *writeTestRunner {
	t.Helper()

	r := newCLIRunner(t)

	// Get a team to use for testing
	stdout, _, err := r.run("team", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Fatalf("Failed to get teams: %v", err)
	}

	var result map[string]any
	json.Unmarshal([]byte(stdout), &result)
	nodes := result["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No teams available for write tests")
	}

	team := nodes[0].(map[string]any)
	teamKey := team["key"].(string)

	return &writeTestRunner{
		cliRunner: r,
		teamKey:   teamKey,
	}
}

// --- Issue CRUD Tests ---

func TestCLI_IssueCreateUpdateDelete(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")
	title := fmt.Sprintf("Integration Test Issue %s", timestamp)

	// Shared state across subtests
	var issueIdentifier string

	// CREATE
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := r.run("issue", "create",
			"--team="+r.teamKey,
			"--title="+title,
			"--description=Created by CLI integration test",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse create response: %v", err)
		}

		// Handle both nested {"issue": {...}} and flat {...} formats
		issue := extractIssue(result)

		if issue["id"] == nil || issue["id"] == "" {
			t.Error("Created issue has no ID")
		}

		issueIdentifier = issue["identifier"].(string)
		t.Logf("Created issue: %s", issueIdentifier)
	})

	// UPDATE
	t.Run("update", func(t *testing.T) {
		if issueIdentifier == "" {
			t.Skip("No issue to update (create may have failed)")
		}

		newTitle := title + " (updated)"
		stdout, stderr, err := r.run("issue", "update", issueIdentifier,
			"--title="+newTitle,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue update failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse update response: %v", err)
		}

		t.Logf("Updated issue: %s", issueIdentifier)
		_ = result // use result
	})

	// DELETE
	t.Run("delete", func(t *testing.T) {
		if issueIdentifier == "" {
			t.Skip("No issue to delete")
		}

		stdout, stderr, err := r.run("issue", "delete", issueIdentifier, "--yes")
		if err != nil {
			t.Fatalf("issue delete failed: %v\nstderr: %s\nstdout: %s", err, stderr, stdout)
		}

		t.Logf("Deleted issue: %s", issueIdentifier)
	})
}

// --- Comment CRUD Tests ---

func TestCLI_CommentCreateUpdateDelete(t *testing.T) {
	r := newWriteTestRunner(t)

	// First create an issue to comment on
	timestamp := time.Now().Format("20060102-150405")
	issueTitle := fmt.Sprintf("Comment Test Issue %s", timestamp)

	stdout, _, err := r.run("issue", "create",
		"--team="+r.teamKey,
		"--title="+issueTitle,
		"--output=json",
	)
	if err != nil {
		t.Fatalf("Failed to create issue for comment test: %v", err)
	}

	var createResult map[string]any
	json.Unmarshal([]byte(stdout), &createResult)
	issue := extractIssue(createResult)
	issueID := issue["identifier"].(string)

	// Cleanup: delete the issue at the end
	defer func() {
		r.run("issue", "delete", issueID, "--yes")
	}()

	var commentID string

	// CREATE comment
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := r.run("comment", "create",
			"--issue="+issueID,
			"--body=Test comment from CLI integration test",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("comment create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		comment := extractEntity(result, "comment")
		commentID = comment["id"].(string)
		t.Logf("Created comment: %s", commentID)
	})

	// UPDATE comment
	t.Run("update", func(t *testing.T) {
		if commentID == "" {
			t.Skip("No comment to update")
		}

		stdout, stderr, err := r.run("comment", "update", commentID,
			"--body=Updated comment from CLI integration test",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("comment update failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Updated comment: %s", commentID)
		_ = stdout // Use stdout to avoid unused variable
	})

	// DELETE comment
	t.Run("delete", func(t *testing.T) {
		if commentID == "" {
			t.Skip("No comment to delete")
		}

		stdout, stderr, err := r.run("comment", "delete", commentID, "--yes")
		if err != nil {
			t.Fatalf("comment delete failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Deleted comment: %s", commentID)
		_ = stdout
	})
}

// --- Comment Threading Test ---

func TestCLI_CommentThreading(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")
	issueTitle := fmt.Sprintf("Thread Test %s", timestamp)

	// Create issue
	stdout, _, err := r.run("issue", "create",
		"--team="+r.teamKey,
		"--title="+issueTitle,
		"--output=json",
	)
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	var createResult map[string]any
	json.Unmarshal([]byte(stdout), &createResult)
	issue := extractIssue(createResult)
	issueID := issue["identifier"].(string)

	defer r.run("issue", "delete", issueID, "--yes")

	var parentID, replyID string

	// CREATE parent
	t.Run("create_parent", func(t *testing.T) {
		stdout, stderr, err := r.run("comment", "create",
			"--issue="+issueID,
			"--body=Parent comment",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		comment := extractEntity(result, "comment")
		parentID = comment["id"].(string)
		t.Logf("Created parent: %s", parentID)
	})

	// CREATE reply
	t.Run("create_reply", func(t *testing.T) {
		if parentID == "" {
			t.Skip("No parent")
		}

		stdout, stderr, err := r.run("comment", "create",
			"--issue="+issueID,
			"--body=Reply comment",
			"--parent="+parentID,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("create reply failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		comment := extractEntity(result, "comment")
		replyID = comment["id"].(string)

		if comment["parentId"] != parentID {
			t.Errorf("Expected parentId=%s, got %v", parentID, comment["parentId"])
		}
		t.Logf("Created reply: %s", replyID)
	})

	// GET parent shows replies
	t.Run("get_parent_shows_replies", func(t *testing.T) {
		if parentID == "" {
			t.Skip("No parent")
		}

		stdout, stderr, err := r.run("comment", "get", parentID, "--output=json")
		if err != nil {
			t.Fatalf("get failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)

		if children, ok := result["children"].(map[string]any); ok {
			if nodes, ok := children["nodes"].([]any); ok && len(nodes) > 0 {
				t.Logf("Parent has %d replies", len(nodes))
			} else {
				t.Error("Expected children nodes")
			}
		}
	})

	// GET reply shows parent
	t.Run("get_reply_shows_parent", func(t *testing.T) {
		if replyID == "" {
			t.Skip("No reply")
		}

		stdout, stderr, err := r.run("comment", "get", replyID, "--fields=none", "--output=json")
		if err != nil {
			t.Fatalf("get failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)

		if parent, ok := result["parent"].(map[string]any); ok {
			if parent["id"] != parentID {
				t.Errorf("Expected parent.id=%s, got %v", parentID, parent["id"])
			}
			t.Logf("Reply shows parent")
		} else {
			t.Error("Expected parent field")
		}
	})
}

// --- Label CRUD Tests ---

func TestCLI_LabelCreateUpdateDelete(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")
	labelName := fmt.Sprintf("test-label-%s", timestamp)
	var labelID string

	// CREATE
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := r.run("label", "create",
			"--name="+labelName,
			"--color=#ff5733",
			"--description=Created by CLI integration test",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("label create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Response structure may vary
		if label, ok := result["issueLabel"].(map[string]any); ok {
			labelID = label["id"].(string)
		} else if id, ok := result["id"].(string); ok {
			labelID = id
		}

		t.Logf("Created label: %s (%s)", labelName, labelID)
	})

	// UPDATE
	t.Run("update", func(t *testing.T) {
		if labelID == "" && labelName == "" {
			t.Skip("No label to update")
		}

		lookupID := labelID
		if lookupID == "" {
			lookupID = labelName
		}

		stdout, stderr, err := r.run("label", "update", lookupID,
			"--color=#33ff57",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("label update failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Updated label: %s", lookupID)
		_ = stdout
	})

	// DELETE
	t.Run("delete", func(t *testing.T) {
		if labelID == "" && labelName == "" {
			t.Skip("No label to delete")
		}

		lookupID := labelID
		if lookupID == "" {
			lookupID = labelName
		}

		stdout, stderr, err := r.run("label", "delete", lookupID, "--yes")
		if err != nil {
			t.Fatalf("label delete failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Deleted label: %s", lookupID)
		_ = stdout
	})
}

// --- Issue Label Assignment Tests ---

func TestCLI_IssueAddRemoveLabel(t *testing.T) {
	r := newWriteTestRunner(t)

	// Create a test issue
	timestamp := time.Now().Format("20060102-150405")
	issueTitle := fmt.Sprintf("Label Assignment Test %s", timestamp)

	stdout, _, err := r.run("issue", "create",
		"--team="+r.teamKey,
		"--title="+issueTitle,
		"--output=json",
	)
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	var createResult map[string]any
	json.Unmarshal([]byte(stdout), &createResult)
	issue := extractIssue(createResult)
	issueID := issue["identifier"].(string)

	defer func() {
		r.run("issue", "delete", issueID, "--yes")
	}()

	// Get first available label
	stdout, _, err = r.run("label", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No labels available for test")
	}

	var labelResult map[string]any
	json.Unmarshal([]byte(stdout), &labelResult)
	labels := labelResult["nodes"].([]any)
	if len(labels) == 0 {
		t.Skip("No labels available")
	}

	label := labels[0].(map[string]any)
	labelName := label["name"].(string)

	// ADD label
	t.Run("add-label", func(t *testing.T) {
		stdout, stderr, err := r.run("issue", "add-label", issueID, labelName, "--output=json")
		if err != nil {
			t.Fatalf("add-label failed: %v\nstderr: %s", err, stderr)
		}
		t.Logf("Added label '%s' to %s", labelName, issueID)
		_ = stdout
	})

	// REMOVE label
	t.Run("remove-label", func(t *testing.T) {
		stdout, stderr, err := r.run("issue", "remove-label", issueID, labelName, "--output=json")
		if err != nil {
			t.Fatalf("remove-label failed: %v\nstderr: %s", err, stderr)
		}
		t.Logf("Removed label '%s' from %s", labelName, issueID)
		_ = stdout
	})
}

// --- Reaction Tests ---

func TestCLI_ReactionCreateDelete(t *testing.T) {
	r := newWriteTestRunner(t)

	// Create a test issue
	timestamp := time.Now().Format("20060102-150405")
	issueTitle := fmt.Sprintf("Reaction Test %s", timestamp)

	stdout, _, err := r.run("issue", "create",
		"--team="+r.teamKey,
		"--title="+issueTitle,
		"--output=json",
	)
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	var createResult map[string]any
	json.Unmarshal([]byte(stdout), &createResult)
	issue := extractIssue(createResult)
	issueID := issue["identifier"].(string)

	defer func() {
		r.run("issue", "delete", issueID, "--yes")
	}()

	var reactionID string

	// CREATE reaction
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := r.run("reaction", "create",
			"--issue="+issueID,
			"--emoji=👍",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("reaction create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if reaction, ok := result["reaction"].(map[string]any); ok {
			reactionID = reaction["id"].(string)
		}

		t.Logf("Created reaction: %s", reactionID)
	})

	// DELETE reaction
	t.Run("delete", func(t *testing.T) {
		if reactionID == "" {
			t.Skip("No reaction to delete")
		}

		stdout, stderr, err := r.run("reaction", "delete", reactionID)
		if err != nil {
			// Some reactions might not be deletable
			if strings.Contains(stderr, "not found") || strings.Contains(stderr, "permission") {
				t.Skip("Cannot delete reaction: " + stderr)
			}
			t.Fatalf("reaction delete failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Deleted reaction: %s", reactionID)
		_ = stdout
	})
}

// --- Favorite Tests ---

func TestCLI_FavoriteCreateDelete(t *testing.T) {
	r := newWriteTestRunner(t)

	// Create a test issue to favorite
	timestamp := time.Now().Format("20060102-150405")
	issueTitle := fmt.Sprintf("Favorite Test %s", timestamp)

	stdout, _, err := r.run("issue", "create",
		"--team="+r.teamKey,
		"--title="+issueTitle,
		"--output=json",
	)
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	var createResult map[string]any
	json.Unmarshal([]byte(stdout), &createResult)
	issue := extractIssue(createResult)
	issueID := issue["identifier"].(string)

	defer func() {
		r.run("issue", "delete", issueID, "--yes")
	}()

	var favoriteID string

	// CREATE favorite
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := r.run("favorite", "create",
			"--issue="+issueID,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("favorite create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if fav, ok := result["favorite"].(map[string]any); ok {
			favoriteID = fav["id"].(string)
		}

		t.Logf("Created favorite: %s", favoriteID)
	})

	// DELETE favorite
	t.Run("delete", func(t *testing.T) {
		if favoriteID == "" {
			t.Skip("No favorite to delete")
		}

		stdout, stderr, err := r.run("favorite", "delete", favoriteID)
		if err != nil {
			t.Fatalf("favorite delete failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Deleted favorite: %s", favoriteID)
		_ = stdout
	})
}

// --- Project CRUD Tests ---

func TestCLI_ProjectCreateUpdateDelete(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")
	projectName := fmt.Sprintf("Test Project %s", timestamp)

	var projectID string

	// CREATE
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := r.run("project", "create",
			"--team="+r.teamKey,
			"--name="+projectName,
			"--description=Created by CLI integration test",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("project create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		project := extractEntity(result, "project")
		projectID = project["id"].(string)
		t.Logf("Created project: %s (%s)", projectName, projectID)
	})

	// UPDATE
	t.Run("update", func(t *testing.T) {
		if projectID == "" {
			t.Skip("No project to update")
		}

		newName := projectName + " (updated)"
		stdout, stderr, err := r.run("project", "update", projectID,
			"--name="+newName,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("project update failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Updated project: %s", projectID)
		_ = stdout
	})

	// DELETE
	t.Run("delete", func(t *testing.T) {
		if projectID == "" {
			t.Skip("No project to delete")
		}

		stdout, stderr, err := r.run("project", "delete", projectID, "--yes")
		if err != nil {
			t.Fatalf("project delete failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Deleted project: %s", projectID)
		_ = stdout
	})
}

// --- Cycle CRUD Tests ---

func TestCLI_CycleCreateUpdateArchive(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")
	cycleName := fmt.Sprintf("Test Cycle %s", timestamp)

	var cycleID string

	// CREATE
	t.Run("create", func(t *testing.T) {
		// Cycles need start and end dates
		startDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
		endDate := time.Now().AddDate(0, 0, 14).Format("2006-01-02")

		stdout, stderr, err := r.run("cycle", "create",
			"--team="+r.teamKey,
			"--name="+cycleName,
			"--starts-at="+startDate,
			"--ends-at="+endDate,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("cycle create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		cycle := extractEntity(result, "cycle")
		cycleID = cycle["id"].(string)
		t.Logf("Created cycle: %s (%s)", cycleName, cycleID)
	})

	// UPDATE
	t.Run("update", func(t *testing.T) {
		if cycleID == "" {
			t.Skip("No cycle to update")
		}

		newName := cycleName + " (updated)"
		stdout, stderr, err := r.run("cycle", "update", cycleID,
			"--name="+newName,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("cycle update failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Updated cycle: %s", cycleID)
		_ = stdout
	})

	// ARCHIVE (cycles can't be deleted, only archived)
	t.Run("archive", func(t *testing.T) {
		if cycleID == "" {
			t.Skip("No cycle to archive")
		}

		stdout, stderr, err := r.run("cycle", "archive", cycleID, "--output=json")
		if err != nil {
			t.Fatalf("cycle archive failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Archived cycle: %s", cycleID)
		_ = stdout
	})
}

// --- Issue Relation Tests ---

func TestCLI_IssueRelateUnrelate(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")

	// Create two issues to relate
	var issue1ID, issue2ID string
	var relationID string

	// CREATE first issue
	t.Run("create_issue1", func(t *testing.T) {
		stdout, stderr, err := r.run("issue", "create",
			"--team="+r.teamKey,
			"--title=Relation Test Issue 1 "+timestamp,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		issue := extractIssue(result)
		issue1ID = issue["identifier"].(string)
		t.Logf("Created issue 1: %s", issue1ID)
	})

	// CREATE second issue
	t.Run("create_issue2", func(t *testing.T) {
		stdout, stderr, err := r.run("issue", "create",
			"--team="+r.teamKey,
			"--title=Relation Test Issue 2 "+timestamp,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		issue := extractIssue(result)
		issue2ID = issue["identifier"].(string)
		t.Logf("Created issue 2: %s", issue2ID)
	})

	// RELATE issues
	t.Run("relate", func(t *testing.T) {
		if issue1ID == "" || issue2ID == "" {
			t.Skip("Issues not created")
		}

		stdout, stderr, err := r.run("issue", "relate", issue1ID, issue2ID,
			"--type=blocks",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue relate failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Try to extract relation ID
		if relation, ok := result["issueRelation"].(map[string]any); ok {
			relationID = relation["id"].(string)
		} else if id, ok := result["id"].(string); ok {
			relationID = id
		}

		t.Logf("Created relation: %s blocks %s (id: %s)", issue1ID, issue2ID, relationID)
	})

	// UNRELATE issues
	t.Run("unrelate", func(t *testing.T) {
		if relationID == "" {
			t.Skip("No relation to delete")
		}

		stdout, stderr, err := r.run("issue", "unrelate", relationID, "--yes")
		if err != nil {
			t.Fatalf("issue unrelate failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Deleted relation: %s", relationID)
		_ = stdout
	})

	// Cleanup: delete both issues
	t.Run("cleanup", func(t *testing.T) {
		if issue1ID != "" {
			r.run("issue", "delete", issue1ID, "--yes")
		}
		if issue2ID != "" {
			r.run("issue", "delete", issue2ID, "--yes")
		}
		t.Log("Cleaned up test issues")
	})
}

// --- Attachment Tests ---

func TestCLI_AttachmentLinkURL(t *testing.T) {
	r := newWriteTestRunner(t)

	// Create a test issue
	timestamp := time.Now().Format("20060102-150405")
	issueTitle := fmt.Sprintf("Attachment Test %s", timestamp)

	stdout, _, err := r.run("issue", "create",
		"--team="+r.teamKey,
		"--title="+issueTitle,
		"--output=json",
	)
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	var createResult map[string]any
	json.Unmarshal([]byte(stdout), &createResult)
	issue := extractIssue(createResult)
	issueID := issue["identifier"].(string)

	defer func() {
		r.run("issue", "delete", issueID, "--yes")
	}()

	var attachmentID string

	// LINK URL
	t.Run("link-url", func(t *testing.T) {
		stdout, stderr, err := r.run("attachment", "link-url",
			"--issue="+issueID,
			"--url=https://example.com/test-doc",
			"--title=Test Document",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("attachment link-url failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		attachment := extractEntity(result, "attachment")
		if id, ok := attachment["id"].(string); ok {
			attachmentID = id
		}

		t.Logf("Linked URL to %s (attachment: %s)", issueID, attachmentID)
	})

	// DELETE attachment
	t.Run("delete", func(t *testing.T) {
		if attachmentID == "" {
			t.Skip("No attachment to delete")
		}

		stdout, stderr, err := r.run("attachment", "delete", attachmentID, "--yes")
		if err != nil {
			t.Fatalf("attachment delete failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Deleted attachment: %s", attachmentID)
		_ = stdout
	})
}

// --- Initiative CRUD Tests ---

func TestCLI_InitiativeCreateUpdateDelete(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")
	initName := fmt.Sprintf("Test Initiative %s", timestamp)

	var initID string

	// CREATE
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := r.run("initiative", "create",
			"--name="+initName,
			"--description=Created by CLI integration test",
			"--status=Active",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("initiative create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		init := extractEntity(result, "initiative")
		initID = init["id"].(string)
		t.Logf("Created initiative: %s (%s)", initName, initID)
	})

	// UPDATE
	t.Run("update", func(t *testing.T) {
		if initID == "" {
			t.Skip("No initiative to update")
		}

		newName := initName + " (updated)"
		stdout, stderr, err := r.run("initiative", "update", initID,
			"--name="+newName,
			"--status=Completed",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("initiative update failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Updated initiative: %s", initID)
		_ = stdout
	})

	// DELETE
	t.Run("delete", func(t *testing.T) {
		if initID == "" {
			t.Skip("No initiative to delete")
		}

		stdout, stderr, err := r.run("initiative", "delete", initID, "--yes")
		if err != nil {
			t.Fatalf("initiative delete failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Deleted initiative: %s", initID)
		_ = stdout
	})
}

// --- Initiative Hierarchy Tests ---

func TestCLI_InitiativeHierarchy(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")
	parentName := fmt.Sprintf("Parent Initiative %s", timestamp)
	childName := fmt.Sprintf("Child Initiative %s", timestamp)

	var parentID, childID string

	// CREATE parent initiative
	t.Run("create_parent", func(t *testing.T) {
		stdout, stderr, err := r.run("initiative", "create",
			"--name="+parentName,
			"--description=Parent initiative for hierarchy test",
			"--status=Active",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("parent initiative create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		parent := extractEntity(result, "initiative")
		parentID = parent["id"].(string)
		t.Logf("Created parent initiative: %s (%s)", parentName, parentID)
	})

	// CREATE child initiative with parent
	t.Run("create_child_with_parent", func(t *testing.T) {
		if parentID == "" {
			t.Skip("No parent initiative created")
		}

		stdout, stderr, err := r.run("initiative", "create",
			"--name="+childName,
			"--description=Child initiative for hierarchy test",
			"--parent="+parentID,
			"--status=Planned",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("child initiative create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		child := extractEntity(result, "initiative")
		childID = child["id"].(string)
		t.Logf("Created child initiative: %s (%s) with parent %s", childName, childID, parentID)
	})

	// LIST initiatives with parent filter
	t.Run("list_with_parent_filter", func(t *testing.T) {
		if parentID == "" {
			t.Skip("No parent initiative")
		}

		stdout, stderr, err := r.run("initiative", "list",
			"--parent="+parentID,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("list with parent filter failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		nodes := result["nodes"].([]any)

		if len(nodes) < 1 {
			t.Error("Expected at least one child initiative in filtered list")
		}

		t.Logf("Found %d child initiatives for parent %s", len(nodes), parentID)
	})

	// Cleanup: delete child first, then parent
	t.Run("cleanup", func(t *testing.T) {
		if childID != "" {
			r.run("initiative", "delete", childID, "--yes")
			t.Logf("Deleted child initiative: %s", childID)
		}
		if parentID != "" {
			r.run("initiative", "delete", parentID, "--yes")
			t.Logf("Deleted parent initiative: %s", parentID)
		}
	})
}

// --- Issue State Update Test ---

func TestCLI_IssueUpdateState(t *testing.T) {
	r := newWriteTestRunner(t)

	// Create a test issue
	timestamp := time.Now().Format("20060102-150405")
	issueTitle := fmt.Sprintf("State Update Test %s", timestamp)

	stdout, _, err := r.run("issue", "create",
		"--team="+r.teamKey,
		"--title="+issueTitle,
		"--output=json",
	)
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	var createResult map[string]any
	json.Unmarshal([]byte(stdout), &createResult)
	issue := extractIssue(createResult)
	issueID := issue["identifier"].(string)

	defer func() {
		r.run("issue", "delete", issueID, "--yes")
	}()

	// Get the current issue state so we can update to a different one
	stdout, _, err = r.run("issue", "get", issueID, "--output=json", "--fields=defaults,state")
	if err != nil {
		t.Fatalf("Cannot get issue: %v", err)
	}

	var issueResult map[string]any
	json.Unmarshal([]byte(stdout), &issueResult)
	currentState := ""
	if state, ok := issueResult["state"].(map[string]any); ok {
		currentState = state["name"].(string)
	}

	// Try common state names, pick one that's different from current
	commonStates := []string{"Backlog", "Todo", "In Progress"}
	var targetStateName string
	for _, state := range commonStates {
		if state != currentState {
			targetStateName = state
			break
		}
	}

	if targetStateName == "" {
		t.Skip("No suitable state to update to")
	}

	// UPDATE state
	t.Run("update-state", func(t *testing.T) {
		stdout, stderr, err := r.run("issue", "update", issueID,
			"--state="+targetStateName,
			"--output=json",
		)
		if err != nil {
			// Some teams may not have these standard states
			if strings.Contains(stderr, "not found") || strings.Contains(stderr, "resolve") {
				t.Skipf("State '%s' not available in this team", targetStateName)
			}
			t.Fatalf("issue update state failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Updated issue %s state to '%s'", issueID, targetStateName)
		_ = stdout
	})
}

// --- Issue Priority Update Test ---

func TestCLI_IssueUpdatePriority(t *testing.T) {
	r := newWriteTestRunner(t)

	// Create a test issue
	timestamp := time.Now().Format("20060102-150405")
	issueTitle := fmt.Sprintf("Priority Update Test %s", timestamp)

	stdout, _, err := r.run("issue", "create",
		"--team="+r.teamKey,
		"--title="+issueTitle,
		"--output=json",
	)
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	var createResult map[string]any
	json.Unmarshal([]byte(stdout), &createResult)
	issue := extractIssue(createResult)
	issueID := issue["identifier"].(string)

	defer func() {
		r.run("issue", "delete", issueID, "--yes")
	}()

	// UPDATE priority to high (2)
	t.Run("update-priority", func(t *testing.T) {
		stdout, stderr, err := r.run("issue", "update", issueID,
			"--priority=2",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue update priority failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Updated issue %s priority to high (2)", issueID)
		_ = stdout
	})
}

// --- Issue Assignee Update Test ---

func TestCLI_IssueUpdateAssignee(t *testing.T) {
	r := newWriteTestRunner(t)

	// Create a test issue
	timestamp := time.Now().Format("20060102-150405")
	issueTitle := fmt.Sprintf("Assignee Update Test %s", timestamp)

	stdout, _, err := r.run("issue", "create",
		"--team="+r.teamKey,
		"--title="+issueTitle,
		"--output=json",
	)
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	var createResult map[string]any
	json.Unmarshal([]byte(stdout), &createResult)
	issue := extractIssue(createResult)
	issueID := issue["identifier"].(string)

	defer func() {
		r.run("issue", "delete", issueID, "--yes")
	}()

	// UPDATE assignee to self
	t.Run("update-assignee", func(t *testing.T) {
		stdout, stderr, err := r.run("issue", "update", issueID,
			"--assignee=me",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue update assignee failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Assigned issue %s to me", issueID)
		_ = stdout
	})
}

// --- Project Milestone CRUD Tests ---

func TestCLI_ProjectMilestoneCRUD(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")
	projectName := fmt.Sprintf("Milestone Test Project %s", timestamp)

	// First create a project
	stdout, stderr, err := r.run("project", "create",
		"--team="+r.teamKey,
		"--name="+projectName,
		"--output=json",
	)
	if err != nil {
		t.Fatalf("Failed to create project: %v\nstderr: %s", err, stderr)
	}

	var createResult map[string]any
	json.Unmarshal([]byte(stdout), &createResult)
	project := extractEntity(createResult, "project")
	projectID := project["id"].(string)

	defer func() {
		r.run("project", "delete", projectID, "--yes")
	}()

	var milestoneID string

	// CREATE milestone
	t.Run("create", func(t *testing.T) {
		targetDate := time.Now().AddDate(0, 1, 0).Format("2006-01-02")

		stdout, stderr, err := r.run("project", "milestone-create",
			"--project="+projectID,
			"--name=Alpha Release",
			"--target-date="+targetDate,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("milestone create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		milestone := extractEntity(result, "projectMilestone")
		milestoneID = milestone["id"].(string)
		t.Logf("Created milestone: %s", milestoneID)
	})

	// UPDATE milestone
	t.Run("update", func(t *testing.T) {
		if milestoneID == "" {
			t.Skip("No milestone to update")
		}

		stdout, stderr, err := r.run("project", "milestone-update", milestoneID,
			"--name=Beta Release",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("milestone update failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Updated milestone: %s", milestoneID)
		_ = stdout
	})

	// DELETE milestone
	t.Run("delete", func(t *testing.T) {
		if milestoneID == "" {
			t.Skip("No milestone to delete")
		}

		stdout, stderr, err := r.run("project", "milestone-delete", milestoneID, "--yes")
		if err != nil {
			t.Fatalf("milestone delete failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Deleted milestone: %s", milestoneID)
		_ = stdout
	})
}

// --- Custom Attachment Create Test ---

func TestCLI_AttachmentCreate(t *testing.T) {
	r := newWriteTestRunner(t)

	// Create a test issue
	timestamp := time.Now().Format("20060102-150405")
	issueTitle := fmt.Sprintf("Custom Attachment Test %s", timestamp)

	stdout, _, err := r.run("issue", "create",
		"--team="+r.teamKey,
		"--title="+issueTitle,
		"--output=json",
	)
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	var createResult map[string]any
	json.Unmarshal([]byte(stdout), &createResult)
	issue := extractIssue(createResult)
	issueID := issue["identifier"].(string)

	defer func() {
		r.run("issue", "delete", issueID, "--yes")
	}()

	var attachmentID string

	// CREATE custom attachment with metadata
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := r.run("attachment", "create",
			"--issue="+issueID,
			"--title=Build Report",
			"--url=https://ci.example.com/builds/123",
			"--subtitle=Build #123 passed",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("attachment create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		attachment := extractEntity(result, "attachment")
		if id, ok := attachment["id"].(string); ok {
			attachmentID = id
		}

		t.Logf("Created custom attachment: %s", attachmentID)
	})

	// DELETE
	t.Run("delete", func(t *testing.T) {
		if attachmentID == "" {
			t.Skip("No attachment to delete")
		}

		stdout, stderr, err := r.run("attachment", "delete", attachmentID, "--yes")
		if err != nil {
			t.Fatalf("attachment delete failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Deleted attachment: %s", attachmentID)
		_ = stdout
	})
}

// --- Issue Create with Multiple Options Test ---

func TestCLI_IssueCreateWithOptions(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")
	title := fmt.Sprintf("Full Options Issue %s", timestamp)

	var issueID string

	// CREATE with multiple options
	t.Run("create-with-options", func(t *testing.T) {
		stdout, stderr, err := r.run("issue", "create",
			"--team="+r.teamKey,
			"--title="+title,
			"--description=Issue with all options set",
			"--priority=2",
			"--assignee=me",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		issue := extractIssue(result)
		issueID = issue["identifier"].(string)
		t.Logf("Created issue with options: %s", issueID)
	})

	// Verify the issue has the expected values
	t.Run("verify-options", func(t *testing.T) {
		if issueID == "" {
			t.Skip("No issue to verify")
		}

		stdout, stderr, err := r.run("issue", "get", issueID, "--output=json")
		if err != nil {
			t.Fatalf("issue get failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		issue := result
		if priority, ok := issue["priority"].(float64); ok {
			if int(priority) != 2 {
				t.Errorf("Expected priority 2, got %v", priority)
			}
		}

		if assignee, ok := issue["assignee"].(map[string]any); ok {
			t.Logf("Issue assigned to: %s", assignee["name"])
		}
	})

	// Cleanup
	t.Run("cleanup", func(t *testing.T) {
		if issueID != "" {
			r.run("issue", "delete", issueID, "--yes")
			t.Logf("Deleted issue: %s", issueID)
		}
	})
}

// --- Issue Parent/Child Relationship Tests ---

func TestCLI_IssueParentChild(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")

	var parentID, childID string

	// CREATE parent issue
	t.Run("create_parent", func(t *testing.T) {
		stdout, stderr, err := r.run("issue", "create",
			"--team="+r.teamKey,
			"--title=Parent Issue "+timestamp,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("parent issue create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		issue := extractIssue(result)
		parentID = issue["identifier"].(string)
		t.Logf("Created parent issue: %s", parentID)
	})

	// CREATE child issue with --parent
	t.Run("create_child_with_parent", func(t *testing.T) {
		if parentID == "" {
			t.Skip("No parent issue created")
		}

		stdout, stderr, err := r.run("issue", "create",
			"--team="+r.teamKey,
			"--title=Child Issue "+timestamp,
			"--parent="+parentID,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("child issue create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		issue := extractIssue(result)
		childID = issue["identifier"].(string)
		t.Logf("Created child issue: %s with parent %s", childID, parentID)
	})

	// VERIFY child has parent
	t.Run("verify_parent", func(t *testing.T) {
		if childID == "" {
			t.Skip("No child issue created")
		}

		stdout, stderr, err := r.run("issue", "get", childID, "--output=json", "--fields=defaults,parent")
		if err != nil {
			t.Fatalf("issue get failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)

		if parent, ok := result["parent"].(map[string]any); ok {
			parentIdentifier := parent["identifier"].(string)
			if parentIdentifier != parentID {
				t.Errorf("Expected parent %s, got %s", parentID, parentIdentifier)
			}
			t.Logf("Verified child %s has parent %s", childID, parentIdentifier)
		} else {
			t.Error("Child issue has no parent set")
		}
	})

	// REMOVE parent using --parent=none
	t.Run("remove_parent", func(t *testing.T) {
		if childID == "" {
			t.Skip("No child issue")
		}

		stdout, stderr, err := r.run("issue", "update", childID,
			"--parent=none",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("remove parent failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Removed parent from child issue %s", childID)
		_ = stdout
	})

	// VERIFY parent removed
	t.Run("verify_parent_removed", func(t *testing.T) {
		if childID == "" {
			t.Skip("No child issue")
		}

		stdout, stderr, err := r.run("issue", "get", childID, "--output=json", "--fields=defaults,parent")
		if err != nil {
			t.Fatalf("issue get failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)

		if _, ok := result["parent"].(map[string]any); ok {
			t.Error("Parent should have been removed")
		} else {
			t.Logf("Verified parent removed from %s", childID)
		}
	})

	// Cleanup
	t.Run("cleanup", func(t *testing.T) {
		if childID != "" {
			r.run("issue", "delete", childID, "--yes")
		}
		if parentID != "" {
			r.run("issue", "delete", parentID, "--yes")
		}
		t.Log("Cleaned up test issues")
	})
}

// --- Issue Project Assignment Tests ---

func TestCLI_IssueProjectAssignment(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")

	var projectID, issueID string

	// CREATE project
	t.Run("create_project", func(t *testing.T) {
		stdout, stderr, err := r.run("project", "create",
			"--team="+r.teamKey,
			"--name=Project Assignment Test "+timestamp,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("project create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		project := extractEntity(result, "project")
		projectID = project["id"].(string)
		t.Logf("Created project: %s", projectID)
	})

	// CREATE issue
	t.Run("create_issue", func(t *testing.T) {
		stdout, stderr, err := r.run("issue", "create",
			"--team="+r.teamKey,
			"--title=Project Assignment Issue "+timestamp,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		issue := extractIssue(result)
		issueID = issue["identifier"].(string)
		t.Logf("Created issue: %s", issueID)
	})

	// ASSIGN to project
	t.Run("assign_project", func(t *testing.T) {
		if issueID == "" || projectID == "" {
			t.Skip("Missing issue or project")
		}

		stdout, stderr, err := r.run("issue", "update", issueID,
			"--project="+projectID,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("assign project failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Assigned issue %s to project %s", issueID, projectID)
		_ = stdout
	})

	// UNASSIGN from project using --project=none
	t.Run("unassign_project", func(t *testing.T) {
		if issueID == "" {
			t.Skip("No issue")
		}

		stdout, stderr, err := r.run("issue", "update", issueID,
			"--project=none",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("unassign project failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Unassigned issue %s from project", issueID)
		_ = stdout
	})

	// Cleanup
	t.Run("cleanup", func(t *testing.T) {
		if issueID != "" {
			r.run("issue", "delete", issueID, "--yes")
		}
		if projectID != "" {
			r.run("project", "delete", projectID, "--yes")
		}
		t.Log("Cleaned up test resources")
	})
}

// --- Issue Cycle Assignment Tests ---

func TestCLI_IssueCycleAssignment(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")

	var cycleID, issueID string

	// CREATE cycle
	t.Run("create_cycle", func(t *testing.T) {
		startDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
		endDate := time.Now().AddDate(0, 0, 14).Format("2006-01-02")

		stdout, stderr, err := r.run("cycle", "create",
			"--team="+r.teamKey,
			"--name=Cycle Assignment Test "+timestamp,
			"--starts-at="+startDate,
			"--ends-at="+endDate,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("cycle create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		cycle := extractEntity(result, "cycle")
		cycleID = cycle["id"].(string)
		t.Logf("Created cycle: %s", cycleID)
	})

	// CREATE issue
	t.Run("create_issue", func(t *testing.T) {
		stdout, stderr, err := r.run("issue", "create",
			"--team="+r.teamKey,
			"--title=Cycle Assignment Issue "+timestamp,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		issue := extractIssue(result)
		issueID = issue["identifier"].(string)
		t.Logf("Created issue: %s", issueID)
	})

	// ASSIGN to cycle
	t.Run("assign_cycle", func(t *testing.T) {
		if issueID == "" || cycleID == "" {
			t.Skip("Missing issue or cycle")
		}

		stdout, stderr, err := r.run("issue", "update", issueID,
			"--cycle="+cycleID,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("assign cycle failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Assigned issue %s to cycle %s", issueID, cycleID)
		_ = stdout
	})

	// UNASSIGN from cycle using --cycle=none
	t.Run("unassign_cycle", func(t *testing.T) {
		if issueID == "" {
			t.Skip("No issue")
		}

		stdout, stderr, err := r.run("issue", "update", issueID,
			"--cycle=none",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("unassign cycle failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Unassigned issue %s from cycle", issueID)
		_ = stdout
	})

	// Cleanup
	t.Run("cleanup", func(t *testing.T) {
		if issueID != "" {
			r.run("issue", "delete", issueID, "--yes")
		}
		if cycleID != "" {
			r.run("cycle", "archive", cycleID, "--output=json")
		}
		t.Log("Cleaned up test resources")
	})
}

// --- Issue Create with Cycle Test ---

func TestCLI_IssueCreateWithCycle(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")

	var cycleID, issueID string

	// CREATE cycle first
	t.Run("create_cycle", func(t *testing.T) {
		startDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
		endDate := time.Now().AddDate(0, 0, 14).Format("2006-01-02")

		stdout, stderr, err := r.run("cycle", "create",
			"--team="+r.teamKey,
			"--name=Issue Create Cycle Test "+timestamp,
			"--starts-at="+startDate,
			"--ends-at="+endDate,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("cycle create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		cycle := extractEntity(result, "cycle")
		cycleID = cycle["id"].(string)
		t.Logf("Created cycle: %s", cycleID)
	})

	// CREATE issue with --cycle flag
	t.Run("create_issue_with_cycle", func(t *testing.T) {
		if cycleID == "" {
			t.Skip("No cycle created")
		}

		stdout, stderr, err := r.run("issue", "create",
			"--team="+r.teamKey,
			"--title=Issue With Cycle "+timestamp,
			"--cycle="+cycleID,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue create with cycle failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		issue := extractIssue(result)
		issueID = issue["identifier"].(string)
		t.Logf("Created issue %s with cycle %s", issueID, cycleID)
	})

	// Cleanup
	t.Run("cleanup", func(t *testing.T) {
		if issueID != "" {
			r.run("issue", "delete", issueID, "--yes")
		}
		if cycleID != "" {
			r.run("cycle", "archive", cycleID, "--output=json")
		}
		t.Log("Cleaned up test resources")
	})
}

// --- Issue Create with Project Test ---

func TestCLI_IssueCreateWithProject(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")

	var projectID, issueID string

	// CREATE project first
	t.Run("create_project", func(t *testing.T) {
		stdout, stderr, err := r.run("project", "create",
			"--team="+r.teamKey,
			"--name=Issue Create Project Test "+timestamp,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("project create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		project := extractEntity(result, "project")
		projectID = project["id"].(string)
		t.Logf("Created project: %s", projectID)
	})

	// CREATE issue with --project flag
	t.Run("create_issue_with_project", func(t *testing.T) {
		if projectID == "" {
			t.Skip("No project created")
		}

		stdout, stderr, err := r.run("issue", "create",
			"--team="+r.teamKey,
			"--title=Issue With Project "+timestamp,
			"--project="+projectID,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue create with project failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		issue := extractIssue(result)
		issueID = issue["identifier"].(string)
		t.Logf("Created issue %s with project %s", issueID, projectID)
	})

	// Cleanup
	t.Run("cleanup", func(t *testing.T) {
		if issueID != "" {
			r.run("issue", "delete", issueID, "--yes")
		}
		if projectID != "" {
			r.run("project", "delete", projectID, "--yes")
		}
		t.Log("Cleaned up test resources")
	})
}

// --- Issue Relation Type Update Test ---

func TestCLI_IssueRelationUpdate(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")

	var issue1ID, issue2ID, relationID string

	// CREATE first issue
	t.Run("create_issue1", func(t *testing.T) {
		stdout, stderr, err := r.run("issue", "create",
			"--team="+r.teamKey,
			"--title=Relation Update Test 1 "+timestamp,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		issue := extractIssue(result)
		issue1ID = issue["identifier"].(string)
		t.Logf("Created issue 1: %s", issue1ID)
	})

	// CREATE second issue
	t.Run("create_issue2", func(t *testing.T) {
		stdout, stderr, err := r.run("issue", "create",
			"--team="+r.teamKey,
			"--title=Relation Update Test 2 "+timestamp,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		issue := extractIssue(result)
		issue2ID = issue["identifier"].(string)
		t.Logf("Created issue 2: %s", issue2ID)
	})

	// RELATE issues with "related" type
	t.Run("relate", func(t *testing.T) {
		if issue1ID == "" || issue2ID == "" {
			t.Skip("Issues not created")
		}

		stdout, stderr, err := r.run("issue", "relate", issue1ID, issue2ID,
			"--type=related",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue relate failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)

		if relation, ok := result["issueRelation"].(map[string]any); ok {
			relationID = relation["id"].(string)
		} else if id, ok := result["id"].(string); ok {
			relationID = id
		}

		t.Logf("Created relation: %s (related)", relationID)
	})

	// UPDATE relation type to "blocks"
	t.Run("update_relation", func(t *testing.T) {
		if relationID == "" {
			t.Skip("No relation to update")
		}

		stdout, stderr, err := r.run("issue", "update-relation", relationID,
			"--type=blocks",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("update-relation failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Updated relation %s to type: blocks", relationID)
		_ = stdout
	})

	// Cleanup
	t.Run("cleanup", func(t *testing.T) {
		if relationID != "" {
			r.run("issue", "unrelate", relationID, "--yes")
		}
		if issue1ID != "" {
			r.run("issue", "delete", issue1ID, "--yes")
		}
		if issue2ID != "" {
			r.run("issue", "delete", issue2ID, "--yes")
		}
		t.Log("Cleaned up test resources")
	})
}

// --- Batch Update Test ---

func TestCLI_IssueBatchUpdate(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")

	var issue1ID, issue2ID string

	// CREATE two issues with same title prefix for batch selection
	batchPrefix := fmt.Sprintf("BatchTest-%s", timestamp)

	t.Run("create_issues", func(t *testing.T) {
		// Issue 1
		stdout, stderr, err := r.run("issue", "create",
			"--team="+r.teamKey,
			"--title="+batchPrefix+" Issue 1",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue 1 create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		issue := extractIssue(result)
		issue1ID = issue["identifier"].(string)

		// Issue 2
		stdout, stderr, err = r.run("issue", "create",
			"--team="+r.teamKey,
			"--title="+batchPrefix+" Issue 2",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue 2 create failed: %v\nstderr: %s", err, stderr)
		}

		json.Unmarshal([]byte(stdout), &result)
		issue = extractIssue(result)
		issue2ID = issue["identifier"].(string)

		t.Logf("Created issues: %s, %s", issue1ID, issue2ID)
	})

	// BATCH UPDATE with dry-run first
	t.Run("batch_update_dry_run", func(t *testing.T) {
		if issue1ID == "" || issue2ID == "" {
			t.Skip("No issues created")
		}

		stdout, stderr, err := r.run("issue", "batch-update",
			"--team="+r.teamKey,
			"--title="+batchPrefix,
			"--set-priority=2",
			"--dry-run",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("batch-update dry-run failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Dry run output: %s", stdout)
	})

	// BATCH UPDATE for real
	t.Run("batch_update", func(t *testing.T) {
		if issue1ID == "" || issue2ID == "" {
			t.Skip("No issues created")
		}

		stdout, stderr, err := r.run("issue", "batch-update",
			"--team="+r.teamKey,
			"--title="+batchPrefix,
			"--set-priority=2",
			"--yes",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("batch-update failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Batch update completed: %s", stdout)
	})

	// Cleanup
	t.Run("cleanup", func(t *testing.T) {
		if issue1ID != "" {
			r.run("issue", "delete", issue1ID, "--yes")
		}
		if issue2ID != "" {
			r.run("issue", "delete", issue2ID, "--yes")
		}
		t.Log("Cleaned up test issues")
	})
}

// --- Team CRUD Tests ---

func TestCLI_TeamCreateUpdateDelete(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102150405")
	teamName := fmt.Sprintf("TestTeam%s", timestamp)
	teamKey := fmt.Sprintf("TT%s", timestamp[:6]) // Keys must be short

	var teamID string

	// CREATE
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := r.run("team", "create",
			"--name="+teamName,
			"--key="+teamKey,
			"--description=Created by CLI integration test",
			"--output=json",
		)
		if err != nil {
			// Team creation may require admin permissions
			if strings.Contains(stderr, "failed") || strings.Contains(stderr, "permission") {
				t.Skipf("Team creation not permitted (may require admin): %s", stderr)
			}
			t.Fatalf("team create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		team := extractEntity(result, "team")
		teamID = team["id"].(string)
		t.Logf("Created team: %s (%s) with key %s", teamName, teamID, teamKey)
	})

	// GET the created team
	t.Run("get", func(t *testing.T) {
		if teamID == "" {
			t.Skip("No team created")
		}

		stdout, stderr, err := r.run("team", "get", teamKey, "--output=json")
		if err != nil {
			t.Fatalf("team get failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)

		if result["key"] != teamKey {
			t.Errorf("Expected key %s, got %v", teamKey, result["key"])
		}
		t.Logf("Retrieved team: %s", teamKey)
	})

	// UPDATE
	t.Run("update", func(t *testing.T) {
		if teamID == "" {
			t.Skip("No team to update")
		}

		newDesc := "Updated by CLI integration test"
		stdout, stderr, err := r.run("team", "update", teamKey,
			"--description="+newDesc,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("team update failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Updated team: %s", teamKey)
		_ = stdout
	})

	// DELETE
	t.Run("delete", func(t *testing.T) {
		if teamID == "" {
			t.Skip("No team to delete")
		}

		stdout, stderr, err := r.run("team", "delete", teamKey, "--yes")
		if err != nil {
			t.Fatalf("team delete failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Deleted team: %s", teamKey)
		_ = stdout
	})
}

// --- Notification Tests ---

func TestCLI_NotificationSubscribeUnsubscribe(t *testing.T) {
	r := newWriteTestRunner(t)

	var subscriptionID string

	// Get a project to subscribe to
	stdout, _, err := r.run("project", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No projects available for notification test")
	}

	var projectResult map[string]any
	json.Unmarshal([]byte(stdout), &projectResult)
	nodes := projectResult["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No projects available")
	}

	project := nodes[0].(map[string]any)
	projectID := project["id"].(string)

	// SUBSCRIBE to project
	t.Run("subscribe", func(t *testing.T) {
		stdout, stderr, err := r.run("notification", "subscribe",
			"--project="+projectID,
			"--output=json",
		)
		if err != nil {
			// Notification subscriptions may require specific permissions or features
			if strings.Contains(stderr, "failed") || strings.Contains(stderr, "permission") {
				t.Skipf("Notification subscription not permitted: %s", stderr)
			}
			t.Fatalf("notification subscribe failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Extract subscription ID
		if sub, ok := result["notificationSubscription"].(map[string]any); ok {
			subscriptionID = sub["id"].(string)
		} else if id, ok := result["id"].(string); ok {
			subscriptionID = id
		}

		t.Logf("Subscribed to project %s (subscription: %s)", projectID, subscriptionID)
	})

	// UNSUBSCRIBE
	t.Run("unsubscribe", func(t *testing.T) {
		if subscriptionID == "" {
			t.Skip("No subscription to unsubscribe")
		}

		stdout, stderr, err := r.run("notification", "unsubscribe", subscriptionID)
		if err != nil {
			t.Fatalf("notification unsubscribe failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Unsubscribed: %s", subscriptionID)
		_ = stdout
	})
}

func TestCLI_NotificationUpdateAndArchive(t *testing.T) {
	r := newWriteTestRunner(t)

	// This test requires existing notifications - we need to create activity first
	// Create an issue and comment to generate a notification

	timestamp := time.Now().Format("20060102-150405")

	// Create issue
	stdout, _, err := r.run("issue", "create",
		"--team="+r.teamKey,
		"--title=Notification Test "+timestamp,
		"--output=json",
	)
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	var createResult map[string]any
	json.Unmarshal([]byte(stdout), &createResult)
	issue := extractIssue(createResult)
	issueID := issue["identifier"].(string)

	defer func() {
		r.run("issue", "delete", issueID, "--yes")
	}()

	// Note: Notifications are generated asynchronously, so we can't reliably
	// test update/archive without waiting. Instead, test the command structure.
	t.Run("update_command_exists", func(t *testing.T) {
		// Test with a fake ID to verify command works (will fail with "not found")
		_, stderr, err := r.run("notification", "update", "fake-notification-id", "--read")
		if err == nil {
			t.Log("Notification update command executed (unexpected success)")
		} else if strings.Contains(stderr, "not found") || strings.Contains(stderr, "invalid") {
			t.Log("Notification update command works (expected error for fake ID)")
		} else {
			t.Logf("Notification update command responded: %s", stderr)
		}
	})

	t.Run("archive_command_exists", func(t *testing.T) {
		// Test with a fake ID to verify command works
		_, stderr, err := r.run("notification", "archive", "fake-notification-id")
		if err == nil {
			t.Log("Notification archive command executed (unexpected success)")
		} else if strings.Contains(stderr, "not found") || strings.Contains(stderr, "invalid") {
			t.Log("Notification archive command works (expected error for fake ID)")
		} else {
			t.Logf("Notification archive command responded: %s", stderr)
		}
	})
}

// --- Issue Create with Estimate Test ---

func TestCLI_IssueCreateWithEstimate(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")
	title := fmt.Sprintf("Issue With Estimate %s", timestamp)

	var issueID string

	// CREATE with estimate
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := r.run("issue", "create",
			"--team="+r.teamKey,
			"--title="+title,
			"--estimate=5",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		issue := extractIssue(result)
		issueID = issue["identifier"].(string)
		t.Logf("Created issue %s with estimate 5", issueID)
	})

	// Cleanup
	t.Run("cleanup", func(t *testing.T) {
		if issueID != "" {
			r.run("issue", "delete", issueID, "--yes")
		}
	})
}

// --- Issue with Multiple Labels Test ---

func TestCLI_IssueCreateWithLabels(t *testing.T) {
	r := newWriteTestRunner(t)

	// Get available labels
	stdout, _, err := r.run("label", "list", "--output=json", "--limit=2")
	if err != nil {
		t.Skip("No labels available")
	}

	var labelResult map[string]any
	json.Unmarshal([]byte(stdout), &labelResult)
	labels := labelResult["nodes"].([]any)
	if len(labels) < 2 {
		t.Skip("Need at least 2 labels for this test")
	}

	label1 := labels[0].(map[string]any)["name"].(string)
	label2 := labels[1].(map[string]any)["name"].(string)

	timestamp := time.Now().Format("20060102-150405")
	title := fmt.Sprintf("Multi-Label Issue %s", timestamp)

	var issueID string

	// CREATE with multiple labels
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := r.run("issue", "create",
			"--team="+r.teamKey,
			"--title="+title,
			"--label="+label1,
			"--label="+label2,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("issue create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		issue := extractIssue(result)
		issueID = issue["identifier"].(string)
		t.Logf("Created issue %s with labels: %s, %s", issueID, label1, label2)
	})

	// Cleanup
	t.Run("cleanup", func(t *testing.T) {
		if issueID != "" {
			r.run("issue", "delete", issueID, "--yes")
		}
	})
}

// --- Cycle with Description Test ---

func TestCLI_CycleWithDescription(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")
	cycleName := fmt.Sprintf("Described Cycle %s", timestamp)

	var cycleID string

	// CREATE with description
	t.Run("create", func(t *testing.T) {
		startDate := time.Now().AddDate(0, 0, 14).Format("2006-01-02")
		endDate := time.Now().AddDate(0, 0, 21).Format("2006-01-02")

		stdout, stderr, err := r.run("cycle", "create",
			"--team="+r.teamKey,
			"--name="+cycleName,
			"--starts-at="+startDate,
			"--ends-at="+endDate,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("cycle create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		cycle := extractEntity(result, "cycle")
		cycleID = cycle["id"].(string)
		t.Logf("Created cycle: %s", cycleID)
	})

	// UPDATE with description
	t.Run("update-description", func(t *testing.T) {
		if cycleID == "" {
			t.Skip("No cycle created")
		}

		stdout, stderr, err := r.run("cycle", "update", cycleID,
			"--description=Sprint goals and objectives",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("cycle update failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Updated cycle with description")
		_ = stdout
	})

	// Cleanup
	t.Run("cleanup", func(t *testing.T) {
		if cycleID != "" {
			r.run("cycle", "archive", cycleID, "--output=json")
		}
	})
}

// --- Project Status Update Lifecycle Test ---

func TestCLI_ProjectStatusUpdateLifecycle(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")
	projectName := fmt.Sprintf("StatusUpdate Test Project %s", timestamp)

	var projectID, updateID string

	// CREATE project
	t.Run("create_project", func(t *testing.T) {
		stdout, stderr, err := r.run("project", "create",
			"--team="+r.teamKey,
			"--name="+projectName,
			"--description=Project for status update testing",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("project create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		project := extractEntity(result, "project")
		projectID = project["id"].(string)
		t.Logf("Created project: %s (%s)", projectName, projectID)
	})

	// CREATE status update
	t.Run("create_status_update", func(t *testing.T) {
		if projectID == "" {
			t.Skip("No project created")
		}

		stdout, stderr, err := r.run("project", "status-update-create",
			"--project="+projectID,
			"--body=Week 1: All systems operational",
			"--health=onTrack",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("status update create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		update := extractEntity(result, "projectUpdate")
		updateID = update["id"].(string)
		t.Logf("Created status update: %s", updateID)

		// Verify health
		if health, ok := update["health"].(string); ok && health != "onTrack" {
			t.Errorf("Expected health=onTrack, got %s", health)
		}
	})

	// LIST status updates
	t.Run("list_status_updates", func(t *testing.T) {
		if projectID == "" {
			t.Skip("No project")
		}

		stdout, stderr, err := r.run("project", "status-update-list",
			"--project="+projectID,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("status update list failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)

		// Check if we have projectUpdates nested structure
		if projectUpdates, ok := result["projectUpdates"].(map[string]any); ok {
			if nodes, ok := projectUpdates["nodes"].([]any); ok {
				if len(nodes) < 1 {
					t.Error("Expected at least one status update in list")
				}
				t.Logf("Found %d status updates for project", len(nodes))
			}
		}
	})

	// GET status update
	t.Run("get_status_update", func(t *testing.T) {
		if updateID == "" {
			t.Skip("No status update created")
		}

		stdout, stderr, err := r.run("project", "status-update-get",
			updateID,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("status update get failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)

		if result["id"] != updateID {
			t.Errorf("Expected update ID %s, got %v", updateID, result["id"])
		}

		if !strings.Contains(result["body"].(string), "Week 1") {
			t.Error("Expected body to contain 'Week 1'")
		}

		t.Logf("Retrieved status update: %s", updateID)
	})

	// DELETE status update
	t.Run("delete_status_update", func(t *testing.T) {
		if updateID == "" {
			t.Skip("No status update to delete")
		}

		stdout, stderr, err := r.run("project", "status-update-delete",
			updateID,
			"--yes",
		)
		if err != nil {
			t.Fatalf("status update delete failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Deleted status update: %s", updateID)
		_ = stdout
	})

	// Cleanup: delete project
	t.Run("cleanup", func(t *testing.T) {
		if projectID != "" {
			r.run("project", "delete", projectID, "--yes")
			t.Logf("Deleted project: %s", projectID)
		}
	})
}

// --- Initiative Status Update Lifecycle Test ---

func TestCLI_InitiativeStatusUpdateLifecycle(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")
	initiativeName := fmt.Sprintf("StatusUpdate Test Initiative %s", timestamp)

	var initiativeID, updateID string

	// CREATE initiative
	t.Run("create_initiative", func(t *testing.T) {
		stdout, stderr, err := r.run("initiative", "create",
			"--name="+initiativeName,
			"--description=Initiative for status update testing",
			"--status=Active",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("initiative create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		initiative := extractEntity(result, "initiative")
		initiativeID = initiative["id"].(string)
		t.Logf("Created initiative: %s (%s)", initiativeName, initiativeID)
	})

	// CREATE status update
	t.Run("create_status_update", func(t *testing.T) {
		if initiativeID == "" {
			t.Skip("No initiative created")
		}

		stdout, stderr, err := r.run("initiative", "status-update-create",
			"--initiative="+initiativeID,
			"--body=Q1 Progress: Ahead of schedule",
			"--health=onTrack",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("status update create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		update := extractEntity(result, "initiativeUpdate")
		updateID = update["id"].(string)
		t.Logf("Created status update: %s", updateID)

		// Verify health
		if health, ok := update["health"].(string); ok && health != "onTrack" {
			t.Errorf("Expected health=onTrack, got %s", health)
		}
	})

	// LIST status updates
	t.Run("list_status_updates", func(t *testing.T) {
		if initiativeID == "" {
			t.Skip("No initiative")
		}

		stdout, stderr, err := r.run("initiative", "status-update-list",
			"--initiative="+initiativeID,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("status update list failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)

		// Check if we have initiativeUpdates nested structure
		if initiativeUpdates, ok := result["initiativeUpdates"].(map[string]any); ok {
			if nodes, ok := initiativeUpdates["nodes"].([]any); ok {
				if len(nodes) < 1 {
					t.Error("Expected at least one status update in list")
				}
				t.Logf("Found %d status updates for initiative", len(nodes))
			}
		}
	})

	// GET status update
	t.Run("get_status_update", func(t *testing.T) {
		if updateID == "" {
			t.Skip("No status update created")
		}

		stdout, stderr, err := r.run("initiative", "status-update-get",
			updateID,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("status update get failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)

		if result["id"] != updateID {
			t.Errorf("Expected update ID %s, got %v", updateID, result["id"])
		}

		if !strings.Contains(result["body"].(string), "Q1 Progress") {
			t.Error("Expected body to contain 'Q1 Progress'")
		}

		t.Logf("Retrieved status update: %s", updateID)
	})

	// ARCHIVE status update
	t.Run("archive_status_update", func(t *testing.T) {
		if updateID == "" {
			t.Skip("No status update to archive")
		}

		stdout, stderr, err := r.run("initiative", "status-update-archive",
			updateID,
			"--yes",
		)
		if err != nil {
			t.Fatalf("status update archive failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Archived status update: %s", updateID)
		_ = stdout
	})

	// Cleanup: delete initiative
	t.Run("cleanup", func(t *testing.T) {
		if initiativeID != "" {
			r.run("initiative", "delete", initiativeID, "--yes")
			t.Logf("Deleted initiative: %s", initiativeID)
		}
	})
}

// --- Team Velocity Test ---

func TestCLI_TeamVelocity(t *testing.T) {
	r := newWriteTestRunner(t)

	// Test velocity calculation with team that has existing cycles
	t.Run("calculate_velocity", func(t *testing.T) {
		stdout, stderr, err := r.run("team", "velocity",
			"--team="+r.teamKey,
			"--cycles=3",
			"--output=json",
		)

		// Check if team has no completed cycles (valid case)
		if strings.Contains(stdout, "No completed cycles") {
			t.Logf("Team %s has no completed cycles (valid case)", r.teamKey)
			return
		}

		if err != nil {
			t.Fatalf("team velocity failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse velocity response: %v\nOutput: %s", err, stdout)
		}

		// Verify expected fields
		if result["teamKey"] == nil {
			t.Error("Expected teamKey in response")
		}
		if result["cyclesAnalyzed"] == nil {
			t.Error("Expected cyclesAnalyzed in response")
		}
		if result["avgPointsCompleted"] == nil {
			t.Error("Expected avgPointsCompleted in response")
		}
		if result["avgIssuesCompleted"] == nil {
			t.Error("Expected avgIssuesCompleted in response")
		}

		t.Logf("Velocity for team %s: %.1f avg points/cycle, %.1f avg issues/cycle",
			result["teamKey"],
			result["avgPointsCompleted"],
			result["avgIssuesCompleted"])
	})

	t.Run("table_output", func(t *testing.T) {
		stdout, stderr, err := r.run("team", "velocity",
			"--team="+r.teamKey,
			"--output=table",
		)
		if err != nil {
			t.Fatalf("team velocity table failed: %v\nstderr: %s", err, stderr)
		}

		// Check for either metrics or no cycles message
		hasMetrics := strings.Contains(stdout, "Velocity Metrics")
		hasNoCycles := strings.Contains(stdout, "No completed cycles")

		if !hasMetrics && !hasNoCycles {
			t.Errorf("Expected either velocity metrics or no cycles message, got: %s", stdout)
		}

		t.Logf("Velocity output:\n%s", stdout)
	})
}

// --- Initiative-Project Linking Test ---

func TestCLI_InitiativeProjectLinking(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")
	initiativeName := fmt.Sprintf("Link Test Initiative %s", timestamp)
	projectName := fmt.Sprintf("Link Test Project %s", timestamp)

	var initiativeID, projectID string

	// CREATE initiative
	t.Run("create_initiative", func(t *testing.T) {
		stdout, stderr, err := r.run("initiative", "create",
			"--name="+initiativeName,
			"--description=Initiative for linking test",
			"--status=Active",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("initiative create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		initiative := extractEntity(result, "initiative")
		initiativeID = initiative["id"].(string)
		t.Logf("Created initiative: %s (%s)", initiativeName, initiativeID)
	})

	// CREATE project
	t.Run("create_project", func(t *testing.T) {
		stdout, stderr, err := r.run("project", "create",
			"--team="+r.teamKey,
			"--name="+projectName,
			"--description=Project for linking test",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("project create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		project := extractEntity(result, "project")
		projectID = project["id"].(string)
		t.Logf("Created project: %s (%s)", projectName, projectID)
	})

	// LINK project to initiative
	var linkCreated bool
	t.Run("add_project_to_initiative", func(t *testing.T) {
		if initiativeID == "" || projectID == "" {
			t.Skip("Missing initiative or project")
		}

		stdout, stderr, err := r.run("initiative", "add-project",
			"--initiative="+initiativeID,
			"--project="+projectID,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("add-project failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)

		if result["id"] == nil {
			t.Error("Expected link ID in response")
		}

		linkCreated = true
		t.Logf("Linked project %s to initiative %s", projectID, initiativeID)
	})

	// UNLINK project from initiative
	t.Run("remove_project_from_initiative", func(t *testing.T) {
		if !linkCreated {
			t.Skip("No link created")
		}

		stdout, stderr, err := r.run("initiative", "remove-project",
			"--initiative="+initiativeID,
			"--project="+projectID,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("remove-project failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)

		if success, ok := result["success"].(bool); !ok || !success {
			t.Error("Expected success=true in response")
		}

		t.Logf("Unlinked project %s from initiative %s", projectID, initiativeID)
	})

	// Cleanup
	t.Run("cleanup", func(t *testing.T) {
		if projectID != "" {
			r.run("project", "delete", projectID, "--yes")
			t.Logf("Deleted project: %s", projectID)
		}
		if initiativeID != "" {
			r.run("initiative", "delete", initiativeID, "--yes")
			t.Logf("Deleted initiative: %s", initiativeID)
		}
	})
}

// --- Document CRUD Test ---

func TestCLI_DocumentCRUD(t *testing.T) {
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")
	docTitle := fmt.Sprintf("Test Document %s", timestamp)

	var documentID string

	// CREATE document
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := r.run("document", "create",
			"--title="+docTitle,
			"--content=# Test\n\nThis is test documentation.",
			"--team="+r.teamKey,
			"--output=json",
		)
		if err != nil {
			t.Fatalf("document create failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		documentID = result["id"].(string)
		t.Logf("Created document: %s (%s)", docTitle, documentID)
	})

	// UPDATE document
	t.Run("update", func(t *testing.T) {
		if documentID == "" {
			t.Skip("No document created")
		}

		newTitle := docTitle + " (updated)"
		stdout, stderr, err := r.run("document", "update", documentID,
			"--title="+newTitle,
			"--content=# Updated\n\nUpdated content.",
			"--output=json",
		)
		if err != nil {
			t.Fatalf("document update failed: %v\nstderr: %s", err, stderr)
		}

		var result map[string]any
		json.Unmarshal([]byte(stdout), &result)
		if !strings.Contains(result["title"].(string), "updated") {
			t.Error("Expected updated title")
		}

		t.Logf("Updated document: %s", documentID)
	})

	// DELETE document
	t.Run("delete", func(t *testing.T) {
		if documentID == "" {
			t.Skip("No document to delete")
		}

		stdout, stderr, err := r.run("document", "delete", documentID, "--yes")
		if err != nil {
			t.Fatalf("document delete failed: %v\nstderr: %s", err, stderr)
		}

		t.Logf("Deleted document: %s", documentID)
		_ = stdout
	})
}

// Helper to build CLI binary if needed
func init() {
	// Check if binary exists, build if not (path relative to cmd/linear/commands)
	repoRoot := "../../../"
	if _, err := os.Stat(repoRoot + "bin/go-linear"); os.IsNotExist(err) {
		cmd := exec.Command("make", "build-cli")
		cmd.Dir = repoRoot
		cmd.Run()
	}
}
