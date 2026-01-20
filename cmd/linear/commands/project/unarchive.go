package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUnarchiveCommand creates the project unarchive command.
func NewUnarchiveCommand(clientFactory cli.ClientFactory) *cobra.Command {
	outputFlags := &cli.OutputOnlyFlags{}

	cmd := &cobra.Command{
		Use:   "unarchive <name-or-id>",
		Short: "Unarchive a project",
		Long: `Restore an archived project. Safe operation.

Example: go-linear project unarchive <uuid>
Example: go-linear project unarchive "Q1 Platform"

Related: project_delete, project_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runUnarchive(cmd, client, args[0], outputFlags)
		},
	}

	outputFlags.Bind(cmd)

	return cmd
}

func runUnarchive(cmd *cobra.Command, client *linear.Client, projectID string, outputFlags *cli.OutputOnlyFlags) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	if err := outputFlags.Validate(); err != nil {
		return err
	}

	// Resolve project ID
	resolvedID, err := res.ResolveProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to resolve project: %w", err)
	}

	err = client.ProjectUnarchive(ctx, resolvedID)
	if err != nil {
		return fmt.Errorf("failed to unarchive project: %w", err)
	}

	if outputFlags.Output == "json" {
		return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
			"success":   true,
			"projectId": projectID,
		}, true)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Project %s unarchived successfully\n", projectID)
	return nil
}
