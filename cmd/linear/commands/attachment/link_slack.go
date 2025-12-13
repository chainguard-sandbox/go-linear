package attachment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewLinkSlackCommand creates the attachment link-slack command.
func NewLinkSlackCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link-slack",
		Short: "Link a Slack message to an issue",
		Long: `Link Slack message to issue. Requires Slack integration. Safe operation.

Required: --issue (ID from issue_list), --url (Slack permalink)

Example: go-linear attachment link-slack --issue=ENG-123 --url=https://workspace.slack.com/archives/C123/p1234567890 --output=json

Related: attachment_link-url, attachment_link-github, issue_get`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runLinkSlack(cmd, client)
		},
	}

	_ = cmd.MarkFlagRequired("issue")
	_ = cmd.MarkFlagRequired("url")
	cmd.Flags().String("issue", "", "Issue identifier or UUID (required)")
	cmd.Flags().String("url", "", "Slack message URL (required)")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runLinkSlack(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	issueID, _ := cmd.Flags().GetString("issue")
	url, _ := cmd.Flags().GetString("url")

	result, err := client.AttachmentLinkSlack(ctx, issueID, url)
	if err != nil {
		return fmt.Errorf("failed to link Slack message: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Linked Slack message to issue %s\n", issueID)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
