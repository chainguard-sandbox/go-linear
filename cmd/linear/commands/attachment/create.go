package attachment

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewCreateCommand creates the attachment create command.
func NewCreateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a custom attachment on an issue",
		Long: `Create custom attachment with metadata. Safe operation.

Required: --issue (ID from issue_list), --title, --url
Optional: --subtitle, --icon-url, --metadata (JSON)

Example: go-linear attachment create --issue=ENG-123 --title="Build #42" --url=https://ci.example.com/42 --metadata='{"status":"passed"}'

Related: issue_get, attachment_get`,
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

	return cmd
}

func runCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

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

	if metadataStr, _ := cmd.Flags().GetString("metadata"); metadataStr != "" {
		var metadata map[string]any
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
			return fmt.Errorf("invalid metadata JSON: %w", err)
		}
		input.Metadata = metadata
	}

	result, err := client.AttachmentCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create attachment: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
}
