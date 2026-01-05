package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/dateparser"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewMilestoneUpdateCommand creates the project milestone-update command.
func NewMilestoneUpdateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "milestone-update <milestone-id>",
		Short: "Update a project milestone",
		Long: `Update project milestone. Modifies existing data.

Fields: --name, --description, --target-date (date formats: see issue_list)

Example: go-linear project milestone-update <uuid> --name="Q2 2025" --target-date=14d --output=json

Related: project_milestone-create, project_milestone-delete, project_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runMilestoneUpdate(cmd, client, args[0])
		},
	}

	cmd.Flags().String("name", "", "New milestone name")
	cmd.Flags().String("description", "", "New description (markdown)")
	cmd.Flags().String("target-date", "", "New target date (ISO8601 or relative)")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runMilestoneUpdate(cmd *cobra.Command, client *linear.Client, milestoneID string) error {
	ctx := cmd.Context()

	input := intgraphql.ProjectMilestoneUpdateInput{}

	if name, _ := cmd.Flags().GetString("name"); name != "" {
		input.Name = &name
	}

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		input.Description = &desc
	}

	if targetDateStr, _ := cmd.Flags().GetString("target-date"); targetDateStr != "" {
		parser := dateparser.New()
		targetDate, err := parser.Parse(targetDateStr)
		if err != nil {
			return fmt.Errorf("invalid target-date: %w", err)
		}
		dateStr := targetDate.Format("2006-01-02")
		input.TargetDate = &dateStr
	}

	result, err := client.ProjectMilestoneUpdate(ctx, milestoneID, input)
	if err != nil {
		return fmt.Errorf("failed to update milestone: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Updated milestone\n")
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
