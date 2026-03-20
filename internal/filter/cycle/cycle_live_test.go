//go:build read

package cycle

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// testSetup creates a Linear client and resolver for integration tests.
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
// Integration Tests
// =============================================================================

func TestLive_CycleFilterBuilder(t *testing.T) {
	client, res, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("empty filter returns nil", func(t *testing.T) {
		builder := NewFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		// Add all flags that would be needed
		cmd.Flags().Int("number", -1, "")
		cmd.Flags().String("team", "", "")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter != nil {
			t.Errorf("Build() with no flags = %v, want nil", filter)
		}
	})

	t.Run("active cycle filter", func(t *testing.T) {
		builder := NewFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Bool("active", false, "")
		mustSet(cmd, "active", "true")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil {
			t.Fatal("Build() = nil, want non-nil")
		}
		if filter.IsActive == nil {
			t.Error("IsActive filter not set")
		}

		// Test against API
		first := int64(10)
		cycles, err := client.CyclesFiltered(ctx, &first, nil, filter)
		if err != nil {
			t.Fatalf("CyclesFiltered() error = %v", err)
		}

		// All returned cycles should be active (if any)
		t.Logf("Found %d active cycles", len(cycles.Nodes))
	})

	t.Run("past cycles filter", func(t *testing.T) {
		builder := NewFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Bool("past", false, "")
		mustSet(cmd, "past", "true")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil {
			t.Fatal("Build() = nil, want non-nil")
		}

		first := int64(10)
		cycles, err := client.CyclesFiltered(ctx, &first, nil, filter)
		if err != nil {
			t.Fatalf("CyclesFiltered() error = %v", err)
		}

		t.Logf("Found %d past cycles", len(cycles.Nodes))
	})

	t.Run("date range filter", func(t *testing.T) {
		builder := NewFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("created-after", "", "")
		cmd.Flags().String("created-before", "", "")
		mustSet(cmd, "created-after", "30d")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil {
			t.Fatal("Build() = nil, want non-nil")
		}

		first := int64(50)
		cycles, err := client.CyclesFiltered(ctx, &first, nil, filter)
		if err != nil {
			t.Fatalf("CyclesFiltered() error = %v", err)
		}

		t.Logf("Found %d cycles created in last 30 days", len(cycles.Nodes))
	})

	t.Run("team filter", func(t *testing.T) {
		// First get a team to filter by
		teams, err := client.Teams(ctx, nil, nil)
		if err != nil {
			t.Fatalf("Teams() error = %v", err)
		}
		if len(teams.Nodes) == 0 {
			t.Skip("No teams available")
		}

		teamName := teams.Nodes[0].Name

		builder := NewFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("team", "", "")
		mustSet(cmd, "team", teamName)

		err = builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil {
			t.Fatal("Build() = nil, want non-nil")
		}

		first := int64(10)
		cycles, err := client.CyclesFiltered(ctx, &first, nil, filter)
		if err != nil {
			t.Fatalf("CyclesFiltered() error = %v", err)
		}

		t.Logf("Found %d cycles for team %s", len(cycles.Nodes), teamName)

		// Verify all returned cycles are from the correct team
		for _, cycle := range cycles.Nodes {
			if cycle.Team.Name != teamName {
				t.Errorf("Cycle %s team = %s, want %s", cycle.ID, cycle.Team.Name, teamName)
			}
		}
	})

	t.Run("number filter", func(t *testing.T) {
		// First get a cycle to filter by number
		first := int64(1)
		cycles, err := client.Cycles(ctx, &first, nil)
		if err != nil {
			t.Fatalf("Cycles() error = %v", err)
		}
		if len(cycles.Nodes) == 0 {
			t.Skip("No cycles available")
		}

		cycleNumber := int(cycles.Nodes[0].Number)

		builder := NewFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Int("number", -1, "")
		mustSet(cmd, "number", itoa(cycleNumber))

		err = builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil {
			t.Fatal("Build() = nil, want non-nil")
		}

		filteredCycles, err := client.CyclesFiltered(ctx, &first, nil, filter)
		if err != nil {
			t.Fatalf("CyclesFiltered() error = %v", err)
		}

		t.Logf("Found %d cycles with number %d", len(filteredCycles.Nodes), cycleNumber)
	})

	t.Run("combined filters", func(t *testing.T) {
		builder := NewFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().Bool("past", false, "")
		cmd.Flags().String("created-after", "", "")
		mustSet(cmd, "past", "true")
		mustSet(cmd, "created-after", "90d")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil {
			t.Fatal("Build() = nil, want non-nil")
		}

		first := int64(20)
		cycles, err := client.CyclesFiltered(ctx, &first, nil, filter)
		if err != nil {
			t.Fatalf("CyclesFiltered() error = %v", err)
		}

		t.Logf("Found %d past cycles created in last 90 days", len(cycles.Nodes))
	})
}
