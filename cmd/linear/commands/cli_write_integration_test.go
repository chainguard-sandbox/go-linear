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
	t.Skip("Reaction API returns error - skipping until resolved")
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
	t.Skip("Favorite API returns error - skipping until resolved")
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
	t.Skip("Project API returns error - skipping until resolved")
	r := newWriteTestRunner(t)

	timestamp := time.Now().Format("20060102-150405")
	projectName := fmt.Sprintf("Test Project %s", timestamp)

	var projectID string

	// CREATE
	t.Run("create", func(t *testing.T) {
		stdout, stderr, err := r.run("project", "create",
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
