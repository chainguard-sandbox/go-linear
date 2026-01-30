package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewStatusListCommand creates the project status-list command.
func NewStatusListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status-list",
		Short: "List project statuses",
		Long: `List organization's project statuses. Returns name and type per status.

Example: go-linear project status-list --output=json

Related: project_update, project_get`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runStatusList(cmd, client)
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runStatusList(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	statuses, err := client.ProjectStatuses(ctx)
	if err != nil {
		return fmt.Errorf("failed to list project statuses: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), statuses, true)
	case "table":
		if len(statuses) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No project statuses found")
			return nil
		}
		for _, s := range statuses {
			fmt.Fprintf(cmd.OutOrStdout(), "%s (%s)\n", s.Name, s.Type)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
