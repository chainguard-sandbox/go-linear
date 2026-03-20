package cycle

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

func NewUpdateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing cycle",
		Long: `Update cycle. Modifies existing data.

Fields: --name, --description

Example: go-linear cycle update <uuid> --name="Sprint 43" --description="Extended"

Related: cycle_get, cycle_list`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()
			input := intgraphql.CycleUpdateInput{}
			updated := false

			if name, _ := cmd.Flags().GetString("name"); name != "" {
				input.Name = &name
				updated = true
			}

			if desc, _ := cmd.Flags().GetString("description"); desc != "" {
				input.Description = &desc
				updated = true
			}

			if !updated {
				return fmt.Errorf("no fields to update specified")
			}

			result, err := client.CycleUpdate(ctx, args[0], input)
			if err != nil {
				return fmt.Errorf("failed to update cycle: %w", err)
			}

			return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
		},
	}

	cmd.Flags().String("name", "", "New cycle name")
	cmd.Flags().String("description", "", "New description")

	return cmd
}
