package initiative

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
)

func NewDeleteCommand(clientFactory cli.ClientFactory) *cobra.Command {
	confirmFlags := &cli.ConfirmationFlags{}
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an initiative permanently",
		Long: `Delete initiative. Cannot be undone. Prompts unless --yes.

Example: go-linear initiative delete <uuid>

Related: initiative_list, initiative_get`,
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
				fmt.Fprintf(cmd.OutOrStderr(), "Delete initiative %s? This cannot be undone.\n", args[0])
				fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				if !strings.EqualFold(strings.TrimSpace(response), "yes") {
					fmt.Fprintln(cmd.OutOrStderr(), "Canceled.")
					return nil
				}
			}

			err = client.InitiativeDelete(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to delete initiative: %w", err)
			}

			return formatter.FormatJSON(cmd.OutOrStdout(), map[string]bool{"success": true}, true)
		},
	}

	confirmFlags.Bind(cmd)
	return cmd
}
