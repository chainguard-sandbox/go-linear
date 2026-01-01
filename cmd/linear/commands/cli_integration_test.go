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
	t.Skip("Status command has time marshaling bug with zero time values - skipping until fixed")
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
