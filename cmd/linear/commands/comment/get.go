package comment

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the comment get command.
func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
	flags := &cli.FieldFlags{}

	cmd := &cobra.Command{
		Use:   "get <comment-id>",
		Short: "Get a single comment by ID",
		Long: `Get comment by UUID. Returns 6 default fields.

Example: go-linear comment get <comment-uuid>

Related: comment_list, issue_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runGet(cmd, client, args[0], flags)
		},
	}

	cmd.Flags().Int64("children-limit", 50, "Max child comments to fetch (0 = fetch all via pagination)")
	flags.Bind(cmd, "defaults (id,body,createdAt,user.name,url,editedAt) | none | defaults,extra")

	return cmd
}
func runGet(cmd *cobra.Command, client *linear.Client, commentID string, flags *cli.FieldFlags) error {
	ctx := cmd.Context()

	childrenLimit, _ := cmd.Flags().GetInt64("children-limit")

	// If limit is 0, fetch all children via pagination
	var comment *intgraphql.GetComment_Comment
	if childrenLimit == 0 {
		// Fetch all children by paginating
		const maxIterations = 100 // Safety limit: 100 pages * 50 = 5000 children max
		allChildren := make([]*intgraphql.GetComment_Comment_Children_Nodes, 0, 50)
		cursor := (*string)(nil)
		batchSize := int64(50)

		for range maxIterations {
			resp, err := client.GetCommentWithChildren(ctx, commentID, &batchSize, cursor)
			if err != nil {
				return fmt.Errorf("failed to get comment: %w", err)
			}

			// First iteration - store the comment
			if comment == nil {
				comment = resp
			}

			allChildren = append(allChildren, resp.Children.Nodes...)

			if !resp.Children.PageInfo.HasNextPage {
				break
			}

			cursor = resp.Children.PageInfo.EndCursor
		}

		if comment == nil {
			return fmt.Errorf("failed to fetch comment")
		}

		// Check if we hit the limit
		if comment.Children.PageInfo.HasNextPage {
			return fmt.Errorf("thread too large: exceeded %d pages (5000 children). Use --children-limit to fetch partial results", maxIterations)
		}

		// Replace children with all fetched
		comment.Children.Nodes = allChildren
	} else {
		var err error
		comment, err = client.GetCommentWithChildren(ctx, commentID, &childrenLimit, nil)
		if err != nil {
			return fmt.Errorf("failed to get comment: %w", err)
		}
	}

	cfg, _ := config.Load()
	var configOverrides map[string]string
	if cfg != nil {
		configOverrides = cfg.FieldDefaults
	}
	defaults := fieldfilter.GetDefaults("comment.get", configOverrides)
	fieldSelector, err := fieldfilter.New(flags.Fields, defaults)
	if err != nil {
		return fmt.Errorf("invalid --fields: %w", err)
	}
	return formatter.FormatJSONFiltered(cmd.OutOrStdout(), comment, true, fieldSelector)
}
