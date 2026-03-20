package initiative

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewStatusUpdateArchiveCommand creates the initiative status-update-archive command.
func NewStatusUpdateArchiveCommand(clientFactory cli.ClientFactory) *cobra.Command {
	confirmFlags := &cli.ConfirmationFlags{}

	cmd := &cobra.Command{
		Use:   "status-update-archive <id>",
		Short: "Archive an initiative status update",
		Long: `Archive initiative status update. Cannot be undone. Prompts unless --yes.

Example: go-linear initiative status-update-archive <uuid>

Related: initiative_status-update-list, initiative_status-update-get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runStatusUpdateArchive(cmd, client, args[0], confirmFlags)
		},
	}

	confirmFlags.Bind(cmd)

	return cmd
}

func runStatusUpdateArchive(cmd *cobra.Command, client *linear.Client, updateID string, confirmFlags *cli.ConfirmationFlags) error {
	ctx := cmd.Context()

	// Confirmation
	if !confirmFlags.Yes {
		fmt.Fprintf(cmd.OutOrStderr(), "Archive initiative status update %s? This cannot be undone.\n", updateID)
		fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if !strings.EqualFold(strings.TrimSpace(response), "yes") {
			fmt.Fprintln(cmd.OutOrStderr(), "Canceled.")
			return nil
		}
	}

	err := client.InitiativeUpdateArchive(ctx, updateID)
	if err != nil {
		return fmt.Errorf("failed to archive initiative status update: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]bool{"success": true}, true)
}
