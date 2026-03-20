package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewArchiveCommand creates the project archive command.
func NewArchiveCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive <name-or-id>",
		Short: "Archive a project",
		Long: `Archive project. Modifies data. Use unarchive to restore.

Note: Linear recommends delete instead of archive.

Example: go-linear project archive "Q1 Platform"

Related: project_unarchive, project_delete, project_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runArchive(cmd, client, args[0])
		},
	}

	return cmd
}

func runArchive(cmd *cobra.Command, client *linear.Client, projectID string) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	// Resolve project ID
	resolvedID, err := res.ResolveProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to resolve project: %w", err)
	}

	err = client.ProjectArchive(ctx, resolvedID)
	if err != nil {
		return fmt.Errorf("failed to archive project: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
		"success":   true,
		"projectId": projectID,
	}, true)
}
