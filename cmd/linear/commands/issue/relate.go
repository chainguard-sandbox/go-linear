package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewRelateCommand creates the issue relation command.
func NewRelateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relate <issue-id> <related-issue-id>",
		Short: "Create a relationship between two issues",
		Long: `Relate two issues. Safe operation.

Types: blocks | blocked-by | duplicate | related (default)

Example: go-linear-cli issue relate ENG-123 ENG-124 --type=blocks --output=json

Related: issue_unrelate, issue_update-relation`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runRelate(cmd, client, args[0], args[1])
		},
	}

	cmd.Flags().String("type", "related", "Relation type: blocks|blocked-by|duplicate|related")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runRelate(cmd *cobra.Command, client *linear.Client, issueID, relatedIssueID string) error {
	ctx := context.Background()

	relationType, _ := cmd.Flags().GetString("type")

	// Convert string to IssueRelationType enum
	var relationTypeEnum intgraphql.IssueRelationType
	switch relationType {
	case "blocks":
		relationTypeEnum = "blocks"
	case "blocked-by":
		relationTypeEnum = "blocked"
	case "duplicate":
		relationTypeEnum = "duplicate"
	case "related":
		relationTypeEnum = "related"
	default:
		return fmt.Errorf("invalid relation type: %s (must be: blocks, blocked-by, duplicate, or related)", relationType)
	}

	input := intgraphql.IssueRelationCreateInput{
		IssueID:        issueID,
		RelatedIssueID: relatedIssueID,
		Type:           relationTypeEnum,
	}

	result, err := client.IssueRelationCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create issue relation: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Created %s relation between %s and %s\n", relationType, issueID, relatedIssueID)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
