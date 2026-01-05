package favorite

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewCreateCommand creates the favorite create command.
func NewCreateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Star an issue, project, cycle, or document",
		Long: `Star item. Safe operation. Must specify exactly one of: --issue, --project, --cycle, --document.

Example: go-linear favorite create --issue=ENG-123 --output=json

Related: favorite_delete, issue_list, project_list`,
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
	ctx := cmd.Context()

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

	// Resolve identifiers to UUIDs
	res := resolver.New(client)
	if issueID != "" {
		resolvedID, err := res.ResolveIssue(ctx, issueID)
		if err != nil {
			return fmt.Errorf("failed to resolve issue: %w", err)
		}
		issueID = resolvedID
	}
	if projectID != "" {
		resolvedID, err := res.ResolveProject(ctx, projectID)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}
		projectID = resolvedID
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
