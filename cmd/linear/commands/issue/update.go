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
		Long: `Update an existing issue in Linear.

⚠️ CAUTION: This modifies existing issue data. Changes are immediate.

Only specified fields are updated. Unspecified fields remain unchanged.

Examples:
  linear issue update ENG-123 --state=Done
  linear issue update ENG-123 --assignee=alice@company.com --priority=1
  linear issue update ENG-123 --title="New title" --description="Updated description"

TIP: Use 'linear issue get ENG-123 --output=json' to see current values before updating`,
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
	cmd.Flags().StringArray("add-label", []string{}, "Add labels (repeatable)")
	cmd.Flags().StringArray("remove-label", []string{}, "Remove labels (repeatable)")

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

	if !updated {
		return fmt.Errorf("no fields to update specified")
	}

	// Update issue
	result, err := client.IssueUpdate(ctx, issueID, input)
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
