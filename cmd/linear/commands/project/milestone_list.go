package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewMilestoneListCommand creates the project milestone-list command.
func NewMilestoneListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "milestone-list <project>",
		Short: "List milestones within a project",
		Long: `List project milestones. Returns milestone id, name, targetDate.

Required: project name or ID

Example: go-linear project milestone-list "Q1 Platform"

Related: project_milestone-create, project_milestone-update, project_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runMilestoneList(cmd, client, args[0])
		},
	}

	return cmd
}

func runMilestoneList(cmd *cobra.Command, client *linear.Client, projectArg string) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	projectID, err := res.ResolveProject(ctx, projectArg)
	if err != nil {
		return fmt.Errorf("failed to resolve project: %w", err)
	}

	project, err := client.Project(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	milestones := project.ProjectMilestones.Nodes

	return formatter.FormatJSON(cmd.OutOrStdout(), milestones, true)
}
