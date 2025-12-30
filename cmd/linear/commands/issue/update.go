package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUpdateCommand creates the issue update command.
func NewUpdateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing issue",
		Long: `Update issue. Modifies existing data.

Fields: --title, --description, --assignee=me, --state, --priority (0-4), --cycle, --project, --add-label, --remove-label, --link-pr (owner/repo#123)

Example: go-linear issue update ENG-123 --state=Done --cycle=none --link-pr=org/repo#123 --output=json

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
	cmd.Flags().String("cycle", "", "Cycle UUID (use 'none' to remove)")
	cmd.Flags().String("project", "", "Project UUID (use 'none' to remove)")
	cmd.Flags().StringArray("add-label", []string{}, "Add labels (repeatable)")
	cmd.Flags().StringArray("remove-label", []string{}, "Remove labels (repeatable)")
	cmd.Flags().String("link-pr", "", "Link GitHub PR (format: owner/repo#number or full URL)")

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runUpdate(cmd *cobra.Command, client *linear.Client, issueID string) error {
	ctx := context.Background()
	res := resolver.New(client)

	// Build input
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
			input.CycleID = &cycle
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
			input.ProjectID = &project
		}
		updated = true
	}

	if !updated {
		return fmt.Errorf("no fields to update specified")
	}

	// Update issue
	result, err := client.IssueUpdate(ctx, issueID, input)
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

		_, err := client.AttachmentLinkGitHubPR(ctx, issueID, fullURL)
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
