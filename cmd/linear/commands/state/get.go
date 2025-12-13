package state

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the workflow state get command.
func NewGetCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <state-id>",
		Short: "Get a single workflow state by ID",
		Long: `Get workflow state by UUID. Returns 5 default fields.

Example: go-linear state get <state-uuid> --output=json

Related: state_list, issue_update`,
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
	cmd.Flags().String("fields", "", "defaults (id,name,type,color,position) | none | defaults,extra")

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, stateID string) error {
	ctx := context.Background()

	state, err := client.WorkflowState(ctx, stateID)
	if err != nil {
		return fmt.Errorf("failed to get workflow state: %w", err)
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
		defaults := fieldfilter.GetDefaults("state.get", configOverrides)
		fieldSelector, err := fieldfilter.New(fieldsSpec, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), state, true, fieldSelector)
	case "table":
		// Simple table output for single state
		fmt.Fprintf(cmd.OutOrStdout(), "ID:          %s\n", state.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", state.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "Type:        %s\n", state.Type)
		if state.Color != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Color:       %s\n", state.Color)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Position:    %.0f\n", state.Position)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
