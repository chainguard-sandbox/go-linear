//go:build read

package attachment

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

func TestLive_AttachmentFilterBuilder(t *testing.T) {
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
		cmd.Flags().String("source-type", "", "")
		cmd.Flags().String("title", "", "")
		cmd.Flags().String("subtitle", "", "")
		cmd.Flags().String("url", "", "")

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
		attachments, err := client.AttachmentsFiltered(context.Background(), &first, nil, filter)
		if err != nil {
			t.Fatalf("AttachmentsFiltered() error: %v", err)
		}
		t.Logf("Found %d attachments created in last 90 days", len(attachments.Nodes))
	})

	t.Run("title filter", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("created-after", "", "")
		cmd.Flags().String("created-before", "", "")
		cmd.Flags().String("updated-after", "", "")
		cmd.Flags().String("updated-before", "", "")
		cmd.Flags().String("id", "", "")
		cmd.Flags().String("creator", "", "")
		cmd.Flags().String("source-type", "", "")
		cmd.Flags().String("title", "a", "") // Broad search
		cmd.Flags().String("subtitle", "", "")
		cmd.Flags().String("url", "", "")

		b := NewFilterBuilder(res)
		if err := b.FromFlags(context.Background(), cmd); err != nil {
			t.Fatalf("FromFlags() error: %v", err)
		}

		filter := b.Build()
		if filter == nil {
			t.Fatal("expected filter, got nil")
		}

		first := int64(10)
		attachments, err := client.AttachmentsFiltered(context.Background(), &first, nil, filter)
		if err != nil {
			t.Fatalf("AttachmentsFiltered() error: %v", err)
		}
		t.Logf("Found %d attachments with 'a' in title", len(attachments.Nodes))
	})

	t.Run("combined filters", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("created-after", "180d", "")
		cmd.Flags().String("created-before", "", "")
		cmd.Flags().String("updated-after", "", "")
		cmd.Flags().String("updated-before", "", "")
		cmd.Flags().String("id", "", "")
		cmd.Flags().String("creator", "", "")
		cmd.Flags().String("source-type", "", "")
		cmd.Flags().String("title", "", "")
		cmd.Flags().String("subtitle", "", "")
		cmd.Flags().String("url", "", "")

		b := NewFilterBuilder(res)
		if err := b.FromFlags(context.Background(), cmd); err != nil {
			t.Fatalf("FromFlags() error: %v", err)
		}

		filter := b.Build()
		if filter == nil {
			t.Fatal("expected filter, got nil")
		}

		first := int64(20)
		attachments, err := client.AttachmentsFiltered(context.Background(), &first, nil, filter)
		if err != nil {
			t.Fatalf("AttachmentsFiltered() error: %v", err)
		}
		t.Logf("Found %d attachments from last 180 days", len(attachments.Nodes))
	})
}
