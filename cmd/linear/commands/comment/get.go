package comment

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the comment get command.
func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
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

			return runGet(cmd, client, args[0])
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,body,createdAt,user.name,url,editedAt) | none | defaults,extra")

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, commentID string) error {
	ctx := cmd.Context()

	comment, err := client.Comment(ctx, commentID)
	if err != nil {
		return fmt.Errorf("failed to get comment: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	switch output {
	case "json":
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("comment.get", configOverrides)
		fieldSelector, err := fieldfilter.New(fieldsSpec, defaults)
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
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
