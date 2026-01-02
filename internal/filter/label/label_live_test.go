//go:build read

package label

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

func TestLive_LabelFilterBuilder(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("LINEAR_API_KEY not set")
	}

	client, err := linear.NewClient(apiKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	res := resolver.New(client)

	t.Run("empty filter returns nil", func(t *testing.T) {
		b := NewFilterBuilder(res)
		if b.Build() != nil {
			t.Error("expected nil filter for empty builder")
		}
	})

	t.Run("date range filter", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("created-after", "90d", "")
		cmd.Flags().String("created-before", "", "")
		cmd.Flags().String("updated-after", "", "")
		cmd.Flags().String("updated-before", "", "")
		cmd.Flags().String("id", "", "")
		cmd.Flags().String("creator", "", "")
		cmd.Flags().String("team", "", "")
		cmd.Flags().String("name", "", "")
		cmd.Flags().Bool("is-group", false, "")

		b := NewFilterBuilder(res)
		if err := b.FromFlags(context.Background(), cmd); err != nil {
			t.Fatalf("FromFlags() error: %v", err)
		}

		filter := b.Build()
		if filter == nil {
			t.Fatal("expected filter, got nil")
		}

		// Query with filter
		first := int64(10)
		labels, err := client.IssueLabelsFiltered(context.Background(), &first, nil, filter)
		if err != nil {
			t.Fatalf("IssueLabelsFiltered() error: %v", err)
		}
		t.Logf("Found %d labels created in last 90 days", len(labels.Nodes))
	})

	t.Run("name filter", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("created-after", "", "")
		cmd.Flags().String("created-before", "", "")
		cmd.Flags().String("updated-after", "", "")
		cmd.Flags().String("updated-before", "", "")
		cmd.Flags().String("id", "", "")
		cmd.Flags().String("creator", "", "")
		cmd.Flags().String("team", "", "")
		cmd.Flags().String("name", "a", "") // Broad search
		cmd.Flags().Bool("is-group", false, "")

		b := NewFilterBuilder(res)
		if err := b.FromFlags(context.Background(), cmd); err != nil {
			t.Fatalf("FromFlags() error: %v", err)
		}

		filter := b.Build()
		if filter == nil {
			t.Fatal("expected filter, got nil")
		}

		first := int64(10)
		labels, err := client.IssueLabelsFiltered(context.Background(), &first, nil, filter)
		if err != nil {
			t.Fatalf("IssueLabelsFiltered() error: %v", err)
		}
		t.Logf("Found %d labels with 'a' in name", len(labels.Nodes))
	})

	t.Run("combined filters", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("created-after", "180d", "")
		cmd.Flags().String("created-before", "", "")
		cmd.Flags().String("updated-after", "", "")
		cmd.Flags().String("updated-before", "", "")
		cmd.Flags().String("id", "", "")
		cmd.Flags().String("creator", "", "")
		cmd.Flags().String("team", "", "")
		cmd.Flags().String("name", "", "")
		cmd.Flags().Bool("is-group", false, "")

		b := NewFilterBuilder(res)
		if err := b.FromFlags(context.Background(), cmd); err != nil {
			t.Fatalf("FromFlags() error: %v", err)
		}

		filter := b.Build()
		if filter == nil {
			t.Fatal("expected filter, got nil")
		}

		first := int64(20)
		labels, err := client.IssueLabelsFiltered(context.Background(), &first, nil, filter)
		if err != nil {
			t.Fatalf("IssueLabelsFiltered() error: %v", err)
		}
		t.Logf("Found %d labels from last 180 days", len(labels.Nodes))
	})
}
