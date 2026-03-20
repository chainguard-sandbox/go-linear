//go:build read

package initiative

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

func TestLive_InitiativeFilterBuilder(t *testing.T) {
	client, res, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("empty filter returns nil", func(t *testing.T) {
		builder := NewFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("name", "", "")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter != nil {
			t.Errorf("Build() with no flags = %v, want nil", filter)
		}
	})

	t.Run("date range filter", func(t *testing.T) {
		builder := NewFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("created-after", "", "")
		cmd.Flags().String("created-before", "", "")
		mustSet(cmd, "created-after", "90d")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil {
			t.Fatal("Build() = nil, want non-nil")
		}

		first := int64(50)
		initiatives, err := client.InitiativesFiltered(ctx, &first, nil, filter)
		if err != nil {
			t.Fatalf("InitiativesFiltered() error = %v", err)
		}

		t.Logf("Found %d initiatives created in last 90 days", len(initiatives.Nodes))
	})

	t.Run("status filter", func(t *testing.T) {
		builder := NewFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("status", "", "")
		mustSet(cmd, "status", "Active")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil {
			t.Fatal("Build() = nil, want non-nil")
		}

		first := int64(20)
		initiatives, err := client.InitiativesFiltered(ctx, &first, nil, filter)
		if err != nil {
			t.Fatalf("InitiativesFiltered() error = %v", err)
		}

		t.Logf("Found %d active initiatives", len(initiatives.Nodes))
	})

	t.Run("combined filters", func(t *testing.T) {
		builder := NewFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("created-after", "", "")
		cmd.Flags().String("status", "", "")
		mustSet(cmd, "created-after", "180d")
		mustSet(cmd, "status", "Active")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil {
			t.Fatal("Build() = nil, want non-nil")
		}

		first := int64(20)
		initiatives, err := client.InitiativesFiltered(ctx, &first, nil, filter)
		if err != nil {
			t.Fatalf("InitiativesFiltered() error = %v", err)
		}

		t.Logf("Found %d active initiatives created in last 180 days", len(initiatives.Nodes))
	})
}
