package comment

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewUpdateCommand creates the comment update command.
func NewUpdateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing comment",
		Long: `Update comment text. Supports markdown. Modifies existing data.

Required: --body

Example: go-linear comment update <uuid> --body="Updated text"

Related: comment_get, comment_create`,
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

	cmd.Flags().String("body", "", "New comment body text (markdown) (required)")
	_ = cmd.MarkFlagRequired("body")

	return cmd
}

func runUpdate(cmd *cobra.Command, client *linear.Client, commentID string) error {
	ctx := cmd.Context()

	body, _ := cmd.Flags().GetString("body")

	input := intgraphql.CommentUpdateInput{
		Body: &body,
	}

	result, err := client.CommentUpdate(ctx, commentID, input)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
}
