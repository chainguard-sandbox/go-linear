//go:build read

package linear

import (
	"context"
	"os"
	"testing"
)

// TestLive_Viewer tests the Viewer query against real Linear API.
// Safe read-only operation.
func TestLive_Viewer(t *testing.T) {
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping live test: LINEAR_API_KEY not set")
	}

	client, err := NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()
	viewer, err := client.Viewer(ctx)
	if err != nil {
		t.Fatalf("Viewer() error = %v", err)
	}

	if viewer == nil {
		t.Fatal("Viewer() returned nil")
	}

	if viewer.ID == "" {
		t.Error("Viewer().ID is empty")
	}

	if viewer.Email == "" {
		t.Error("Viewer().Email is empty")
	}

	t.Logf("Authenticated as: %s (%s)", viewer.Name, viewer.Email)
}
