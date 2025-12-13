package attachment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the attachment get command.
func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <attachment-id>",
		Short: "Get a single attachment by ID",
		Long: `Get attachment by UUID. Returns 5 default fields.

Example: go-linear attachment get <uuid> --output=json

Related: issue_get, attachment_delete`,
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
	cmd.Flags().String("fields", "", "defaults (id,title,url,source,createdAt) | none | defaults,extra")

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, attachmentID string) error {
	ctx := context.Background()

	attachment, err := client.Attachment(ctx, attachmentID)
	if err != nil {
		return fmt.Errorf("failed to get attachment: %w", err)
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
		defaults := fieldfilter.GetDefaults("attachment.get", configOverrides)
		fieldSelector, err := fieldfilter.New(fieldsSpec, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), attachment, true, fieldSelector)
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
