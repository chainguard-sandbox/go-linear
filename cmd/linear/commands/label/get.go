package label

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the label get command.
func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <label-id>",
		Short: "Get a single label by ID",
		Long: `Get label by UUID. Returns 5 default fields.

Example: go-linear label get <label-uuid> --output=json

Related: label_list, label_create`,
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
	cmd.Flags().String("fields", "", "defaults (id,name,color,createdAt,description) | none | defaults,extra")

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, labelID string) error {
	ctx := context.Background()

	label, err := client.IssueLabel(ctx, labelID)
	if err != nil {
		return fmt.Errorf("failed to get label: %w", err)
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
		defaults := fieldfilter.GetDefaults("label.get", configOverrides)
		fieldSelector, err := fieldfilter.New(fieldsSpec, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), label, true, fieldSelector)
	case "table":
		// Simple table output for single label
		fmt.Fprintf(cmd.OutOrStdout(), "ID:          %s\n", label.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", label.Name)
		if label.Description != nil && *label.Description != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", *label.Description)
		}
		if label.Color != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Color:       %s\n", label.Color)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created:     %s\n", label.CreatedAt)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
