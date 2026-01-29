package issue

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUpdateCommand creates the issue update command.
func NewUpdateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing issue",
		Long: `Update issue. Modifies existing data.

Fields: --title, --description, --assignee=me, --state, --priority (0-4), --cycle, --project, --parent, --due-date (YYYY-MM-DD or 'none'), --milestone (uuid or 'none'), --add-label, --remove-label, --link-pr

Example: go-linear issue update ENG-123 --state=Done --due-date=2025-03-01 --output=json

Related: issue_get, issue_create`,
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

	// All optional update fields
	cmd.Flags().String("title", "", "New title")
	cmd.Flags().String("description", "", "New description (markdown)")
	cmd.Flags().String("assignee", "", "New assignee name, email, or ID")
	cmd.Flags().String("state", "", "New state name or ID")
	cmd.Flags().Int("priority", -1, "New priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low")
	cmd.Flags().String("cycle", "", "Cycle name or UUID (use 'none' to remove)")
	cmd.Flags().String("project", "", "Project name or UUID (use 'none' to remove)")
	cmd.Flags().String("parent", "", "Parent issue ID/identifier (use 'none' to remove)")
	cmd.Flags().StringArray("add-label", []string{}, "Add labels (repeatable)")
	cmd.Flags().StringArray("remove-label", []string{}, "Remove labels (repeatable)")
	cmd.Flags().String("link-pr", "", "Link GitHub PR (format: owner/repo#number or full URL)")
	cmd.Flags().String("due-date", "", "Due date (YYYY-MM-DD, use 'none' to remove)")
	cmd.Flags().String("milestone", "", "Project milestone UUID (use 'none' to remove)")

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runUpdate(cmd *cobra.Command, client *linear.Client, issueID string) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	// Resolve issue ID (converts identifier like ENG-123 to UUID)
	resolvedIssueID, err := res.ResolveIssue(ctx, issueID)
	if err != nil {
		return fmt.Errorf("failed to resolve issue: %w", err)
	}

	// Check if we need nullable support (for removing parent/cycle/project with 'none')
	needsNullable := false
	if cmd.Flags().Changed("parent") {
		if parent, _ := cmd.Flags().GetString("parent"); parent == "none" {
			needsNullable = true
		}
	}
	if cmd.Flags().Changed("cycle") {
		if cycle, _ := cmd.Flags().GetString("cycle"); cycle == "none" {
			needsNullable = true
		}
	}
	if cmd.Flags().Changed("project") {
		if project, _ := cmd.Flags().GetString("project"); project == "none" {
			needsNullable = true
		}
	}
	if cmd.Flags().Changed("due-date") {
		if dueDate, _ := cmd.Flags().GetString("due-date"); dueDate == "none" {
			needsNullable = true
		}
	}
	if cmd.Flags().Changed("milestone") {
		if milestone, _ := cmd.Flags().GetString("milestone"); milestone == "none" {
			needsNullable = true
		}
	}

	if needsNullable {
		return runUpdateWithNullable(cmd, client, resolvedIssueID, res)
	}

	// Build standard input
	input := intgraphql.IssueUpdateInput{}
	updated := false

	if title, _ := cmd.Flags().GetString("title"); title != "" {
		input.Title = &title
		updated = true
	}

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		input.Description = &desc
		updated = true
	}

	if assignee, _ := cmd.Flags().GetString("assignee"); assignee != "" {
		userID, err := res.ResolveUser(ctx, assignee)
		if err != nil {
			return fmt.Errorf("failed to resolve assignee: %w", err)
		}
		input.AssigneeID = &userID
		updated = true
	}

	if state, _ := cmd.Flags().GetString("state"); state != "" {
		stateID, err := res.ResolveState(ctx, state)
		if err != nil {
			return fmt.Errorf("failed to resolve state: %w", err)
		}
		input.StateID = &stateID
		updated = true
	}

	if priority, _ := cmd.Flags().GetInt("priority"); priority >= 0 {
		p := int64(priority)
		input.Priority = &p
		updated = true
	}

	addLabels, _ := cmd.Flags().GetStringArray("add-label")
	if len(addLabels) > 0 {
		labelIDs := make([]string, 0, len(addLabels))
		for _, label := range addLabels {
			labelID, err := res.ResolveLabel(ctx, label)
			if err != nil {
				return fmt.Errorf("failed to resolve label %q: %w", label, err)
			}
			labelIDs = append(labelIDs, labelID)
		}
		input.AddedLabelIds = labelIDs
		updated = true
	}

	removeLabels, _ := cmd.Flags().GetStringArray("remove-label")
	if len(removeLabels) > 0 {
		labelIDs := make([]string, 0, len(removeLabels))
		for _, label := range removeLabels {
			labelID, err := res.ResolveLabel(ctx, label)
			if err != nil {
				return fmt.Errorf("failed to resolve label %q: %w", label, err)
			}
			labelIDs = append(labelIDs, labelID)
		}
		input.RemovedLabelIds = labelIDs
		updated = true
	}

	// Cycle assignment (supports 'none' to remove)
	if cmd.Flags().Changed("cycle") {
		cycle, _ := cmd.Flags().GetString("cycle")
		if cycle == "none" {
			empty := ""
			input.CycleID = &empty // Set to empty string to remove cycle
		} else {
			// Resolve cycle name or UUID
			cycleID, err := res.ResolveCycle(ctx, cycle)
			if err != nil {
				return fmt.Errorf("failed to resolve cycle: %w", err)
			}
			input.CycleID = &cycleID
		}
		updated = true
	}

	// Project assignment (supports 'none' to remove)
	if cmd.Flags().Changed("project") {
		project, _ := cmd.Flags().GetString("project")
		if project == "none" {
			empty := ""
			input.ProjectID = &empty // Set to empty string to remove project
		} else {
			// Resolve project name or UUID
			projectID, err := res.ResolveProject(ctx, project)
			if err != nil {
				return fmt.Errorf("failed to resolve project: %w", err)
			}
			input.ProjectID = &projectID
		}
		updated = true
	}

	// Parent assignment (supports 'none' to remove)
	if cmd.Flags().Changed("parent") {
		parent, _ := cmd.Flags().GetString("parent")
		if parent == "none" {
			empty := ""
			input.ParentID = &empty // Set to empty string to remove parent
		} else {
			// Resolve parent issue identifier to UUID
			parentID, err := res.ResolveIssue(ctx, parent)
			if err != nil {
				return fmt.Errorf("failed to resolve parent issue: %w", err)
			}
			input.ParentID = &parentID
		}
		updated = true
	}

	// Due date assignment (supports 'none' to remove)
	if cmd.Flags().Changed("due-date") {
		dueDate, _ := cmd.Flags().GetString("due-date")
		if dueDate == "none" {
			empty := ""
			input.DueDate = &empty
		} else {
			input.DueDate = &dueDate
		}
		updated = true
	}

	// Milestone assignment (supports 'none' to remove)
	if cmd.Flags().Changed("milestone") {
		milestone, _ := cmd.Flags().GetString("milestone")
		if milestone == "none" {
			empty := ""
			input.ProjectMilestoneID = &empty
		} else {
			milestoneID, err := res.ResolveMilestone(ctx, milestone)
			if err != nil {
				return fmt.Errorf("failed to resolve milestone: %w", err)
			}
			input.ProjectMilestoneID = &milestoneID
		}
		updated = true
	}

	if !updated {
		return fmt.Errorf("no fields to update specified")
	}

	// Update issue
	result, err := client.IssueUpdate(ctx, resolvedIssueID, input)
	if err != nil {
		return fmt.Errorf("failed to update issue: %w", err)
	}

	// Link GitHub PR if specified
	if prURL, _ := cmd.Flags().GetString("link-pr"); prURL != "" {
		// Convert short format (owner/repo#123) to full URL if needed
		fullURL := prURL
		if !contains(prURL, "://") {
			// Assume GitHub format: owner/repo#123
			fullURL = "https://github.com/" + prURL
		}

		_, err := client.AttachmentLinkGitHubPR(ctx, resolvedIssueID, fullURL)
		if err != nil {
			return fmt.Errorf("failed to link GitHub PR: %w", err)
		}
	}

	// Format output
	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Updated issue: %s\n", result.Identifier)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// runUpdateWithNullable handles updates that require explicit null (e.g., removing parent).
