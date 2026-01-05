//go:build read

package issue

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// =============================================================================
// Test Setup/Teardown
// =============================================================================

// testSetup creates a Linear client and resolver for integration tests.
// Returns cleanup function that should be deferred.
func testSetup(t *testing.T) (*linear.Client, *resolver.Resolver, func()) {
	t.Helper()

	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("LINEAR_API_KEY not set")
	}

	client, err := linear.NewClient(apiKey)
	if err != nil {
		t.Fatal(err)
	}

	res := resolver.New(client)

	cleanup := func() {
		client.Close()
	}

	return client, res, cleanup
}

// =============================================================================
// Basic Filter Tests
// =============================================================================

func TestLive_IssueFilterBuilder(t *testing.T) {
	client, res, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("empty filter returns nil", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Int("priority", -1, "")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter != nil {
			t.Errorf("Build() with no flags = %v, want nil", filter)
		}
	})

	t.Run("priority filter", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Int("priority", -1, "")
		mustSet(cmd, "priority", "1")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil {
			t.Fatal("Build() = nil, want non-nil")
		}
		if filter.Priority == nil || filter.Priority.Eq == nil {
			t.Error("Priority filter not set")
		}
		if *filter.Priority.Eq != 1.0 {
			t.Errorf("Priority = %f, want 1.0", *filter.Priority.Eq)
		}
	})

	t.Run("date filter with ISO8601", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("created-after", "", "")
		mustSet(cmd, "created-after", "2025-12-01")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.CreatedAt == nil || filter.CreatedAt.Gte == nil {
			t.Error("CreatedAt filter not set")
		}
	})

	t.Run("date filter with relative date", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("completed-after", "", "")
		mustSet(cmd, "completed-after", "7d")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.CompletedAt == nil || filter.CompletedAt.Gte == nil {
			t.Error("CompletedAt filter not set")
		}
	})

	t.Run("invalid date returns error", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("created-after", "", "")
		mustSet(cmd, "created-after", "invalid-date")

		err := builder.FromFlags(ctx, cmd)
		if err == nil {
			t.Error("FromFlags() with invalid date should return error")
		}
	})

	// =========================================================================
	// Entity Resolution Tests (require live API)
	// =========================================================================

	t.Run("team resolution with live API", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("team", "", "")

		// Get first team from API
		teams, err := client.Teams(ctx, nil, nil)
		if err != nil || len(teams.Nodes) == 0 {
			t.Skip("No teams available for testing")
		}

		teamName := teams.Nodes[0].Name
		mustSet(cmd, "team", teamName)

		err = builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.Team == nil {
			t.Error("Team filter not set")
		}
	})

	t.Run("team resolution by key", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("team", "", "")

		// Get first team from API
		teams, err := client.Teams(ctx, nil, nil)
		if err != nil || len(teams.Nodes) == 0 {
			t.Skip("No teams available for testing")
		}

		teamKey := teams.Nodes[0].Key
		mustSet(cmd, "team", teamKey)

		err = builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.Team == nil {
			t.Error("Team filter not set when using team key")
		}
	})

	t.Run("assignee me resolution", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("assignee", "", "")
		mustSet(cmd, "assignee", "me")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.Assignee == nil {
			t.Error("Assignee filter not set")
		}
	})

	t.Run("state resolution with live API", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("state", "", "")

		// Get first state from API
		states, err := client.WorkflowStates(ctx, nil, nil)
		if err != nil || len(states.Nodes) == 0 {
			t.Skip("No workflow states available for testing")
		}

		stateName := states.Nodes[0].Name
		mustSet(cmd, "state", stateName)

		err = builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.State == nil {
			t.Error("State filter not set")
		}
	})

	t.Run("creator me resolution", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("creator", "", "")
		mustSet(cmd, "creator", "me")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.Creator == nil {
			t.Error("Creator filter not set")
		}
	})

	t.Run("label resolution with live API", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().StringArray("label", nil, "")

		// Get first label from API
		labels, err := client.IssueLabels(ctx, nil, nil)
		if err != nil || len(labels.Nodes) == 0 {
			t.Skip("No labels available for testing")
		}

		labelName := labels.Nodes[0].Name
		mustSet(cmd, "label", labelName)

		err = builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.Labels == nil {
			t.Error("Labels filter not set")
		}
	})

	t.Run("multiple labels resolution", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().StringArray("label", nil, "")

		// Get labels from API
		labels, err := client.IssueLabels(ctx, nil, nil)
		if err != nil || len(labels.Nodes) < 2 {
			t.Skip("Need at least 2 labels for testing")
		}

		mustSet(cmd, "label", labels.Nodes[0].Name)
		mustSet(cmd, "label", labels.Nodes[1].Name)

		err = builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.Labels == nil {
			t.Error("Labels filter not set for multiple labels")
		}
	})

	// =========================================================================
	// Error Handling Tests
	// =========================================================================

	t.Run("invalid team returns error", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("team", "", "")
		mustSet(cmd, "team", "nonexistent-team-12345")

		err := builder.FromFlags(ctx, cmd)
		if err == nil {
			t.Error("Expected error for invalid team name")
		}
	})

	t.Run("invalid state returns error", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("state", "", "")
		mustSet(cmd, "state", "nonexistent-state-12345")

		err := builder.FromFlags(ctx, cmd)
		if err == nil {
			t.Error("Expected error for invalid state name")
		}
	})

	t.Run("invalid label returns error", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().StringArray("label", nil, "")
		mustSet(cmd, "label", "nonexistent-label-12345")

		err := builder.FromFlags(ctx, cmd)
		if err == nil {
			t.Error("Expected error for invalid label name")
		}
	})

	// =========================================================================
	// Combined Filter Tests
	// =========================================================================

	t.Run("multiple filters combined", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Int("priority", -1, "")
		cmd.Flags().String("created-after", "", "")
		cmd.Flags().Bool("has-blocked-by", false, "")

		mustSet(cmd, "priority", "2")
		mustSet(cmd, "created-after", "2025-01-01")
		mustSet(cmd, "has-blocked-by", "true")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil {
			t.Fatal("Build() = nil, want non-nil")
		}
		if filter.Priority == nil {
			t.Error("Priority filter not set in combined filter")
		}
		if filter.CreatedAt == nil {
			t.Error("CreatedAt filter not set in combined filter")
		}
		if filter.HasBlockedByRelations == nil {
			t.Error("HasBlockedByRelations filter not set in combined filter")
		}
	})

	t.Run("team and state combined", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("team", "", "")
		cmd.Flags().String("state", "", "")

		// Get team and state from API
		teams, err := client.Teams(ctx, nil, nil)
		if err != nil || len(teams.Nodes) == 0 {
			t.Skip("No teams available")
		}

		states, err := client.WorkflowStates(ctx, nil, nil)
		if err != nil || len(states.Nodes) == 0 {
			t.Skip("No states available")
		}

		mustSet(cmd, "team", teams.Nodes[0].Name)
		mustSet(cmd, "state", states.Nodes[0].Name)

		err = builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil {
			t.Fatal("Build() = nil")
		}
		if filter.Team == nil {
			t.Error("Team filter not set")
		}
		if filter.State == nil {
			t.Error("State filter not set")
		}
	})
}

