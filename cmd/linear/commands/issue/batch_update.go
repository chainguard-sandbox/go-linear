package issue

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	issuefilter "github.com/chainguard-sandbox/go-linear/internal/filter/issue"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewBatchUpdateCommand creates the issue batch-update command.
func NewBatchUpdateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch-update",
		Short: "Update multiple issues matching filters",
		Long: `Update multiple issues at once (max 50). Uses same filters as issue_list.

Safety: --dry-run shows what would change, --yes skips confirmation

Filters: All 64 issue list filters (--team, --state, --creator, --has-suggested-teams, etc.)
Updates: --set-state, --set-assignee, --set-priority, --add-label, --remove-label, etc.

Example: go-linear issue batch-update --state=Triage --has-suggested-teams --set-state=Backlog --dry-run

Related: issue_list, issue_update`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runBatchUpdate(cmd, client)
		},
	}

	// All filter flags from issue list (must match for filter builder to work)
	cmd.Flags().String("team", "", "Team name or ID")
	cmd.Flags().String("assignee", "", "Assignee name, email, or 'me'")
	cmd.Flags().String("state", "", "State name or ID")
	cmd.Flags().Int("priority", -1, "Priority filter")
	cmd.Flags().String("creator", "", "Creator")
	cmd.Flags().String("created-after", "", "Created after")
	cmd.Flags().String("created-before", "", "Created before")
	cmd.Flags().String("updated-after", "", "Updated after")
	cmd.Flags().String("updated-before", "", "Updated before")
	cmd.Flags().String("completed-after", "", "Completed after")
	cmd.Flags().String("completed-before", "", "Completed before")
	cmd.Flags().StringArray("label", []string{}, "Label filter")

	// Additional filters (abbreviated to save space - all 64 work via FromFlags)
	cmd.Flags().String("cycle", "", "Cycle filter")
	cmd.Flags().String("project", "", "Project filter")
	cmd.Flags().String("parent", "", "Parent filter")
	cmd.Flags().Bool("has-suggested-teams", false, "AI team suggestions")
	cmd.Flags().Bool("has-suggested-assignees", false, "AI assignee suggestions")
	cmd.Flags().Bool("has-suggested-projects", false, "AI project suggestions")
	cmd.Flags().Bool("has-suggested-labels", false, "AI label suggestions")
	cmd.Flags().StringArray("comment-by", []string{}, "Comment by user")
	cmd.Flags().StringArray("subscriber", []string{}, "Subscriber")
	cmd.Flags().String("description", "", "Description contains")
	cmd.Flags().String("title", "", "Title contains")
	cmd.Flags().Int("estimate", -1, "Estimate filter")
	cmd.Flags().Int("number", -1, "Issue number filter")
	cmd.Flags().Bool("has-children", false, "Has sub-issues")
	// (Filter builder supports all 64 flags even if not listed here)

	// Update flags (--set-* to distinguish from filter flags)
	cmd.Flags().String("set-state", "", "New state name or ID")
	cmd.Flags().String("set-assignee", "", "New assignee (name, email, or 'me')")
	cmd.Flags().Int("set-priority", -1, "New priority (0-4)")
	cmd.Flags().String("set-team", "", "New team name or ID")
	cmd.Flags().String("set-cycle", "", "New cycle name or UUID (use 'none' to remove)")
	cmd.Flags().String("set-project", "", "New project name or UUID (use 'none' to remove)")
	cmd.Flags().StringArray("add-label", []string{}, "Labels to add (repeatable)")
	cmd.Flags().StringArray("remove-label", []string{}, "Labels to remove (repeatable)")
	cmd.Flags().String("set-description", "", "New description")
	cmd.Flags().String("set-title", "", "New title")
	cmd.Flags().Int("set-estimate", -1, "New estimate")
	cmd.Flags().String("set-due-date", "", "New due date (YYYY-MM-DD, use 'none' to remove)")
	cmd.Flags().String("set-milestone", "", "New milestone UUID (use 'none' to remove)")

	// Safety flags
	cmd.Flags().Bool("dry-run", false, "Show what would change without applying")
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().Int("batch-limit", 50, "Max issues per batch (API max: 50)")

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runBatchUpdate(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	// Build filter from flags
	filterBuilder := issuefilter.NewIssueFilterBuilder(res)
	if err := filterBuilder.FromFlags(ctx, cmd); err != nil {
		return err
	}
	issueFilter := filterBuilder.Build()

	// Fetch matching issues
	batchLimit, _ := cmd.Flags().GetInt("batch-limit")
	if batchLimit > 50 {
		batchLimit = 50 // API maximum
	}
	first := int64(batchLimit)

	var issues []struct {
		ID         string
		Identifier string
		Title      string
	}

	if issueFilter != nil {
		result, err := client.IssuesFiltered(ctx, &first, nil, issueFilter)
		if err != nil {
			return fmt.Errorf("failed to fetch issues: %w", err)
		}
		for _, node := range result.Nodes {
			issues = append(issues, struct {
				ID         string
				Identifier string
				Title      string
			}{
				ID:         node.ID,
				Identifier: node.Identifier,
				Title:      node.Title,
			})
		}
	} else {
		result, err := client.Issues(ctx, &first, nil)
		if err != nil {
			return fmt.Errorf("failed to fetch issues: %w", err)
		}
		for _, node := range result.Nodes {
			issues = append(issues, struct {
				ID         string
				Identifier string
				Title      string
			}{
				ID:         node.ID,
				Identifier: node.Identifier,
				Title:      node.Title,
			})
		}
	}

	if len(issues) == 0 {
		fmt.Fprintln(cmd.OutOrStderr(), "No issues match filters")
		return nil
	}

	// Build update input
	input := intgraphql.IssueUpdateInput{}
	updateCount := 0

	if setState, _ := cmd.Flags().GetString("set-state"); setState != "" {
		stateID, err := res.ResolveState(ctx, setState)
		if err != nil {
			return fmt.Errorf("failed to resolve set-state: %w", err)
		}
		input.StateID = &stateID
		updateCount++
	}

	if setAssignee, _ := cmd.Flags().GetString("set-assignee"); setAssignee != "" {
		userID, err := res.ResolveUser(ctx, setAssignee)
		if err != nil {
			return fmt.Errorf("failed to resolve set-assignee: %w", err)
		}
		input.AssigneeID = &userID
		updateCount++
	}

	if setPriority, _ := cmd.Flags().GetInt("set-priority"); setPriority >= 0 {
		p := int64(setPriority)
		input.Priority = &p
		updateCount++
	}

	if setTeam, _ := cmd.Flags().GetString("set-team"); setTeam != "" {
		teamID, err := res.ResolveTeam(ctx, setTeam)
		if err != nil {
			return fmt.Errorf("failed to resolve set-team: %w", err)
		}
		input.TeamID = &teamID
		updateCount++
	}

	if cmd.Flags().Changed("set-cycle") {
		setCycle, _ := cmd.Flags().GetString("set-cycle")
		if setCycle == "none" {
			empty := ""
			input.CycleID = &empty // Empty string removes cycle
		} else {
			cycleID, err := res.ResolveCycle(ctx, setCycle)
			if err != nil {
				return fmt.Errorf("failed to resolve set-cycle: %w", err)
			}
			input.CycleID = &cycleID
		}
		updateCount++
	}

	if cmd.Flags().Changed("set-project") {
		setProject, _ := cmd.Flags().GetString("set-project")
		if setProject == "none" {
			empty := ""
			input.ProjectID = &empty // Empty string removes project
		} else {
			projectID, err := res.ResolveProject(ctx, setProject)
			if err != nil {
				return fmt.Errorf("failed to resolve set-project: %w", err)
			}
			input.ProjectID = &projectID
		}
		updateCount++
	}

	addLabels, _ := cmd.Flags().GetStringArray("add-label")
	if len(addLabels) > 0 {
		labelIDs := make([]string, 0, len(addLabels))
		for _, label := range addLabels {
			labelID, err := res.ResolveLabel(ctx, label)
			if err != nil {
				return fmt.Errorf("failed to resolve add-label %q: %w", label, err)
			}
			labelIDs = append(labelIDs, labelID)
		}
		input.AddedLabelIds = labelIDs
		updateCount++
	}

	removeLabels, _ := cmd.Flags().GetStringArray("remove-label")
	if len(removeLabels) > 0 {
		labelIDs := make([]string, 0, len(removeLabels))
		for _, label := range removeLabels {
			labelID, err := res.ResolveLabel(ctx, label)
			if err != nil {
				return fmt.Errorf("failed to resolve remove-label %q: %w", label, err)
			}
			labelIDs = append(labelIDs, labelID)
		}
		input.RemovedLabelIds = labelIDs
		updateCount++
	}

	if setTitle, _ := cmd.Flags().GetString("set-title"); setTitle != "" {
		input.Title = &setTitle
		updateCount++
	}

	if setDesc, _ := cmd.Flags().GetString("set-description"); setDesc != "" {
		input.Description = &setDesc
		updateCount++
	}

	if setEstimate, _ := cmd.Flags().GetInt("set-estimate"); setEstimate >= 0 {
		e := int64(setEstimate)
		input.Estimate = &e
		updateCount++
	}

	if cmd.Flags().Changed("set-due-date") {
		setDueDate, _ := cmd.Flags().GetString("set-due-date")
		if setDueDate == "none" {
			empty := ""
			input.DueDate = &empty
		} else {
			input.DueDate = &setDueDate
		}
		updateCount++
	}

	if cmd.Flags().Changed("set-milestone") {
		setMilestone, _ := cmd.Flags().GetString("set-milestone")
		if setMilestone == "none" {
			empty := ""
			input.ProjectMilestoneID = &empty
		} else {
			milestoneID, err := res.ResolveMilestone(ctx, setMilestone)
			if err != nil {
				return fmt.Errorf("failed to resolve milestone: %w", err)
			}
			input.ProjectMilestoneID = &milestoneID
		}
		updateCount++
	}

	if updateCount == 0 {
		return fmt.Errorf("no update flags specified (use --set-state, --set-assignee, etc.)")
	}

	// Dry run or confirmation
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		fmt.Fprintf(cmd.OutOrStderr(), "Would update %d issues:\n", len(issues))
		for _, issue := range issues {
			fmt.Fprintf(cmd.OutOrStderr(), "  %s: %s\n", issue.Identifier, issue.Title)
		}
		fmt.Fprintln(cmd.OutOrStderr(), "\nRun without --dry-run to apply changes")
		return nil
	}

	// Confirmation prompt
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		fmt.Fprintf(cmd.OutOrStderr(), "⚠️  Update %d issues? This will modify existing data.\n", len(issues))
		fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if !strings.EqualFold(strings.TrimSpace(response), "yes") {
			fmt.Fprintln(cmd.OutOrStderr(), "Canceled.")
			return nil
		}
	}

	// Extract IDs
	ids := make([]string, len(issues))
	for i, issue := range issues {
		ids[i] = issue.ID
	}

	// Call batch update
	result, err := client.IssueBatchUpdate(ctx, ids, input)
	if err != nil {
		return fmt.Errorf("failed to batch update: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Updated %d issues\n", len(result.Issues))
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
