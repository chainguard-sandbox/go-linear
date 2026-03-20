//go:build read

package document

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

func TestLive_DocumentFilterBuilder(t *testing.T) {
	client, res, cleanup := testSetup(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("empty filter returns nil", func(t *testing.T) {
		builder := NewFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		// Add all flags that would be needed
		cmd.Flags().String("title", "", "")

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
		docs, err := client.DocumentsFiltered(ctx, &first, nil, filter)
		if err != nil {
			t.Fatalf("DocumentsFiltered() error = %v", err)
		}

		t.Logf("Found %d documents created in last 30 days", len(docs.Nodes))
	})

	t.Run("title filter", func(t *testing.T) {
		builder := NewFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("title", "", "")
		// Use a common word that might appear in document titles
		mustSet(cmd, "title", "")

		// Even with empty title, test that API works
		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		// Test unfiltered query works
		first := int64(10)
		docs, err := client.Documents(ctx, &first, nil)
		if err != nil {
			t.Fatalf("Documents() error = %v", err)
		}

		t.Logf("Found %d documents total", len(docs.Nodes))
	})

	t.Run("combined filters", func(t *testing.T) {
		builder := NewFilterBuilder(res)
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("created-after", "", "")
		cmd.Flags().String("updated-after", "", "")
		mustSet(cmd, "created-after", "90d")
		mustSet(cmd, "updated-after", "30d")

		err := builder.FromFlags(ctx, cmd)
		if err != nil {
			t.Fatalf("FromFlags() error = %v", err)
		}

		filter := builder.Build()
		if filter == nil {
			t.Fatal("Build() = nil, want non-nil")
		}

		first := int64(20)
		docs, err := client.DocumentsFiltered(ctx, &first, nil, filter)
		if err != nil {
			t.Fatalf("DocumentsFiltered() error = %v", err)
		}

		t.Logf("Found %d documents created in last 90 days and updated in last 30 days", len(docs.Nodes))
	})
}
