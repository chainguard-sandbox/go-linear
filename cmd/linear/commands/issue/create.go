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

// NewCreateCommand creates the issue create command.
func NewCreateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issue",
		Long: `Create issue. Safe operation.

Required: --team (name/key from team_list), --title
Optional: --description, --assignee=me or email, --priority (0=none, 1=urgent, 2=high, 3=normal, 4=low), --state, --label (repeatable)

Example: go-linear-cli issue create --team=ENG --title="Fix bug" --assignee=me --priority=1 --description="Details" --output=json

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
	cmd.Flags().String("team", "", "Team name or ID (required)")
	_ = cmd.MarkFlagRequired("team")

	cmd.Flags().String("title", "", "Issue title (required)")
	_ = cmd.MarkFlagRequired("title")

	// Optional
	cmd.Flags().String("description", "", "Issue description (markdown)")
	cmd.Flags().String("assignee", "", "Assignee name, email, or ID")
	cmd.Flags().String("state", "", "Initial state name or ID")
	cmd.Flags().Int("priority", -1, "Priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low")
	cmd.Flags().StringArray("label", []string{}, "Label names (repeatable)")

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()
	res := resolver.New(client)

	// Resolve team
	teamName, _ := cmd.Flags().GetString("team")
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