func runUpdateWithNullable(cmd *cobra.Command, client *linear.Client, issueID string, res *resolver.Resolver) error {
	ctx := cmd.Context()

	// Build nullable input
	input := linear.IssueUpdateNullableInput{}

	if title, _ := cmd.Flags().GetString("title"); title != "" {
		input.Title = &title
	}

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		input.Description = &desc
	}

	if assignee, _ := cmd.Flags().GetString("assignee"); assignee != "" {
		userID, err := res.ResolveUser(ctx, assignee)
		if err != nil {
			return fmt.Errorf("failed to resolve assignee: %w", err)
		}
		input.AssigneeID = &userID
	}

	if state, _ := cmd.Flags().GetString("state"); state != "" {
		stateID, err := res.ResolveState(ctx, state)
		if err != nil {
			return fmt.Errorf("failed to resolve state: %w", err)
		}
		input.StateID = &stateID
	}

	if priority, _ := cmd.Flags().GetInt("priority"); priority >= 0 {
		p := int64(priority)
		input.Priority = &p
	}

	addLabels, _ := cmd.Flags().GetStringArray("add-label")
	if len(addLabels) > 0 {
		labelIDs := make([]string, 0, len(addLabels))
		for _, label := range addLabels {
			labelID, err := res.ResolveLabel(ctx, label)
			if err != nil {
				return fmt.Errorf("failed to resolve label %q: %w", label, err)
			}
			labelIDs = append(labelIDs, labelID)
		}
		input.AddedLabelIds = labelIDs
	}

	removeLabels, _ := cmd.Flags().GetStringArray("remove-label")
	if len(removeLabels) > 0 {
		labelIDs := make([]string, 0, len(removeLabels))
		for _, label := range removeLabels {
			labelID, err := res.ResolveLabel(ctx, label)
			if err != nil {
				return fmt.Errorf("failed to resolve label %q: %w", label, err)
			}
			labelIDs = append(labelIDs, labelID)
		}
		input.RemovedLabelIds = labelIDs
	}

	// Nullable fields - support 'none' for removal
	if cmd.Flags().Changed("cycle") {
		cycle, _ := cmd.Flags().GetString("cycle")
		if cycle == "none" {
			input.CycleID = linear.NewNull[string]()
		} else {
			// Resolve cycle name or UUID
			cycleID, err := res.ResolveCycle(ctx, cycle)
			if err != nil {
				return fmt.Errorf("failed to resolve cycle: %w", err)
			}
			input.CycleID = linear.NewValue(cycleID)
		}
	}

	if cmd.Flags().Changed("project") {
		project, _ := cmd.Flags().GetString("project")
		if project == "none" {
			input.ProjectID = linear.NewNull[string]()
		} else {
			// Resolve project name or UUID
			projectID, err := res.ResolveProject(ctx, project)
			if err != nil {
				return fmt.Errorf("failed to resolve project: %w", err)
			}
			input.ProjectID = linear.NewValue(projectID)
		}
	}

	if cmd.Flags().Changed("parent") {
		parent, _ := cmd.Flags().GetString("parent")
		if parent == "none" {
			input.ParentID = linear.NewNull[string]()
		} else {
			parentID, err := res.ResolveIssue(ctx, parent)
			if err != nil {
				return fmt.Errorf("failed to resolve parent issue: %w", err)
			}
			input.ParentID = linear.NewValue(parentID)
		}
	}

	if cmd.Flags().Changed("due-date") {
		dueDate, _ := cmd.Flags().GetString("due-date")
		if dueDate == "none" {
			input.DueDate = linear.NewNull[string]()
		} else {
			input.DueDate = linear.NewValue(dueDate)
		}
	}

	if cmd.Flags().Changed("milestone") {
		milestone, _ := cmd.Flags().GetString("milestone")
		if milestone == "none" {
			input.ProjectMilestoneID = linear.NewNull[string]()
		} else {
			milestoneID, err := res.ResolveMilestone(ctx, milestone)
			if err != nil {
				return fmt.Errorf("failed to resolve milestone: %w", err)
			}
			input.ProjectMilestoneID = linear.NewValue(milestoneID)
		}
	}

	// Use nullable update method
	result, err := client.IssueUpdateNullable(ctx, issueID, input)
	if err != nil {
		return fmt.Errorf("failed to update issue: %w", err)
	}

	// Format output
	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Updated issue: %s\n", result.Identifier)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
