package comment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	commentfilter "github.com/chainguard-sandbox/go-linear/internal/filter/comment"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the comment list command.
func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List comments with filtering",
		Long: `List comments with filtering. Returns 5 default fields per comment.

Filters: --body, --creator, --issue
Date filters: --created-after, --created-before, --updated-after, --updated-before (date formats: see issue_list)

Example: go-linear comment list --created-after=7d --output=json

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

	// Pagination
	cmd.Flags().IntP("limit", "l", 50, "Number of comments to return")
	cmd.Flags().String("after", "", "Cursor for pagination")

	// Date filters
	cmd.Flags().String("created-after", "", "Created after date (ISO8601, 'yesterday', '7d')")
	cmd.Flags().String("created-before", "", "Created before date")
	cmd.Flags().String("updated-after", "", "Updated after date")
	cmd.Flags().String("updated-before", "", "Updated before date")

	// Entity filters
	cmd.Flags().String("id", "", "Comment UUID")
	cmd.Flags().String("creator", "", "Creator name, email, or 'me'")
	cmd.Flags().String("issue", "", "Issue identifier or UUID")

	// Text filters
	cmd.Flags().String("body", "", "Body contains (case-insensitive)")

	// Output
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,body,createdAt,user.name,url) | none | defaults,extra")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()
	res := resolver.New(client)

	// Build filter from flags
	filterBuilder := commentfilter.NewFilterBuilder(res)
	if err := filterBuilder.FromFlags(ctx, cmd); err != nil {
		return err
	}
	commentFilter := filterBuilder.Build()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	after, _ := cmd.Flags().GetString("after")
	var afterPtr *string
	if after != "" {
		afterPtr = &after
	}

	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	// Use filtered or unfiltered query based on whether filters were set
	if commentFilter != nil {
		comments, err := client.CommentsFiltered(ctx, &first, afterPtr, commentFilter)
		if err != nil {
			return fmt.Errorf("failed to list comments: %w", err)
		}

		switch output {
		case "json":
			cfg, _ := config.Load()
			var configOverrides map[string]string
			if cfg != nil {
				configOverrides = cfg.FieldDefaults
			}
			defaults := fieldfilter.GetDefaults("comment.list", configOverrides)
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

	// No filters: use regular query
	comments, err := client.Comments(ctx, &first, afterPtr)
	if err != nil {
		return fmt.Errorf("failed to list comments: %w", err)
	}

	switch output {
	case "json":
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("comment.list", configOverrides)
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
