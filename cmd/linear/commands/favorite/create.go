package favorite

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewCreateCommand creates the favorite create command.
func NewCreateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Star an issue, project, cycle, or document",
		Long: `Star (favorite) an item for quick access in Linear.

This operation creates new favorite data and is safe to execute.
Starred items appear in Linear's "Favorites" section for easy retrieval.

Use favorites for:
- Bookmarking important issues to track
- Quick access to active projects
- Pinning frequently referenced documents
- Following current sprint cycles

Parameters (mutually exclusive - specify exactly one):
  --issue: Issue identifier (e.g., ENG-123) or UUID
  --project: Project name or UUID
  --cycle: Cycle UUID
  --document: Document UUID

Examples:
  # Star an issue for quick access
  linear favorite create --issue=ENG-123

  # Star a project
  linear favorite create --project="Platform Redesign"

  # Star with JSON output
  linear favorite create --issue=ENG-123 --output=json

TIP: Favorites sync across web, desktop, and mobile Linear apps

Common Errors:
  - "must specify exactly one resource": Provide --issue OR --project, not both
  - "resource not found": Check ID/name is valid

Related Commands:
  - linear favorite delete - Unstar an item
  - linear issue list - Find issues to favorite
  - linear project list - Find projects to favorite`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runCreate(cmd, client)
		},
	}

	cmd.Flags().String("issue", "", "Issue identifier or UUID")
	cmd.Flags().String("project", "", "Project name or UUID")
	cmd.Flags().String("cycle", "", "Cycle UUID")
	cmd.Flags().String("document", "", "Document UUID")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	issueID, _ := cmd.Flags().GetString("issue")
	projectID, _ := cmd.Flags().GetString("project")
	cycleID, _ := cmd.Flags().GetString("cycle")
	documentID, _ := cmd.Flags().GetString("document")

	// Validate exactly one resource type specified
	resourceCount := 0
	resourceType := ""
	if issueID != "" {
		resourceCount++
		resourceType = "issue"
	}
	if projectID != "" {
		resourceCount++
		resourceType = "project"
	}
	if cycleID != "" {
		resourceCount++
		resourceType = "cycle"
	}
	if documentID != "" {
		resourceCount++
		resourceType = "document"
	}

	if resourceCount == 0 {
		return fmt.Errorf("must specify one of: --issue, --project, --cycle, or --document")
	}
	if resourceCount > 1 {
		return fmt.Errorf("must specify exactly one resource type, not multiple")
	}

	// Build input
	input := intgraphql.FavoriteCreateInput{}
	switch resourceType {
	case "issue":
		input.IssueID = &issueID
	case "project":
		input.ProjectID = &projectID
	case "cycle":
		input.CycleID = &cycleID
	case "document":
		input.DocumentID = &documentID
	}

	result, err := client.FavoriteCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create favorite: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Starred %s\n", resourceType)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
