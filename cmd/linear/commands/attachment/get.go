package attachment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the attachment get command.
func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <attachment-id>",
		Short: "Get a single attachment by ID",
		Long: `Get detailed information about a specific attachment.

Retrieve full attachment details including title, URL, subtitle, metadata, and associated issue.
Attachments can be files, external links, GitHub PRs, or Slack threads.

Parameters:
  <attachment-id>: Attachment UUID (required)

Output (--output=json):
  Returns JSON with: id, title, url, subtitle, source, issue, creator, createdAt

Examples:
  # Get attachment by UUID
  linear attachment get <attachment-uuid>

  # Get attachment with JSON output
  linear attachment get <attachment-uuid> --output=json

TIP: Use 'linear attachment list' to discover attachment IDs from issues

Related Commands:
  - linear attachment list - List all attachments
  - linear attachment link-url - Link external URL to issue
  - linear attachment link-github - Link GitHub PR to issue`,
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

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, attachmentID string) error {
	ctx := context.Background()

	attachment, err := client.Attachment(ctx, attachmentID)
	if err != nil {
		return fmt.Errorf("failed to get attachment: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), attachment, true)
	case "table":
		// Simple table output for single attachment
		fmt.Fprintf(cmd.OutOrStdout(), "ID:          %s\n", attachment.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "Title:       %s\n", attachment.Title)
		fmt.Fprintf(cmd.OutOrStdout(), "URL:         %s\n", attachment.URL)
		if attachment.Subtitle != nil && *attachment.Subtitle != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Subtitle:    %s\n", *attachment.Subtitle)
		}
		if attachment.SourceType != nil && *attachment.SourceType != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Source:      %s\n", *attachment.SourceType)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created:     %s\n", attachment.CreatedAt)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
