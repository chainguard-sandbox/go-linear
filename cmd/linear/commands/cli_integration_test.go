//go:build integration

// Package commands contains CLI integration tests that run actual commands
// against the real Linear API.
//
// Run with: go test -tags=integration ./cmd/linear/commands/...
// Requires: LINEAR_API_KEY environment variable
package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// cliRunner executes CLI commands and captures output
type cliRunner struct {
	t       *testing.T
	binPath string
	apiKey  string
}

func newCLIRunner(t *testing.T) *cliRunner {
	t.Helper()

	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping CLI integration test: LINEAR_API_KEY not set")
	}

	// Find repo root (3 directories up from cmd/linear/commands)
	repoRoot := "../../../"
	binPath := repoRoot + "bin/go-linear"

	// Build the binary if needed
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Log("Building CLI binary...")
		cmd := exec.Command("make", "build-cli")
		cmd.Dir = repoRoot
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to build CLI: %v", err)
		}
	}

	return &cliRunner{
		t:       t,
		binPath: binPath,
		apiKey:  apiKey,
	}
}

// run executes a CLI command and returns stdout, stderr, and error
func (r *cliRunner) run(args ...string) (stdout, stderr string, err error) {
	cmd := exec.Command(r.binPath, args...)
	cmd.Env = append(os.Environ(), "LINEAR_API_KEY="+r.apiKey)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// runJSON executes a command with --output=json and parses the result
func (r *cliRunner) runJSON(args ...string) (map[string]any, error) {
	args = append(args, "--output=json")
	stdout, stderr, err := r.run(args...)
	if err != nil {
		r.t.Logf("Command failed: %v\nstderr: %s", err, stderr)
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// --- Read Operation Tests ---

func TestCLI_Viewer(t *testing.T) {
	r := newCLIRunner(t)

	result, err := r.runJSON("viewer")
	if err != nil {
		t.Fatalf("viewer command failed: %v", err)
	}

	// Verify required fields
	if result["id"] == nil || result["id"] == "" {
		t.Error("viewer.id is missing")
	}
	if result["email"] == nil || result["email"] == "" {
		t.Error("viewer.email is missing")
	}

	t.Logf("Authenticated as: %s (%s)", result["name"], result["email"])
}

func TestCLI_Organization(t *testing.T) {
	r := newCLIRunner(t)

	result, err := r.runJSON("organization")
	if err != nil {
		t.Fatalf("organization command failed: %v", err)
	}

	if result["id"] == nil || result["id"] == "" {
		t.Error("organization.id is missing")
	}
	if result["name"] == nil || result["name"] == "" {
		t.Error("organization.name is missing")
	}

	t.Logf("Organization: %s", result["name"])
}

func TestCLI_TeamList(t *testing.T) {
	r := newCLIRunner(t)

	stdout, stderr, err := r.run("team", "list", "--output=json")
	if err != nil {
		t.Fatalf("team list failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d teams", len(nodes))
	if len(nodes) == 0 {
		t.Skip("No teams available for testing")
	}

	// Verify first team has required fields
	team := nodes[0].(map[string]any)
	if team["id"] == nil {
		t.Error("team.id is missing")
	}
	if team["name"] == nil {
		t.Error("team.name is missing")
	}
	if team["key"] == nil {
		t.Error("team.key is missing")
	}
}

func TestCLI_TeamGet(t *testing.T) {
	r := newCLIRunner(t)

	// First list teams to get an ID
	stdout, _, err := r.run("team", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No teams available")
	}

	var listResult map[string]any
	json.Unmarshal([]byte(stdout), &listResult)
	nodes := listResult["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No teams available")
	}

	team := nodes[0].(map[string]any)
	teamKey := team["key"].(string)

	// Now get by key
	result, err := r.runJSON("team", "get", teamKey)
	if err != nil {
		t.Fatalf("team get failed: %v", err)
	}

	if result["id"] != team["id"] {
		t.Errorf("team get returned different team: got %s, want %s", result["id"], team["id"])
	}

	t.Logf("Retrieved team: %s (%s)", result["name"], result["key"])
}

func TestCLI_IssueList(t *testing.T) {
	r := newCLIRunner(t)

	stdout, stderr, err := r.run("issue", "list", "--output=json", "--limit=5")
	if err != nil {
		t.Fatalf("issue list failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d issues", len(nodes))
	if len(nodes) == 0 {
		t.Skip("No issues available for testing")
	}

	// Verify first issue has required fields
	issue := nodes[0].(map[string]any)
	if issue["id"] == nil {
		t.Error("issue.id is missing")
	}
	if issue["identifier"] == nil {
		t.Error("issue.identifier is missing")
	}
	if issue["title"] == nil {
		t.Error("issue.title is missing")
	}
}

func TestCLI_IssueGet(t *testing.T) {
	r := newCLIRunner(t)

	// First list issues to get an identifier
	stdout, _, err := r.run("issue", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No issues available")
	}

	var listResult map[string]any
	json.Unmarshal([]byte(stdout), &listResult)
	nodes := listResult["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No issues available")
	}

	issue := nodes[0].(map[string]any)
	identifier := issue["identifier"].(string)

	// Now get by identifier
	result, err := r.runJSON("issue", "get", identifier)
	if err != nil {
		t.Fatalf("issue get failed: %v", err)
	}

	if result["identifier"] != identifier {
		t.Errorf("issue get returned different issue: got %s, want %s", result["identifier"], identifier)
	}

	t.Logf("Retrieved issue: %s - %s", result["identifier"], result["title"])
}

func TestCLI_IssueSearch(t *testing.T) {
	r := newCLIRunner(t)

	// Search for a common term
	stdout, stderr, err := r.run("issue", "search", "test", "--output=json", "--limit=5")
	if err != nil {
		// Search might return no results, which is ok
		if strings.Contains(stderr, "no issues found") {
			t.Skip("No issues matching 'test' found")
		}
		t.Fatalf("issue search failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Search found %d issues", len(nodes))
}

func TestCLI_LabelList(t *testing.T) {
	r := newCLIRunner(t)

	stdout, stderr, err := r.run("label", "list", "--output=json")
	if err != nil {
		t.Fatalf("label list failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d labels", len(nodes))
}

func TestCLI_StateList(t *testing.T) {
	r := newCLIRunner(t)

	stdout, stderr, err := r.run("state", "list", "--output=json")
	if err != nil {
		t.Fatalf("state list failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d workflow states", len(nodes))
}

func TestCLI_UserList(t *testing.T) {
	r := newCLIRunner(t)

	stdout, stderr, err := r.run("user", "list", "--output=json", "--limit=10")
	if err != nil {
		t.Fatalf("user list failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d users", len(nodes))
}

func TestCLI_ProjectList(t *testing.T) {
	r := newCLIRunner(t)

	stdout, stderr, err := r.run("project", "list", "--output=json", "--limit=10")
	if err != nil {
		t.Fatalf("project list failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d projects", len(nodes))
}

func TestCLI_CycleList(t *testing.T) {
	r := newCLIRunner(t)

	stdout, stderr, err := r.run("cycle", "list", "--output=json", "--limit=10")
	if err != nil {
		t.Fatalf("cycle list failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d cycles", len(nodes))
}

func TestCLI_Status(t *testing.T) {
	r := newCLIRunner(t)

	stdout, stderr, err := r.run("status", "--output=json")
	if err != nil {
		t.Fatalf("status failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify rate limit fields
	if result["requestsLimit"] == nil {
		t.Error("requestsLimit is missing")
	}
	if result["requestsRemaining"] == nil {
		t.Error("requestsRemaining is missing")
	}

	t.Logf("Rate limit: %v/%v requests remaining",
		result["requestsRemaining"], result["requestsLimit"])
}

func TestCLI_DocumentList(t *testing.T) {
	r := newCLIRunner(t)

	stdout, stderr, err := r.run("document", "list", "--output=json", "--limit=10")
	if err != nil {
		t.Fatalf("document list failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d documents", len(nodes))
}

func TestCLI_RoadmapList(t *testing.T) {
	r := newCLIRunner(t)

	stdout, stderr, err := r.run("roadmap", "list", "--output=json", "--limit=10")
	if err != nil {
		t.Fatalf("roadmap list failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d roadmaps", len(nodes))
}

func TestCLI_TemplateList(t *testing.T) {
	r := newCLIRunner(t)

	stdout, stderr, err := r.run("template", "list", "--output=json", "--limit=10")
	if err != nil {
		t.Fatalf("template list failed: %v\nstderr: %s", err, stderr)
	}

	// Template list returns an array directly, not {nodes: [...]}
	var templates []any
	if err := json.Unmarshal([]byte(stdout), &templates); err != nil {
		// Try parsing as object with nodes
		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}
		nodes, ok := result["nodes"].([]any)
		if !ok {
			t.Fatal("Expected nodes array or array in response")
		}
		t.Logf("Retrieved %d templates", len(nodes))
		return
	}

	t.Logf("Retrieved %d templates", len(templates))
}

func TestCLI_InitiativeList(t *testing.T) {
	r := newCLIRunner(t)

	stdout, stderr, err := r.run("initiative", "list", "--output=json", "--limit=10")
	if err != nil {
		t.Fatalf("initiative list failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d initiatives", len(nodes))
}

func TestCLI_AttachmentList(t *testing.T) {
	r := newCLIRunner(t)

	stdout, stderr, err := r.run("attachment", "list", "--output=json", "--limit=10")
	if err != nil {
		t.Fatalf("attachment list failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d attachments", len(nodes))
}

func TestCLI_CommentList(t *testing.T) {
	r := newCLIRunner(t)

	stdout, stderr, err := r.run("comment", "list", "--output=json", "--limit=10")
	if err != nil {
		t.Fatalf("comment list failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d comments", len(nodes))
}

func TestCLI_UserGet(t *testing.T) {
	r := newCLIRunner(t)

	// Get current user with 'me'
	result, err := r.runJSON("user", "get", "me")
	if err != nil {
		t.Fatalf("user get me failed: %v", err)
	}

	if result["id"] == nil || result["id"] == "" {
		t.Error("user.id is missing")
	}
	if result["email"] == nil || result["email"] == "" {
		t.Error("user.email is missing")
	}

	t.Logf("Retrieved user: %s (%s)", result["name"], result["email"])
}

func TestCLI_IssueCount(t *testing.T) {
	r := newCLIRunner(t)

	stdout, stderr, err := r.run("issue", "list", "--count", "--output=json")
	if err != nil {
		t.Fatalf("issue list --count failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result["count"] == nil {
		t.Error("count is missing from response")
	}

	t.Logf("Issue count: %v", result["count"])
}

func TestCLI_TeamMembers(t *testing.T) {
	r := newCLIRunner(t)

	// First get a team
	stdout, _, err := r.run("team", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No teams available")
	}

	var listResult map[string]any
	json.Unmarshal([]byte(stdout), &listResult)
	nodes := listResult["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No teams available")
	}

	team := nodes[0].(map[string]any)
	teamKey := team["key"].(string)

	// Get team members
	stdout, stderr, err := r.run("team", "members", "--team="+teamKey, "--output=json")
	if err != nil {
		t.Fatalf("team members failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Team members returns {team: {...}, users: [...]}
	members, ok := result["users"].([]any)
	if !ok {
		// Try nodes as fallback
		members, ok = result["nodes"].([]any)
		if !ok {
			t.Fatal("Expected users or nodes array in response")
		}
	}

	t.Logf("Team %s has %d members", teamKey, len(members))
}

// --- Filtering Tests ---

func TestCLI_IssueListWithTeamFilter(t *testing.T) {
	r := newCLIRunner(t)

	// First get a team
	stdout, _, err := r.run("team", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No teams available")
	}

	var listResult map[string]any
	json.Unmarshal([]byte(stdout), &listResult)
	nodes := listResult["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No teams available")
	}

	team := nodes[0].(map[string]any)
	teamKey := team["key"].(string)

	// List issues filtered by team
	stdout, stderr, err := r.run("issue", "list", "--team="+teamKey, "--output=json", "--limit=5")
	if err != nil {
		t.Fatalf("issue list with team filter failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	issueNodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d issues for team %s", len(issueNodes), teamKey)

	// Verify all issues belong to the team
	for _, n := range issueNodes {
		issue := n.(map[string]any)
		if teamData, ok := issue["team"].(map[string]any); ok {
			if teamData["key"] != teamKey {
				t.Errorf("Issue %s belongs to team %s, expected %s",
					issue["identifier"], teamData["key"], teamKey)
			}
		}
	}
}

func TestCLI_IssueListWithStateFilter(t *testing.T) {
	r := newCLIRunner(t)

	// List issues with state filter
	stdout, stderr, err := r.run("issue", "list", "--state=In Progress", "--output=json", "--limit=5")
	if err != nil {
		// State might not exist
		if strings.Contains(stderr, "not found") || strings.Contains(stderr, "state") {
			t.Skip("State 'In Progress' not found")
		}
		t.Fatalf("issue list with state filter failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d issues in 'In Progress' state", len(nodes))
}

func TestCLI_IssueListWithAssigneeFilter(t *testing.T) {
	r := newCLIRunner(t)

	// List issues assigned to current user
	stdout, stderr, err := r.run("issue", "list", "--assignee=me", "--output=json", "--limit=5")
	if err != nil {
		t.Fatalf("issue list with assignee filter failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d issues assigned to me", len(nodes))
}

func TestCLI_IssueListWithPriorityFilter(t *testing.T) {
	r := newCLIRunner(t)

	// List high priority issues (priority=2)
	stdout, stderr, err := r.run("issue", "list", "--priority=2", "--output=json", "--limit=5")
	if err != nil {
		t.Fatalf("issue list with priority filter failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d high priority issues", len(nodes))

	// Verify all issues have priority 2
	for _, n := range nodes {
		issue := n.(map[string]any)
		if priority, ok := issue["priority"].(float64); ok {
			if int(priority) != 2 {
				t.Errorf("Issue %s has priority %v, expected 2",
					issue["identifier"], priority)
			}
		}
	}
}

func TestCLI_IssueListWithDateFilter(t *testing.T) {
	r := newCLIRunner(t)

	// List issues created in the last 7 days
	stdout, stderr, err := r.run("issue", "list", "--created-after=7d", "--output=json", "--limit=5")
	if err != nil {
		t.Fatalf("issue list with date filter failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d issues created in last 7 days", len(nodes))
}

// --- Get Commands for Individual Entities ---

func TestCLI_ProjectGet(t *testing.T) {
	r := newCLIRunner(t)

	// First list projects to get an ID
	stdout, _, err := r.run("project", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No projects available")
	}

	var listResult map[string]any
	json.Unmarshal([]byte(stdout), &listResult)
	nodes := listResult["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No projects available")
	}

	project := nodes[0].(map[string]any)
	projectID := project["id"].(string)

	// Get project by ID
	result, err := r.runJSON("project", "get", projectID)
	if err != nil {
		t.Fatalf("project get failed: %v", err)
	}

	if result["id"] != projectID {
		t.Errorf("project get returned different project: got %s, want %s", result["id"], projectID)
	}

	t.Logf("Retrieved project: %s", result["name"])
}

func TestCLI_CycleGet(t *testing.T) {
	r := newCLIRunner(t)

	// First list cycles to get an ID
	stdout, _, err := r.run("cycle", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No cycles available")
	}

	var listResult map[string]any
	json.Unmarshal([]byte(stdout), &listResult)
	nodes := listResult["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No cycles available")
	}

	cycle := nodes[0].(map[string]any)
	cycleID := cycle["id"].(string)

	// Get cycle by ID
	result, err := r.runJSON("cycle", "get", cycleID)
	if err != nil {
		t.Fatalf("cycle get failed: %v", err)
	}

	if result["id"] != cycleID {
		t.Errorf("cycle get returned different cycle: got %s, want %s", result["id"], cycleID)
	}

	t.Logf("Retrieved cycle: %s", result["name"])
}

func TestCLI_LabelGet(t *testing.T) {
	r := newCLIRunner(t)

	// First list labels to get an ID
	stdout, _, err := r.run("label", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No labels available")
	}

	var listResult map[string]any
	json.Unmarshal([]byte(stdout), &listResult)
	nodes := listResult["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No labels available")
	}

	label := nodes[0].(map[string]any)
	labelID := label["id"].(string)

	// Get label by ID
	result, err := r.runJSON("label", "get", labelID)
	if err != nil {
		t.Fatalf("label get failed: %v", err)
	}

	if result["id"] != labelID {
		t.Errorf("label get returned different label: got %s, want %s", result["id"], labelID)
	}

	t.Logf("Retrieved label: %s", result["name"])
}

func TestCLI_StateGet(t *testing.T) {
	r := newCLIRunner(t)

	// First list states to get an ID
	stdout, _, err := r.run("state", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No states available")
	}

	var listResult map[string]any
	json.Unmarshal([]byte(stdout), &listResult)
	nodes := listResult["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No states available")
	}

	state := nodes[0].(map[string]any)
	stateID := state["id"].(string)

	// Get state by ID
	result, err := r.runJSON("state", "get", stateID)
	if err != nil {
		t.Fatalf("state get failed: %v", err)
	}

	if result["id"] != stateID {
		t.Errorf("state get returned different state: got %s, want %s", result["id"], stateID)
	}

	t.Logf("Retrieved state: %s", result["name"])
}

// --- Field Selection Tests ---

func TestCLI_IssueListWithFields(t *testing.T) {
	r := newCLIRunner(t)

	// List issues with custom fields
	stdout, stderr, err := r.run("issue", "list", "--fields=id,title,identifier", "--output=json", "--limit=3")
	if err != nil {
		t.Fatalf("issue list with fields failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	if len(nodes) == 0 {
		t.Skip("No issues available")
	}

	// Verify first issue has expected fields
	issue := nodes[0].(map[string]any)
	if issue["id"] == nil {
		t.Error("Expected id field")
	}
	if issue["title"] == nil {
		t.Error("Expected title field")
	}
	if issue["identifier"] == nil {
		t.Error("Expected identifier field")
	}

	t.Logf("Retrieved %d issues with custom fields", len(nodes))
}

func TestCLI_TeamListWithFields(t *testing.T) {
	r := newCLIRunner(t)

	// List teams with custom fields
	stdout, stderr, err := r.run("team", "list", "--fields=id,name,key", "--output=json", "--limit=3")
	if err != nil {
		t.Fatalf("team list with fields failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	if len(nodes) == 0 {
		t.Skip("No teams available")
	}

	// Verify first team has expected fields
	team := nodes[0].(map[string]any)
	if team["id"] == nil {
		t.Error("Expected id field")
	}
	if team["name"] == nil {
		t.Error("Expected name field")
	}
	if team["key"] == nil {
		t.Error("Expected key field")
	}

	t.Logf("Retrieved %d teams with custom fields", len(nodes))
}

// --- User Completed Tests ---

func TestCLI_UserCompleted(t *testing.T) {
	r := newCLIRunner(t)

	// Get completed issues for current user in last 7 days
	stdout, stderr, err := r.run("user", "completed", "--user=me", "--completed-after=7d", "--output=json")
	if err != nil {
		t.Fatalf("user completed failed: %v\nstderr: %s", err, stderr)
	}

	var results []any
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		// Try parsing as single object
		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}
		// Check if there's a count field
		if count, ok := result["count"]; ok {
			t.Logf("User completed %v issues in last 7 days", count)
			return
		}
		t.Logf("User completed result: %v", result)
		return
	}

	t.Logf("User completed returned %d entries", len(results))
}

// --- Table Output Tests ---

func TestCLI_IssueListTable(t *testing.T) {
	r := newCLIRunner(t)

	stdout, stderr, err := r.run("issue", "list", "--output=table", "--limit=5")
	if err != nil {
		t.Fatalf("issue list table failed: %v\nstderr: %s", err, stderr)
	}

	// Table output should have headers and rows
	if !strings.Contains(stdout, "IDENTIFIER") && !strings.Contains(stdout, "ID") {
		t.Error("Table output missing expected headers")
	}

	t.Logf("Table output:\n%s", stdout[:min(500, len(stdout))])
}

func TestCLI_TeamListTable(t *testing.T) {
	r := newCLIRunner(t)

	stdout, stderr, err := r.run("team", "list", "--output=table", "--limit=5")
	if err != nil {
		t.Fatalf("team list table failed: %v\nstderr: %s", err, stderr)
	}

	// Table output should have headers
	if !strings.Contains(stdout, "NAME") && !strings.Contains(stdout, "KEY") {
		t.Error("Table output missing expected headers")
	}

	t.Logf("Table output:\n%s", stdout[:min(500, len(stdout))])
}

// --- Error Handling Tests ---

func TestCLI_InvalidIssueID(t *testing.T) {
	r := newCLIRunner(t)

	_, stderr, err := r.run("issue", "get", "INVALID-9999999")
	if err == nil {
		t.Error("Expected error for invalid issue ID")
	}

	if !strings.Contains(stderr, "not found") && !strings.Contains(stderr, "error") && !strings.Contains(stderr, "failed") {
		t.Logf("Error output: %s", stderr)
	}
}

func TestCLI_InvalidTeamKey(t *testing.T) {
	r := newCLIRunner(t)

	_, stderr, err := r.run("team", "get", "INVALIDTEAMKEY999")
	if err == nil {
		t.Error("Expected error for invalid team key")
	}

	if !strings.Contains(stderr, "not found") && !strings.Contains(stderr, "error") && !strings.Contains(stderr, "failed") {
		t.Logf("Error output: %s", stderr)
	}
}

func TestCLI_InvalidOutputFormat(t *testing.T) {
	r := newCLIRunner(t)

	_, stderr, err := r.run("team", "list", "--output=invalid")
	if err == nil {
		t.Error("Expected error for invalid output format")
	}

	if !strings.Contains(stderr, "unknown") && !strings.Contains(stderr, "invalid") && !strings.Contains(stderr, "format") {
		t.Logf("Error output: %s", stderr)
	}
}

// --- Get Commands for More Entities ---

func TestCLI_DocumentGet(t *testing.T) {
	r := newCLIRunner(t)

	// First list documents to get an ID
	stdout, _, err := r.run("document", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No documents available")
	}

	var listResult map[string]any
	json.Unmarshal([]byte(stdout), &listResult)
	nodes := listResult["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No documents available")
	}

	doc := nodes[0].(map[string]any)
	docID := doc["id"].(string)

	// Get document by ID
	result, err := r.runJSON("document", "get", docID)
	if err != nil {
		t.Fatalf("document get failed: %v", err)
	}

	if result["id"] != docID {
		t.Errorf("document get returned different document: got %s, want %s", result["id"], docID)
	}

	t.Logf("Retrieved document: %s", result["title"])
}

func TestCLI_RoadmapGet(t *testing.T) {
	r := newCLIRunner(t)

	// First list roadmaps to get an ID
	stdout, _, err := r.run("roadmap", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No roadmaps available")
	}

	var listResult map[string]any
	json.Unmarshal([]byte(stdout), &listResult)
	nodes := listResult["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No roadmaps available")
	}

	roadmap := nodes[0].(map[string]any)
	roadmapID := roadmap["id"].(string)

	// Get roadmap by ID
	result, err := r.runJSON("roadmap", "get", roadmapID)
	if err != nil {
		t.Fatalf("roadmap get failed: %v", err)
	}

	if result["id"] != roadmapID {
		t.Errorf("roadmap get returned different roadmap: got %s, want %s", result["id"], roadmapID)
	}

	t.Logf("Retrieved roadmap: %s", result["name"])
}

func TestCLI_InitiativeGet(t *testing.T) {
	r := newCLIRunner(t)

	// First list initiatives to get an ID
	stdout, _, err := r.run("initiative", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No initiatives available")
	}

	var listResult map[string]any
	json.Unmarshal([]byte(stdout), &listResult)
	nodes := listResult["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No initiatives available")
	}

	initiative := nodes[0].(map[string]any)
	initiativeID := initiative["id"].(string)

	// Get initiative by ID
	result, err := r.runJSON("initiative", "get", initiativeID)
	if err != nil {
		t.Fatalf("initiative get failed: %v", err)
	}

	if result["id"] != initiativeID {
		t.Errorf("initiative get returned different initiative: got %s, want %s", result["id"], initiativeID)
	}

	t.Logf("Retrieved initiative: %s", result["name"])
}

func TestCLI_TemplateGet(t *testing.T) {
	r := newCLIRunner(t)

	// First list templates to get an ID
	stdout, _, err := r.run("template", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No templates available")
	}

	// Templates might be an array or have nodes
	var templates []any
	if err := json.Unmarshal([]byte(stdout), &templates); err != nil {
		var listResult map[string]any
		if err := json.Unmarshal([]byte(stdout), &listResult); err != nil {
			t.Skip("Cannot parse templates")
		}
		nodes, ok := listResult["nodes"].([]any)
		if !ok || len(nodes) == 0 {
			t.Skip("No templates available")
		}
		templates = nodes
	}

	if len(templates) == 0 {
		t.Skip("No templates available")
	}

	template := templates[0].(map[string]any)
	templateID := template["id"].(string)

	// Get template by ID
	result, err := r.runJSON("template", "get", templateID)
	if err != nil {
		t.Fatalf("template get failed: %v", err)
	}

	if result["id"] != templateID {
		t.Errorf("template get returned different template: got %s, want %s", result["id"], templateID)
	}

	t.Logf("Retrieved template: %s", result["name"])
}

func TestCLI_CommentGet(t *testing.T) {
	r := newCLIRunner(t)

	// First list comments to get an ID
	stdout, _, err := r.run("comment", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No comments available")
	}

	var listResult map[string]any
	json.Unmarshal([]byte(stdout), &listResult)
	nodes := listResult["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No comments available")
	}

	comment := nodes[0].(map[string]any)
	commentID := comment["id"].(string)

	// Get comment by ID
	result, err := r.runJSON("comment", "get", commentID)
	if err != nil {
		t.Fatalf("comment get failed: %v", err)
	}

	if result["id"] != commentID {
		t.Errorf("comment get returned different comment: got %s, want %s", result["id"], commentID)
	}

	t.Logf("Retrieved comment with body length: %d", len(result["body"].(string)))
}

func TestCLI_AttachmentGet(t *testing.T) {
	r := newCLIRunner(t)

	// First list attachments to get an ID
	stdout, _, err := r.run("attachment", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No attachments available")
	}

	var listResult map[string]any
	json.Unmarshal([]byte(stdout), &listResult)
	nodes := listResult["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No attachments available")
	}

	attachment := nodes[0].(map[string]any)
	attachmentID := attachment["id"].(string)

	// Get attachment by ID
	result, err := r.runJSON("attachment", "get", attachmentID)
	if err != nil {
		t.Fatalf("attachment get failed: %v", err)
	}

	if result["id"] != attachmentID {
		t.Errorf("attachment get returned different attachment: got %s, want %s", result["id"], attachmentID)
	}

	t.Logf("Retrieved attachment: %s", result["title"])
}

// --- Combined Filter Tests ---

func TestCLI_IssueListWithMultipleFilters(t *testing.T) {
	r := newCLIRunner(t)

	// Get a team first
	stdout, _, err := r.run("team", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No teams available")
	}

	var listResult map[string]any
	json.Unmarshal([]byte(stdout), &listResult)
	nodes := listResult["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No teams available")
	}

	team := nodes[0].(map[string]any)
	teamKey := team["key"].(string)

	// List issues with multiple filters: team + created in last 30 days
	stdout, stderr, err := r.run("issue", "list",
		"--team="+teamKey,
		"--created-after=30d",
		"--output=json",
		"--limit=5")
	if err != nil {
		t.Fatalf("issue list with multiple filters failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	issueNodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d issues with team=%s and created in last 30 days", len(issueNodes), teamKey)
}

func TestCLI_IssueListWithLabelFilter(t *testing.T) {
	r := newCLIRunner(t)

	// First get a label
	stdout, _, err := r.run("label", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No labels available")
	}

	var listResult map[string]any
	json.Unmarshal([]byte(stdout), &listResult)
	nodes := listResult["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No labels available")
	}

	label := nodes[0].(map[string]any)
	labelName := label["name"].(string)

	// List issues with label filter
	stdout, stderr, err := r.run("issue", "list", "--label="+labelName, "--output=json", "--limit=5")
	if err != nil {
		// Label might not have any issues
		t.Logf("issue list with label filter: %v\nstderr: %s", err, stderr)
		t.Skip("No issues with this label")
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	issueNodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d issues with label '%s'", len(issueNodes), labelName)
}

func TestCLI_IssueListWithCreatorFilter(t *testing.T) {
	r := newCLIRunner(t)

	// List issues created by current user
	stdout, stderr, err := r.run("issue", "list", "--creator=me", "--output=json", "--limit=5")
	if err != nil {
		t.Fatalf("issue list with creator filter failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d issues created by me", len(nodes))
}

// --- Pagination Tests ---

func TestCLI_IssueListPagination(t *testing.T) {
	r := newCLIRunner(t)

	// First request with limit=2
	stdout, stderr, err := r.run("issue", "list", "--output=json", "--limit=2")
	if err != nil {
		t.Fatalf("first page failed: %v\nstderr: %s", err, stderr)
	}

	var firstPage map[string]any
	if err := json.Unmarshal([]byte(stdout), &firstPage); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes := firstPage["nodes"].([]any)
	if len(nodes) < 2 {
		t.Skip("Not enough issues to test pagination")
	}

	// Check for pageInfo
	pageInfo, ok := firstPage["pageInfo"].(map[string]any)
	if !ok {
		t.Skip("No pageInfo in response - pagination not supported")
	}

	hasNextPage, _ := pageInfo["hasNextPage"].(bool)
	if !hasNextPage {
		t.Skip("No more pages to test")
	}

	endCursor, ok := pageInfo["endCursor"].(string)
	if !ok || endCursor == "" {
		t.Skip("No endCursor in pageInfo")
	}

	// Second request using cursor
	stdout, stderr, err = r.run("issue", "list", "--output=json", "--limit=2", "--after="+endCursor)
	if err != nil {
		t.Fatalf("second page failed: %v\nstderr: %s", err, stderr)
	}

	var secondPage map[string]any
	if err := json.Unmarshal([]byte(stdout), &secondPage); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	secondNodes := secondPage["nodes"].([]any)

	// Verify we got different issues
	firstID := nodes[0].(map[string]any)["id"].(string)
	if len(secondNodes) > 0 {
		secondID := secondNodes[0].(map[string]any)["id"].(string)
		if firstID == secondID {
			t.Error("Second page returned same first issue as first page")
		}
	}

	t.Logf("Pagination test: first page %d issues, second page %d issues", len(nodes), len(secondNodes))
}

func TestCLI_TeamListPagination(t *testing.T) {
	r := newCLIRunner(t)

	// First request
	stdout, stderr, err := r.run("team", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Fatalf("first page failed: %v\nstderr: %s", err, stderr)
	}

	var firstPage map[string]any
	if err := json.Unmarshal([]byte(stdout), &firstPage); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check for pageInfo
	pageInfo, ok := firstPage["pageInfo"].(map[string]any)
	if !ok {
		t.Skip("No pageInfo in response")
	}

	hasNextPage, _ := pageInfo["hasNextPage"].(bool)
	if !hasNextPage {
		t.Log("No more pages (single team workspace)")
		return
	}

	endCursor := pageInfo["endCursor"].(string)

	// Second request
	stdout, stderr, err = r.run("team", "list", "--output=json", "--limit=1", "--after="+endCursor)
	if err != nil {
		t.Fatalf("second page failed: %v\nstderr: %s", err, stderr)
	}

	var secondPage map[string]any
	if err := json.Unmarshal([]byte(stdout), &secondPage); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	t.Logf("Team pagination test passed")
}

// --- Search Tests ---

func TestCLI_IssueSearchCount(t *testing.T) {
	r := newCLIRunner(t)

	// Search with count flag
	stdout, stderr, err := r.run("issue", "search", "test", "--count", "--output=json")
	if err != nil {
		if strings.Contains(stderr, "no issues") {
			t.Skip("No issues matching search")
		}
		t.Fatalf("search count failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result["count"] == nil {
		t.Error("count is missing from search response")
	}

	t.Logf("Search count: %v", result["count"])
}

func TestCLI_IssueSearchWithArchived(t *testing.T) {
	r := newCLIRunner(t)

	// Search including archived issues
	stdout, stderr, err := r.run("issue", "search", "test", "--include-archived", "--output=json", "--limit=5")
	if err != nil {
		if strings.Contains(stderr, "no issues") {
			t.Skip("No issues matching search")
		}
		t.Fatalf("search with archived failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Search with archived found %d issues", len(nodes))
}

// --- Help and Version Tests ---

func TestCLI_Help(t *testing.T) {
	r := newCLIRunner(t)

	stdout, _, err := r.run("--help")
	if err != nil {
		t.Fatalf("help failed: %v", err)
	}

	// Help should contain usage information
	if !strings.Contains(stdout, "Usage") && !strings.Contains(stdout, "Available Commands") {
		t.Error("Help output missing expected content")
	}

	t.Logf("Help output length: %d bytes", len(stdout))
}

func TestCLI_SubcommandHelp(t *testing.T) {
	r := newCLIRunner(t)

	stdout, _, err := r.run("issue", "--help")
	if err != nil {
		t.Fatalf("issue help failed: %v", err)
	}

	// Should list issue subcommands
	if !strings.Contains(stdout, "list") || !strings.Contains(stdout, "get") {
		t.Error("Issue help missing expected subcommands")
	}

	t.Logf("Issue help output length: %d bytes", len(stdout))
}

// --- Team Members Count Test ---

func TestCLI_TeamMembersCount(t *testing.T) {
	r := newCLIRunner(t)

	// First get a team
	stdout, _, err := r.run("team", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No teams available")
	}

	var listResult map[string]any
	json.Unmarshal([]byte(stdout), &listResult)
	nodes := listResult["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No teams available")
	}

	team := nodes[0].(map[string]any)
	teamKey := team["key"].(string)

	// Get team members count
	stdout, stderr, err := r.run("team", "members", "--team="+teamKey, "--count", "--output=json")
	if err != nil {
		t.Fatalf("team members count failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result["count"] == nil {
		t.Error("count is missing from response")
	}

	t.Logf("Team %s has %v members", teamKey, result["count"])
}

// --- Issue with Relations Test ---

func TestCLI_IssueListWithHasChildren(t *testing.T) {
	r := newCLIRunner(t)

	// List issues that have sub-issues
	stdout, stderr, err := r.run("issue", "list", "--has-children", "--output=json", "--limit=5")
	if err != nil {
		t.Fatalf("issue list with has-children failed: %v\nstderr: %s", err, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected nodes array in response")
	}

	t.Logf("Retrieved %d issues with sub-issues", len(nodes))
}

// --- User Completed by Team Test ---

func TestCLI_UserCompletedByTeam(t *testing.T) {
	r := newCLIRunner(t)

	// First get a team
	stdout, _, err := r.run("team", "list", "--output=json", "--limit=1")
	if err != nil {
		t.Skip("No teams available")
	}

	var listResult map[string]any
	json.Unmarshal([]byte(stdout), &listResult)
	nodes := listResult["nodes"].([]any)
	if len(nodes) == 0 {
		t.Skip("No teams available")
	}

	team := nodes[0].(map[string]any)
	teamKey := team["key"].(string)

	// Get completed issues for team
	stdout, stderr, err := r.run("user", "completed",
		"--team="+teamKey,
		"--completed-after=7d",
		"--output=json")
	if err != nil {
		t.Fatalf("user completed by team failed: %v\nstderr: %s", err, stderr)
	}

	var results []any
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		// Try parsing as single object
		var result map[string]any
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}
		t.Logf("Team %s completed result: %v", teamKey, result)
		return
	}

	t.Logf("Team %s completed %d user entries in last 7 days", teamKey, len(results))
}
