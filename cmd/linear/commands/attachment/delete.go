package attachment

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
)

func NewDeleteCommand(clientFactory cli.ClientFactory) *cobra.Command {
	confirmFlags := &cli.ConfirmationFlags{}
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an attachment permanently",
		Long: `⚠️ Delete attachment. Cannot be undone. Prompts unless --yes.

Example: go-linear attachment delete <uuid>

Related: attachment_get, issue_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()

			// Confirmation
			if !confirmFlags.Yes {
				fmt.Fprintf(cmd.OutOrStderr(), "⚠️  Delete attachment %s? This cannot be undone.\n", args[0])
				fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				if !strings.EqualFold(strings.TrimSpace(response), "yes") {
					fmt.Fprintln(cmd.OutOrStderr(), "Canceled.")
					return nil
				}
			}

			err = client.AttachmentDelete(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to delete attachment: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "json" {
				return formatter.FormatJSON(cmd.OutOrStdout(), map[string]bool{"success": true}, true)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Attachment deleted\n")
			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	confirmFlags.Bind(cmd)
	return cmd
}
