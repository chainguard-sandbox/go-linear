//go:build read

package filter

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

func TestLive_IssueFilterBuilder(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("LINEAR_API_KEY not set")
	}

	client, err := linear.NewClient(apiKey)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	res := resolver.New(client)
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
		cmd.Flags().Set("priority", "1")

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
		cmd.Flags().Set("created-after", "2025-12-01")

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
		cmd.Flags().Set("completed-after", "7d")

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
		cmd.Flags().Set("created-after", "invalid-date")

		err := builder.FromFlags(ctx, cmd)
		if err == nil {
			t.Error("FromFlags() with invalid date should return error")
		}
	})

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
		cmd.Flags().Set("team", teamName)

		err = builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.Team == nil {
			t.Error("Team filter not set")
		}
	})

	t.Run("assignee me resolution", func(t *testing.T) {
		builder := NewIssueFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("assignee", "", "")
		cmd.Flags().Set("assignee", "me")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil || filter.Assignee == nil {
			t.Error("Assignee filter not set")
		}
	})
}