// =============================================================================
// Cycle Filter Tests
// =============================================================================

func TestLive_CycleFilter(t *testing.T) {
	client, res, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("cycle resolution with live API", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("cycle", "", "")

		// Get first cycle from API
		cycles, err := client.Cycles(ctx, nil, nil)
		if err != nil || len(cycles.Nodes) == 0 {
			t.Skip("No cycles available for testing")
		}

		cycleID := cycles.Nodes[0].ID
		mustSet(cmd, "cycle", cycleID)

		err = builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.Cycle == nil {
			t.Error("Cycle filter not set")
		}
	})
}

// =============================================================================
// Project Filter Tests
// =============================================================================

func TestLive_ProjectFilter(t *testing.T) {
	client, res, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("project resolution with live API", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("project", "", "")

		// Get first project from API
		projects, err := client.Projects(ctx, nil, nil)
		if err != nil || len(projects.Nodes) == 0 {
			t.Skip("No projects available for testing")
		}

		projectName := projects.Nodes[0].Name
		mustSet(cmd, "project", projectName)

		err = builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.Project == nil {
			t.Error("Project filter not set")
		}
	})
}

// =============================================================================
// Subscriber Filter Tests
// =============================================================================

func TestLive_SubscriberFilter(t *testing.T) {
	_, res, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("subscriber me resolution", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().StringArray("subscriber", nil, "")
		mustSet(cmd, "subscriber", "me")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.Subscribers == nil {
			t.Error("Subscribers filter not set")
		}
	})
}

// =============================================================================
// Comment Filter Tests
// =============================================================================

func TestLive_CommentFilter(t *testing.T) {
	_, res, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("comment-by me resolution", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().StringArray("comment-by", nil, "")
		mustSet(cmd, "comment-by", "me")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.Comments == nil {
			t.Error("Comments filter not set")
		}
	})
}

// =============================================================================
// Attachment Filter Tests
// =============================================================================

func TestLive_AttachmentFilter(t *testing.T) {
	_, res, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("attachment-by me resolution", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().StringArray("attachment-by", nil, "")
		mustSet(cmd, "attachment-by", "me")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.Attachments == nil {
			t.Error("Attachments filter not set")
		}
	})
}

// =============================================================================
// Snoozed Filter Tests
// =============================================================================

func TestLive_SnoozedFilter(t *testing.T) {
	_, res, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("snoozed-by me resolution", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("snoozed-by", "", "")
		mustSet(cmd, "snoozed-by", "me")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.SnoozedBy == nil {
			t.Error("SnoozedBy filter not set")
		}
	})
}

// =============================================================================
// Delegate Filter Tests
// =============================================================================

func TestLive_DelegateFilter(t *testing.T) {
	_, res, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("delegate me resolution", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("delegate", "", "")
		mustSet(cmd, "delegate", "me")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.Delegate == nil {
			t.Error("Delegate filter not set")
		}
	})
}
