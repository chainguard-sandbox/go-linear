package cycle

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

func NewUpdateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing cycle",
		Long: `Update cycle. Modifies existing data.

Fields: --name, --description

Example: go-linear cycle update <uuid> --name="Sprint 43" --description="Extended" --output=json

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

			output, _ := cmd.Flags().GetString("output")
			if output == "json" {
				return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
			}
			name := "cycle"
			if result.Name != nil {
				name = *result.Name
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Updated cycle: %s\n", name)
			return nil
		},
	}

	cmd.Flags().String("name", "", "New cycle name")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}
