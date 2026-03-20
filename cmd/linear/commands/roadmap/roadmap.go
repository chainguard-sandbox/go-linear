// Package roadmap provides roadmap commands for the Linear CLI.
package roadmap

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
)

// cli.ClientFactory is a function that creates a Linear client.

// NewRoadmapCommand creates the roadmap command group.
func NewRoadmapCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:        "roadmap",
		Short:      "Manage Linear roadmaps",
		Long:       "Commands for listing and viewing Linear roadmaps.",
		Deprecated: "Roadmaps are deprecated by Linear (GraphQL schema). Use 'initiative' commands instead.",
	}

	cmd.AddCommand(NewListCommand(clientFactory))
	cmd.AddCommand(NewGetCommand(clientFactory))

	return cmd
}
