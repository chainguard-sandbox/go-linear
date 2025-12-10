package team

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
		Short: "Delete a team permanently",
		Long: `Delete a team from Linear permanently.

🚨 DESTRUCTIVE OPERATION - CANNOT BE UNDONE 🚨

This PERMANENTLY removes the team and may affect issues assigned to it.
Confirmation prompt appears unless --yes flag is used.

Examples:
  linear team delete TestTeam
  linear team delete <uuid> --yes

RECOMMENDATION: Ensure team has no active issues before deleting.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()
			res := resolver.New(client)

			teamID, err := res.ResolveTeam(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve team: %w", err)
			}

			// Confirmation
			yes, _ := cmd.Flags().GetBool("yes")
			if !yes {
				fmt.Fprintf(cmd.OutOrStderr(), "🚨 Delete team %s? This CANNOT be undone.\n", args[0])
				fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				if strings.TrimSpace(strings.ToLower(response)) != "yes" {
					fmt.Fprintln(cmd.OutOrStderr(), "Canceled.")
					return nil
				}
			}

			err = client.TeamDelete(ctx, teamID)
			if err != nil {
				return fmt.Errorf("failed to delete team: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "json" {
				return formatter.FormatJSON(cmd.OutOrStdout(), map[string]bool{"success": true}, true)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Team deleted\n")
			return nil
		},
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
	return cmd
}
