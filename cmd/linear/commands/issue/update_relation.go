package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUpdateRelationCommand creates the issue update-relation command.
func NewUpdateRelationCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-relation <relation-id>",
		Short: "Update the type of an existing issue relationship",
		Long: `Update issue relationship type. Modifies existing data.

Required: --type (blocks|blocked-by|duplicate|related, see issue_relate)

Example: go-linear-cli issue update-relation <relation-uuid> --type=blocks --output=json

Related: issue_relate, issue_unrelate, issue_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runUpdateRelation(cmd, client, args[0])
		},
	}

	_ = cmd.MarkFlagRequired("type")
	cmd.Flags().String("type", "", "New relation type: blocks|blocked-by|duplicate|related (required)")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runUpdateRelation(cmd *cobra.Command, client *linear.Client, relationID string) error {
	ctx := context.Background()

	relationType, _ := cmd.Flags().GetString("type")

	// Convert string to relation type
	var typeStr string
	switch relationType {
	case "blocks":
		typeStr = "blocks"
	case "blocked-by":
		typeStr = "blocked"
	case "duplicate":
		typeStr = "duplicate"
	case "related":
		typeStr = "related"
	default:
		return fmt.Errorf("invalid relation type: %s (must be: blocks, blocked-by, duplicate, or related)", relationType)
	}

	input := intgraphql.IssueRelationUpdateInput{
		Type: &typeStr,
	}

	result, err := client.IssueRelationUpdate(ctx, relationID, input)
	if err != nil {
		return fmt.Errorf("failed to update issue relation: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Updated relation to type: %s\n", relationType)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
