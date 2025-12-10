// Package commands provides the Cobra command structure for the Linear CLI.
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/cmd/linear/commands/attachment"
	"github.com/chainguard-sandbox/go-linear/cmd/linear/commands/comment"
	"github.com/chainguard-sandbox/go-linear/cmd/linear/commands/cycle"
	"github.com/chainguard-sandbox/go-linear/cmd/linear/commands/document"
	"github.com/chainguard-sandbox/go-linear/cmd/linear/commands/favorite"
	"github.com/chainguard-sandbox/go-linear/cmd/linear/commands/initiative"
	"github.com/chainguard-sandbox/go-linear/cmd/linear/commands/issue"
	"github.com/chainguard-sandbox/go-linear/cmd/linear/commands/label"
	"github.com/chainguard-sandbox/go-linear/cmd/linear/commands/organization"
	"github.com/chainguard-sandbox/go-linear/cmd/linear/commands/project"
	"github.com/chainguard-sandbox/go-linear/cmd/linear/commands/reaction"
	"github.com/chainguard-sandbox/go-linear/cmd/linear/commands/roadmap"
	"github.com/chainguard-sandbox/go-linear/cmd/linear/commands/state"
	"github.com/chainguard-sandbox/go-linear/cmd/linear/commands/team"
	"github.com/chainguard-sandbox/go-linear/cmd/linear/commands/template"
	"github.com/chainguard-sandbox/go-linear/cmd/linear/commands/user"
	"github.com/chainguard-sandbox/go-linear/cmd/linear/commands/viewer"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewRootCommand creates the root command for the Linear CLI.
func NewRootCommand() *cobra.Command {
	var apiKey string

	rootCmd := &cobra.Command{
		Use:   "linear",
		Short: "Linear CLI for humans and AI agents",
		Long: `Linear CLI provides command-line access to Linear's issue tracking system.

Optimized for both human users and AI agents via MCP (Model Context Protocol).
Supports parameter-rich commands for complex queries without multi-step workflows.

Examples:
  # List my urgent issues
  linear issue list --assignee=me --priority=1

  # Find completed issues from yesterday
  linear issue list --team=Engineering --completed-after=yesterday --completed-before=today

  # Get user's completed work
  linear user completed --user=alice@company.com --completed-after=7d

Environment Variables:
  LINEAR_API_KEY    Linear API key (required if not provided via flag)`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize Linear client for all subcommands
			if apiKey == "" {
				apiKey = os.Getenv("LINEAR_API_KEY")
			}
			if apiKey == "" {
				return fmt.Errorf("LINEAR_API_KEY environment variable or --api-key flag required")
			}

			// Validation only - actual client creation happens in subcommands
			return nil
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Linear API key (or set LINEAR_API_KEY env var)")

	// Get API key for subcommands
	getClient := func() (*linear.Client, error) {
		key := apiKey
		if key == "" {
			key = os.Getenv("LINEAR_API_KEY")
		}
		if key == "" {
			return nil, fmt.Errorf("LINEAR_API_KEY environment variable or --api-key flag required")
		}
		return linear.NewClient(key)
	}

	// Add subcommands (ordered alphabetically)
	rootCmd.AddCommand(attachment.NewAttachmentCommand(getClient))
	rootCmd.AddCommand(comment.NewCommentCommand(getClient))
	rootCmd.AddCommand(cycle.NewCycleCommand(getClient))
	rootCmd.AddCommand(document.NewDocumentCommand(getClient))
	rootCmd.AddCommand(favorite.NewFavoriteCommand(getClient))
	rootCmd.AddCommand(initiative.NewInitiativeCommand(getClient))
	rootCmd.AddCommand(issue.NewIssueCommand(getClient))
	rootCmd.AddCommand(label.NewLabelCommand(getClient))
	rootCmd.AddCommand(organization.NewOrganizationCommand(getClient))
	rootCmd.AddCommand(project.NewProjectCommand(getClient))
	rootCmd.AddCommand(reaction.NewReactionCommand(getClient))
	rootCmd.AddCommand(roadmap.NewRoadmapCommand(getClient))
	rootCmd.AddCommand(state.NewStateCommand(getClient))
	rootCmd.AddCommand(team.NewTeamCommand(getClient))
	rootCmd.AddCommand(template.NewTemplateCommand(getClient))
	rootCmd.AddCommand(user.NewUserCommand(getClient))
	rootCmd.AddCommand(viewer.NewViewerCommand(getClient))

	return rootCmd
}
