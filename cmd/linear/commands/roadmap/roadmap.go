// Package roadmap provides roadmap commands for the Linear CLI.
package roadmap

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// ClientFactory is a function that creates a Linear client.
type ClientFactory func() (*linear.Client, error)

// NewRoadmapCommand creates the roadmap command group.
func NewRoadmapCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "roadmap",
		Short: "Manage Linear roadmaps",
		Long:  "Commands for listing and viewing Linear roadmaps.",
	}

	cmd.AddCommand(NewListCommand(clientFactory))
	cmd.AddCommand(NewGetCommand(clientFactory))

	return cmd
}
