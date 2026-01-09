package initiative

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewStatusUpdateGetCommand creates the initiative status-update-get command.
func NewStatusUpdateGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
	outputFlags := &cli.OutputFlags{}

	cmd := &cobra.Command{
		Use:   "status-update-get <id>",
		Short: "Get an initiative status update by ID",
		Long: `Get initiative status update by UUID. Returns 5 default fields.

Example: go-linear initiative status-update-get <uuid> --output=json

Related: initiative_status-update-list, initiative_status-update-create`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runStatusUpdateGet(cmd, client, args[0], outputFlags)
		},
	}

	outputFlags.Bind(cmd, "defaults (id,body,health,createdAt,url) | none | defaults,extra")

	return cmd
}

func runStatusUpdateGet(cmd *cobra.Command, client *linear.Client, updateID string, outputFlags *cli.OutputFlags) error {
	ctx := cmd.Context()

	if err := outputFlags.Validate(); err != nil {
		return err
	}

	update, err := client.GetInitiativeUpdate(ctx, updateID)
	if err != nil {
		return fmt.Errorf("failed to get initiative status update: %w", err)
	}

	switch outputFlags.Output {
	case "json":
		cfg, _ := config.Load()
		var configOverrides map[string]string
		if cfg != nil {
			configOverrides = cfg.FieldDefaults
		}
		defaults := fieldfilter.GetDefaults("initiative.status-update-get", configOverrides)
		fieldSelector, err := fieldfilter.New(outputFlags.Fields, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), update, true, fieldSelector)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "ID: %s\n", update.ID)
		if update.Health != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Health: %s\n", update.Health)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Body: %s\n", update.Body)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", outputFlags.Output)
	}
}
