package attachment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewCreateCommand creates the attachment create command.
func NewCreateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a custom attachment on an issue",
		Long: `Create a custom attachment with metadata on an issue.

This operation creates new attachment data and is safe to execute.
Use for custom integrations (CI/CD, monitoring tools, external systems).

Parameters:
  --issue: Issue identifier (e.g., ENG-123) or UUID (required)
  --title: Attachment title (required)
  --url: Attachment URL (required)
  --subtitle: Optional subtitle text
  --icon-url: Optional icon URL (20x20px PNG/JPG, max 1MB)
  --metadata: Optional JSON metadata object

Examples:
  # Create CI/CD build attachment
  linear attachment create --issue=ENG-123 --title="Build #42" --url=https://ci.example.com/build/42

  # With subtitle and icon
  linear attachment create --issue=ENG-123 --title="Deploy" --subtitle="Production" \\
    --url=https://deploy.example.com/123 --icon-url=https://example.com/icon.png

  # With JSON metadata
  linear attachment create --issue=ENG-123 --title="Test Report" --url=https://tests.example.com \\
    --metadata='{"status":"passed","coverage":"95%"}'

TIP: Use for integrating external tools (CI/CD, monitoring, analytics) with Linear issues

Related Commands:
  - linear attachment link-url - Link simple URL without metadata
  - linear attachment link-github - Link GitHub PR
  - linear attachment link-slack - Link Slack thread`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runCreate(cmd, client)
		},
	}

	_ = cmd.MarkFlagRequired("issue")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("url")
	cmd.Flags().String("issue", "", "Issue identifier or UUID (required)")
	cmd.Flags().String("title", "", "Attachment title (required)")
	cmd.Flags().String("url", "", "Attachment URL (required)")
	cmd.Flags().String("subtitle", "", "Subtitle text")
	cmd.Flags().String("icon-url", "", "Icon URL (20x20px PNG/JPG, max 1MB)")
	cmd.Flags().String("metadata", "", "JSON metadata object")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	issueID, _ := cmd.Flags().GetString("issue")
	title, _ := cmd.Flags().GetString("title")
	url, _ := cmd.Flags().GetString("url")

	input := intgraphql.AttachmentCreateInput{
		IssueID: issueID,
		Title:   title,
		URL:     url,
	}

	if subtitle, _ := cmd.Flags().GetString("subtitle"); subtitle != "" {
		input.Subtitle = &subtitle
	}

	if iconURL, _ := cmd.Flags().GetString("icon-url"); iconURL != "" {
		input.IconURL = &iconURL
	}

	if metadata, _ := cmd.Flags().GetString("metadata"); metadata != "" {
		input.Metadata = &metadata
	}

	result, err := client.AttachmentCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create attachment: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Created attachment '%s' on issue %s\n", title, issueID)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
