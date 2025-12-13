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

// NewListCommand creates the state list command.
func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all workflow states",
		Long: `List workflow states. Returns 5 default fields per state. Use for discovering state names.

States define issue lifecycle: triage, backlog, unstarted, started, completed, canceled.

Example: go-linear state list --output=json

Related: state_get, issue_update, issue_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client)
		},
	}

	cmd.Flags().IntP("limit", "l", 100, "Number of states to return")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	cmd.Flags().String("fields", "", "defaults (id,name,type,color,position) | none | defaults,extra")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	states, err := client.WorkflowStates(ctx, &first, nil)
	if err != nil {
		return fmt.Errorf("failed to list workflow states: %w", err)
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
		defaults := fieldfilter.GetDefaults("state.list", configOverrides)
		fieldSelector, err := fieldfilter.NewForList(fieldsSpec, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), states, true, fieldSelector)
	case "table":
		if len(states.Nodes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No workflow states found")
			return nil
		}
		for _, state := range states.Nodes {
			fmt.Fprintf(cmd.OutOrStdout(), "%s (%s)\n", state.Name, state.Type)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
