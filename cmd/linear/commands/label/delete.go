package label

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
)

func NewDeleteCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name|id>",
		Short: "Delete a label permanently",
		Long: `⚠️ Delete label. Removes from all issues. Cannot be undone. Prompts unless --yes.

Example: go-linear label delete bug

Related: label_list, label_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()
			res := resolver.New(client)

			labelID, err := res.ResolveLabel(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve label: %w", err)
			}

			// Confirmation
			yes, _ := cmd.Flags().GetBool("yes")
			if !yes {
				fmt.Fprintf(cmd.OutOrStderr(), "⚠️  Delete label %s? This cannot be undone.\n", args[0])
				fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				if strings.TrimSpace(strings.ToLower(response)) != "yes" {
					fmt.Fprintln(cmd.OutOrStderr(), "Canceled.")
					return nil
				}
			}

			err = client.IssueLabelDelete(ctx, labelID)
			if err != nil {
				return fmt.Errorf("failed to delete label: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "json" {
				return formatter.FormatJSON(cmd.OutOrStdout(), map[string]bool{"success": true}, true)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Label deleted\n")
			return nil
		},
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	return cmd
}
