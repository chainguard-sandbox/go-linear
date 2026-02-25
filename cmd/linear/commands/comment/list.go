package comment

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	commentfilter "github.com/chainguard-sandbox/go-linear/internal/filter/comment"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the comment list command.
func NewListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	fieldFlags := &cli.FieldFlags{}
	paginationFlags := &cli.PaginationFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List comments with filtering",
		Long: `List comments with filtering. Returns 5 default fields per comment.

Filters: --body, --creator, --issue
Date filters: --created-after, --created-before, --updated-after, --updated-before (date formats: see issue_list)

Example: go-linear comment list --created-after=7d

Related: comment_get, comment_create, issue_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client, fieldFlags, paginationFlags)
		},
	}

	// Pagination

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
	paginationFlags.Bind(cmd, 50)
	fieldFlags.Bind(cmd, "defaults (...) | none | defaults,extra")
	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client, fieldFlags *cli.FieldFlags, paginationFlags *cli.PaginationFlags) error {
	ctx := cmd.Context()

	res := resolver.New(client)

	// Build filter from flags
	filterBuilder := commentfilter.NewFilterBuilder(res)
	if err := filterBuilder.FromFlags(ctx, cmd); err != nil {
		return err
	}
	commentFilter := filterBuilder.Build()

	first := paginationFlags.LimitPtr()

	var afterPtr *string
	if paginationFlags.After != "" {
		afterPtr = &paginationFlags.After
	}

	// Use filtered or unfiltered query based on whether filters were set
	if commentFilter != nil {
		comments, err := client.CommentsFiltered(ctx, first, afterPtr, commentFilter)
		if err != nil {
			return fmt.Errorf("failed to list comments: %w", err)
		}

		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("comment.list", configOverrides)
		fieldSelector, err := fieldfilter.NewForList(fieldFlags.Fields, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), comments, true, fieldSelector)
	}

	// No filters: use regular query
	comments, err := client.Comments(ctx, first, afterPtr)
	if err != nil {
		return fmt.Errorf("failed to list comments: %w", err)
	}

	cfg, _ := config.Load()
	var configOverrides map[string]string
	if cfg != nil {
		configOverrides = cfg.FieldDefaults
	}
	defaults := fieldfilter.GetDefaults("comment.list", configOverrides)
	fieldSelector, err := fieldfilter.NewForList(fieldFlags.Fields, defaults)
	if err != nil {
		return fmt.Errorf("invalid --fields: %w", err)
	}
	return formatter.FormatJSONFiltered(cmd.OutOrStdout(), comments, true, fieldSelector)
}
