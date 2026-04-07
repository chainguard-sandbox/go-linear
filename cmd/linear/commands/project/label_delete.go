package project

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewLabelDeleteCommand creates the project label-delete command.
func NewLabelDeleteCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "label-delete <label-id>",
		Short: "Delete a project label",
		Long: `⚠️ Delete a project label. Cannot be undone. Prompts unless --yes.

Example: go-linear project label-delete <uuid>

Related: project_label-list, project_label-create`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runLabelDelete(cmd, client, args[0])
		},
	}

	cmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	return cmd
}

func runLabelDelete(cmd *cobra.Command, client *linear.Client, labelID string) error {
	ctx := cmd.Context()

	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		fmt.Fprintf(cmd.OutOrStderr(), "⚠️  Are you sure you want to delete this project label? This cannot be undone.\n")
		fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)

		if !strings.EqualFold(response, "yes") {
			fmt.Fprintln(cmd.OutOrStderr(), "Canceled")
			return nil
		}
	}

	err := client.ProjectLabelDelete(ctx, labelID)
	if err != nil {
		return fmt.Errorf("failed to delete project label: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Deleted project label\n")
	return nil
}
