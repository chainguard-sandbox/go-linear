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
	confirmFlags := &cli.ConfirmationFlags{}

	cmd := &cobra.Command{
		Use:   "label-delete <label-id>",
		Short: "Delete a project label",
		Long: `Delete a project label. Cannot be undone. Prompts unless --yes.

Example: go-linear project label-delete <uuid>

Related: project_label-list, project_label-create`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runLabelDelete(cmd, client, args[0], confirmFlags)
		},
	}

	confirmFlags.Bind(cmd)

	return cmd
}

func runLabelDelete(cmd *cobra.Command, client *linear.Client, labelID string, confirmFlags *cli.ConfirmationFlags) error {
	ctx := cmd.Context()

	if !confirmFlags.Yes {
		fmt.Fprintf(cmd.OutOrStderr(), "Delete project label %s? This cannot be undone.\n", labelID)
		fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if !strings.EqualFold(strings.TrimSpace(response), "yes") {
			fmt.Fprintln(cmd.OutOrStderr(), "Canceled.")
			return nil
		}
	}

	err := client.ProjectLabelDelete(ctx, labelID)
	if err != nil {
		return fmt.Errorf("failed to delete project label: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted project label\n")
	return nil
}
