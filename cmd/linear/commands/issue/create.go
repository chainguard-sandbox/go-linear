package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewCreateCommand creates the issue create command.
func NewCreateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issue",
		Long: `Create issue. Safe operation. Uses workspace/user config defaults.

Required: --title
Optional: --team (defaults from config), --description, --assignee=me, --priority (0-4), --state, --label (defaults from config)

Example: go-linear issue create --team=ENG --title="Fix bug" --assignee=me --priority=1 --output=json

Related: issue_get, issue_list, team_list, user_list, label_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runCreate(cmd, client)
		},
	}

	// Required
	cmd.Flags().String("title", "", "Issue title (required)")
	_ = cmd.MarkFlagRequired("title")

	// Team is optional now (can come from config)
	cmd.Flags().String("team", "", "Team name or ID (uses config default if not specified)")

	// Optional
	cmd.Flags().String("description", "", "Issue description (markdown)")
	cmd.Flags().String("assignee", "", "Assignee name, email, or ID")
	cmd.Flags().String("state", "", "Initial state name or ID")
	cmd.Flags().Int("priority", -1, "Priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low")
	cmd.Flags().StringArray("label", []string{}, "Label names (repeatable)")
	cmd.Flags().String("cycle", "", "Cycle UUID")
	cmd.Flags().String("project", "", "Project UUID")
	cmd.Flags().String("parent", "", "Parent issue ID (creates sub-issue)")
	cmd.Flags().Int("estimate", -1, "Story points/estimate")

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()
	res := resolver.New(client)

	// Load config for defaults
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve team (from flag or config default)
	teamName, _ := cmd.Flags().GetString("team")
	if teamName == "" && cfg.Defaults.Team != "" {
		teamName = cfg.Defaults.Team
	}
	if teamName == "" {
		return fmt.Errorf("--team required (or set defaults.team in config)")
	}

	teamID, err := res.ResolveTeam(ctx, teamName)
	if err != nil {
		return fmt.Errorf("failed to resolve team: %w", err)
	}

	// Build input
	title, _ := cmd.Flags().GetString("title")
	input := intgraphql.IssueCreateInput{
		TeamID: teamID,
		Title:  &title,
	}

	// Optional fields
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

	// Labels: merge config defaults with flag values
	labels, _ := cmd.Flags().GetStringArray("label")
	allLabels := append([]string{}, cfg.Defaults.Labels...)
	allLabels = append(allLabels, labels...)

	if len(allLabels) > 0 {
		labelIDs := make([]string, 0, len(allLabels))
		for _, label := range allLabels {
			labelID, err := res.ResolveLabel(ctx, label)
			if err != nil {
				return fmt.Errorf("failed to resolve label %q: %w", label, err)
			}
			labelIDs = append(labelIDs, labelID)
		}
		input.LabelIds = labelIDs
	}

	// Additional optional fields
	if cycle, _ := cmd.Flags().GetString("cycle"); cycle != "" {
		input.CycleID = &cycle
	}

	if project, _ := cmd.Flags().GetString("project"); project != "" {
		input.ProjectID = &project
	}

	if parent, _ := cmd.Flags().GetString("parent"); parent != "" {
		input.ParentID = &parent
	}

	if estimate, _ := cmd.Flags().GetInt("estimate"); estimate >= 0 {
		e := int64(estimate)
		input.Estimate = &e
	}

	// Create issue
	result, err := client.IssueCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}

	// Format output
	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Created issue: %s\n", result.Identifier)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
