package project

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUpdateCommand creates the project update command.
func NewUpdateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing project",
		Long: `Update project. Modifies existing data.

Fields: --name, --description, --content, --lead (user), --member (user, repeatable), --team (repeatable), --state, --start-date, --target-date (YYYY-MM-DD), --priority (0-4), --color, --icon, --label (repeatable), --slack-new-issue, --slack-issue-comments, --slack-issue-statuses

Example: go-linear project update <id> --state="In Progress" --target-date=2024-06-30 --priority=2

Related: project_get, project_list, project_create`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runUpdate(cmd, client, args[0])
		},
	}

	// Core fields
	cmd.Flags().String("name", "", "New project name")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().String("content", "", "Project content (markdown body)")
	cmd.Flags().String("lead", "", "Project lead (user name, email, or ID)")
	cmd.Flags().StringArray("member", []string{}, "Set project members (replaces all)")
	cmd.Flags().StringArray("team", []string{}, "Set associated teams (replaces all)")

	// Status & Dates
	cmd.Flags().String("state", "", "Project status (e.g., Backlog, In Progress, Completed)")
	cmd.Flags().String("start-date", "", "Planned start date (YYYY-MM-DD)")
	cmd.Flags().String("target-date", "", "Target completion date (YYYY-MM-DD)")
	cmd.Flags().Int("priority", -1, "Priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low")

	// Appearance
	cmd.Flags().String("color", "", "Project color (hex code, e.g., #ff0000)")
	cmd.Flags().String("icon", "", "Project icon")
	cmd.Flags().StringArray("label", []string{}, "Set project labels (replaces all)")

	// Slack integration
	cmd.Flags().Bool("slack-new-issue", false, "Send new issue notifications to Slack")
	cmd.Flags().Bool("slack-issue-comments", false, "Send comment notifications to Slack")
	cmd.Flags().Bool("slack-issue-statuses", false, "Send status update notifications to Slack")

	return cmd
}

func runUpdate(cmd *cobra.Command, client *linear.Client, projectID string) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	input := intgraphql.ProjectUpdateInput{}
	updated := false

	// Core fields
	if name, _ := cmd.Flags().GetString("name"); name != "" {
		input.Name = &name
		updated = true
	}

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		input.Description = &desc
		updated = true
	}

	if content, _ := cmd.Flags().GetString("content"); content != "" {
		input.Content = &content
		updated = true
	}

	if lead, _ := cmd.Flags().GetString("lead"); lead != "" {
		leadID, err := res.ResolveUser(ctx, lead)
		if err != nil {
			return fmt.Errorf("failed to resolve lead: %w", err)
		}
		input.LeadID = &leadID
		updated = true
	}

	members, _ := cmd.Flags().GetStringArray("member")
	if len(members) > 0 {
		memberIDs := make([]string, 0, len(members))
		for _, member := range members {
			memberID, err := res.ResolveUser(ctx, member)
			if err != nil {
				return fmt.Errorf("failed to resolve member %q: %w", member, err)
			}
			memberIDs = append(memberIDs, memberID)
		}
		input.MemberIds = memberIDs
		updated = true
	}

	teams, _ := cmd.Flags().GetStringArray("team")
	if len(teams) > 0 {
		teamIDs := make([]string, 0, len(teams))
		for _, team := range teams {
			teamID, err := res.ResolveTeam(ctx, team)
			if err != nil {
				return fmt.Errorf("failed to resolve team %q: %w", team, err)
			}
			teamIDs = append(teamIDs, teamID)
		}
		input.TeamIds = teamIDs
		updated = true
	}

	// Status & Dates
	if state, _ := cmd.Flags().GetString("state"); state != "" {
		statusID, err := res.ResolveProjectStatus(ctx, state)
		if err != nil {
			return fmt.Errorf("failed to resolve project status: %w", err)
		}
		input.StatusID = &statusID
		updated = true
	}

	if startDate, _ := cmd.Flags().GetString("start-date"); startDate != "" {
		// Validate date format
		if _, err := time.Parse("2006-01-02", startDate); err != nil {
			return fmt.Errorf("invalid start-date format (use YYYY-MM-DD): %w", err)
		}
		input.StartDate = &startDate
		updated = true
	}

	if targetDate, _ := cmd.Flags().GetString("target-date"); targetDate != "" {
		// Validate date format
		if _, err := time.Parse("2006-01-02", targetDate); err != nil {
			return fmt.Errorf("invalid target-date format (use YYYY-MM-DD): %w", err)
		}
		input.TargetDate = &targetDate
		updated = true
	}

	if priority, _ := cmd.Flags().GetInt("priority"); priority >= 0 {
		if priority > 4 {
			return fmt.Errorf("invalid priority %d: must be 0-4 (0=none, 1=urgent, 2=high, 3=normal, 4=low)", priority)
		}
		p := int64(priority)
		input.Priority = &p
		updated = true
	}

	// Appearance
	if color, _ := cmd.Flags().GetString("color"); color != "" {
		input.Color = &color
		updated = true
	}

	if icon, _ := cmd.Flags().GetString("icon"); icon != "" {
		input.Icon = &icon
		updated = true
	}

	labels, _ := cmd.Flags().GetStringArray("label")
	if len(labels) > 0 {
		labelIDs := make([]string, 0, len(labels))
		for _, label := range labels {
			labelID, err := res.ResolveLabel(ctx, label)
			if err != nil {
				return fmt.Errorf("failed to resolve label %q: %w", label, err)
			}
			labelIDs = append(labelIDs, labelID)
		}
		input.LabelIds = labelIDs
		updated = true
	}

	// Slack integration
	if cmd.Flags().Changed("slack-new-issue") {
		v, _ := cmd.Flags().GetBool("slack-new-issue")
		input.SlackNewIssue = &v
		updated = true
	}

	if cmd.Flags().Changed("slack-issue-comments") {
		v, _ := cmd.Flags().GetBool("slack-issue-comments")
		input.SlackIssueComments = &v
		updated = true
	}

	if cmd.Flags().Changed("slack-issue-statuses") {
		v, _ := cmd.Flags().GetBool("slack-issue-statuses")
		input.SlackIssueStatuses = &v
		updated = true
	}

	if !updated {
		return fmt.Errorf("no fields to update specified")
	}

	result, err := client.ProjectUpdate(ctx, projectID, input)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
}
