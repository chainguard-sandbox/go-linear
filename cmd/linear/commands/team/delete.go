package team

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
)

func NewDeleteCommand(clientFactory cli.ClientFactory) *cobra.Command {
	confirmFlags := &cli.ConfirmationFlags{}
	outputFlags := &cli.OutputOnlyFlags{}
	cmd := &cobra.Command{
		Use:   "delete <name|id>",
		Short: "Delete a team permanently",
		Long: `⚠️ Delete team. Cannot be undone. Prompts unless --yes.

Example: go-linear team delete TestTeam

Related: team_list, team_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()

			if err := outputFlags.Validate(); err != nil {
				return err
			}

			res := resolver.New(client)

			teamID, err := res.ResolveTeam(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve team: %w", err)
			}

			// Confirmation
			if !confirmFlags.Yes {
				fmt.Fprintf(cmd.OutOrStderr(), "🚨 Delete team %s? This CANNOT be undone.\n", args[0])
				fmt.Fprint(cmd.OutOrStderr(), "Type 'yes' to confirm: ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				if !strings.EqualFold(strings.TrimSpace(response), "yes") {
					fmt.Fprintln(cmd.OutOrStderr(), "Canceled.")
					return nil
				}
			}

			err = client.TeamDelete(ctx, teamID)
			if err != nil {
				return fmt.Errorf("failed to delete team: %w", err)
			}

			if outputFlags.Output == "json" {
				return formatter.FormatJSON(cmd.OutOrStdout(), map[string]bool{"success": true}, true)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Team deleted\n")
			return nil
		},
	}

	outputFlags.Bind(cmd)
	confirmFlags.Bind(cmd)
	return cmd
}
