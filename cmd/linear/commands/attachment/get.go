package attachment

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the attachment get command.
func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
	flags := &cli.FieldFlags{}

	cmd := &cobra.Command{
		Use:   "get <attachment-id>",
		Short: "Get a single attachment by ID",
		Long: `Get attachment by UUID. Returns 5 default fields.

Example: go-linear attachment get <uuid>

Related: issue_get, attachment_delete`,
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

	flags.Bind(cmd, "defaults (id,title,url,source,createdAt) | none | defaults,extra")

	return cmd
}
func runGet(cmd *cobra.Command, client *linear.Client, attachmentID string, flags *cli.FieldFlags) error {
	ctx := cmd.Context()

	attachment, err := client.Attachment(ctx, attachmentID)
	if err != nil {
		return fmt.Errorf("failed to get attachment: %w", err)
	}

	cfg, _ := config.Load()
	var configOverrides map[string]string
	if cfg != nil {
		configOverrides = cfg.FieldDefaults
	}
	defaults := fieldfilter.GetDefaults("attachment.get", configOverrides)
	fieldSelector, err := fieldfilter.New(flags.Fields, defaults)
	if err != nil {
		return fmt.Errorf("invalid --fields: %w", err)
	}
	return formatter.FormatJSONFiltered(cmd.OutOrStdout(), attachment, true, fieldSelector)
}
