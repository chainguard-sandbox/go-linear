package initiative

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUpdateCommand creates the initiative update command.
func NewUpdateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing initiative",
		Long: `Update initiative. Modifies existing data.

Fields: --name, --description, --parent, --target-date, --owner, --status

Example: go-linear initiative update <uuid> --name="Updated Security Policy" --status=Completed --output=json

Related: initiative_get, initiative_create`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runUpdate(cmd, client, args[0])
		},
	}

	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("description", "", "New description (markdown)")
	cmd.Flags().String("target-date", "", "New target completion date (ISO8601 format: YYYY-MM-DD)")
	cmd.Flags().String("owner", "", "New owner name, email, or 'me'")
	cmd.Flags().String("status", "", "New status: planned, active, completed")

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runUpdate(cmd *cobra.Command, client *linear.Client, initiativeID string) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	input := intgraphql.InitiativeUpdateInput{}
	updated := false

	if name, _ := cmd.Flags().GetString("name"); name != "" {
		input.Name = &name
		updated = true
	}

	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		input.Description = &desc
		updated = true
	}

	// Resolve owner
	if owner, _ := cmd.Flags().GetString("owner"); owner != "" {
		ownerID, err := res.ResolveUser(ctx, owner)
		if err != nil {
			return fmt.Errorf("failed to resolve owner: %w", err)
		}
		input.OwnerID = &ownerID
		updated = true
	}

	// Set target date (ISO8601 format: YYYY-MM-DD)
	if targetDate, _ := cmd.Flags().GetString("target-date"); targetDate != "" {
		input.TargetDate = &targetDate
		updated = true
	}

	// Set status
	if status, _ := cmd.Flags().GetString("status"); status != "" {
		statusType := intgraphql.InitiativeStatus(status)
		input.Status = &statusType
		updated = true
	}

	if !updated {
		return fmt.Errorf("no fields to update specified")
	}

	result, err := client.InitiativeUpdate(ctx, initiativeID, input)
	if err != nil {
		return fmt.Errorf("failed to update initiative: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Updated initiative: %s\n", result.Name)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
