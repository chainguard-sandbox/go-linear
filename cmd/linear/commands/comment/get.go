package comment

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the comment get command.
func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
	flags := &cli.OutputFlags{}

	cmd := &cobra.Command{
		Use:   "get <comment-id>",
		Short: "Get a single comment by ID",
		Long: `Get comment by UUID. Returns 6 default fields.

Example: go-linear comment get <comment-uuid> --output=json

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

	flags.Bind(cmd, "defaults (id,body,createdAt,user.name,url,editedAt) | none | defaults,extra")

	return cmd
}
func runGet(cmd *cobra.Command, client *linear.Client, commentID string, flags *cli.OutputFlags) error {
	ctx := cmd.Context()

	if err := flags.Validate(); err != nil {
		return err
	}

	comment, err := client.Comment(ctx, commentID)
	if err != nil {
		return fmt.Errorf("failed to get comment: %w", err)
	}

	switch flags.Output {
	case "json":
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
	case "table":
		// Simple table output for single comment
		fmt.Fprintf(cmd.OutOrStdout(), "ID:          %s\n", comment.ID)
		if comment.User != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Author:      %s\n", comment.User.Name)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Body:        %s\n", comment.Body)
		fmt.Fprintf(cmd.OutOrStdout(), "Created:     %s\n", comment.CreatedAt)
		fmt.Fprintf(cmd.OutOrStdout(), "Updated:     %s\n", comment.UpdatedAt)

		// Show parent comment if this is a reply
		if comment.Parent != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "\nReplying to: %s\n", comment.Parent.User.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", comment.Parent.Body)
		}

		// Show child comments (replies)
		if len(comment.Children.Nodes) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "\nReplies: %d\n", len(comment.Children.Nodes))
			for _, child := range comment.Children.Nodes {
				fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s: %s\n",
					child.CreatedAt.Format("Jan 2"), child.User.Name, child.Body)
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", flags.Output)
	}
}
