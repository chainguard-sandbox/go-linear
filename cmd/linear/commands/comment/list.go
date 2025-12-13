package comment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the comment list command.
func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List comments",
		Long: `List comments. Returns 5 default fields per comment.

Example: go-linear comment list --limit=100 --output=json

Related: comment_get, comment_create, issue_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client)
		},
	}

	cmd.Flags().IntP("limit", "l", 50, "Number of comments to return")
	cmd.Flags().String("after", "", "Cursor for pagination")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,body,createdAt,user.name,url) | none | defaults,extra")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	after, _ := cmd.Flags().GetString("after")
	var afterPtr *string
	if after != "" {
		afterPtr = &after
	}

	comments, err := client.Comments(ctx, &first, afterPtr)
	if err != nil {
		return fmt.Errorf("failed to list comments: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	switch output {
	case "json":
		// Load config for field defaults
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}

		// Get command defaults
		defaults := fieldfilter.GetDefaults("comment.list", configOverrides)

		// Parse field selector with defaults (list command preserves nodes/pageInfo)
		fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}

		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), comments, true, fieldSelector)
	case "table":
		if len(comments.Nodes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No comments found")
			return nil
		}
		for _, comment := range comments.Nodes {
			fmt.Fprintf(cmd.OutOrStdout(), "Comment by %s:\n", comment.User.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n\n", comment.Body)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
