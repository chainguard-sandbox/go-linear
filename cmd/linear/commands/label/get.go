package label

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the label get command.
func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
	flags := &cli.FieldFlags{}

	cmd := &cobra.Command{
		Use:   "get <label-id>",
		Short: "Get a single label by ID",
		Long: `Get label by UUID. Returns 5 default fields.

Example: go-linear label get <label-uuid>

Related: label_list, label_create`,
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

	flags.Bind(cmd, "defaults (id,name,color,createdAt,description) | none | defaults,extra")

	return cmd
}
func runGet(cmd *cobra.Command, client *linear.Client, labelID string, flags *cli.FieldFlags) error {
	ctx := cmd.Context()

	label, err := client.IssueLabel(ctx, labelID)
	if err != nil {
		return fmt.Errorf("failed to get label: %w", err)
	}

	cfg, _ := config.Load()
	var configOverrides map[string]string
	if cfg != nil {
		configOverrides = cfg.FieldDefaults
	}
	defaults := fieldfilter.GetDefaults("label.get", configOverrides)
	fieldSelector, err := fieldfilter.New(flags.Fields, defaults)
	if err != nil {
		return fmt.Errorf("invalid --fields: %w", err)
	}
	return formatter.FormatJSONFiltered(cmd.OutOrStdout(), label, true, fieldSelector)
}
