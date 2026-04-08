package customview

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

// NewDeleteCommand creates the custom-view delete command.
func NewDeleteCommand(clientFactory cli.ClientFactory) *cobra.Command {
	confirmFlags := &cli.ConfirmationFlags{}

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a custom view",
		Long: `Delete a custom view. Cannot be undone. Prompts unless --yes.

Example: go-linear custom-view delete <uuid>
         go-linear custom-view delete <uuid> --yes

Related: custom-view_list, custom-view_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runDelete(cmd, client, args[0], confirmFlags)
		},
	}

	confirmFlags.Bind(cmd)

	return cmd
}

func runDelete(cmd *cobra.Command, client *linear.Client, id string, confirmFlags *cli.ConfirmationFlags) error {
	ctx := cmd.Context()

	if !confirmFlags.Yes {
		fmt.Fprintf(cmd.OutOrStderr(), "Delete custom view %s? This cannot be undone.\n", id)
		fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if !strings.EqualFold(strings.TrimSpace(response), "yes") {
			fmt.Fprintln(cmd.OutOrStderr(), "Canceled.")
			return nil
		}
	}

	err := client.CustomViewDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete custom view: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]bool{"success": true}, true)
}
