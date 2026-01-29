package project

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewStatusUpdateDeleteCommand creates the project status-update-delete command.
func NewStatusUpdateDeleteCommand(clientFactory cli.ClientFactory) *cobra.Command {
	confirmFlags := &cli.ConfirmationFlags{}
	outputFlags := &cli.OutputOnlyFlags{}

	cmd := &cobra.Command{
		Use:   "status-update-delete <id>",
		Short: "Delete a project status update",
		Long: `⚠️ Delete project status update. Cannot be undone. Prompts unless --yes.

Example: go-linear project status-update-delete <uuid>

Related: project_status-update-list, project_status-update-get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runStatusUpdateDelete(cmd, client, args[0], confirmFlags, outputFlags)
		},
	}

	outputFlags.Bind(cmd)
	confirmFlags.Bind(cmd)

	return cmd
}

func runStatusUpdateDelete(cmd *cobra.Command, client *linear.Client, updateID string, confirmFlags *cli.ConfirmationFlags, outputFlags *cli.OutputOnlyFlags) error {
	ctx := cmd.Context()

	if err := outputFlags.Validate(); err != nil {
		return err
	}

	// Confirmation
	if !confirmFlags.Yes {
		fmt.Fprintf(cmd.OutOrStderr(), "⚠️  Delete project status update %s? This cannot be undone.\n", updateID)
		fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if !strings.EqualFold(strings.TrimSpace(response), "yes") {
			fmt.Fprintln(cmd.OutOrStderr(), "Canceled.")
			return nil
		}
	}

	err := client.ProjectUpdateDelete(ctx, updateID)
	if err != nil {
		return fmt.Errorf("failed to delete project status update: %w", err)
	}

	switch outputFlags.Output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), map[string]bool{"success": true}, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Project status update deleted\n")
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", outputFlags.Output)
	}
}
