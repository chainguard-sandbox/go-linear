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

// NewStatusUpdateCreateCommand creates the initiative status-update-create command.
func NewStatusUpdateCreateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status-update-create",
		Short: "Create a status update for an initiative",
		Long: `Create initiative status update. Safe operation.

Required: --initiative (UUID or name), --body
Optional: --health (onTrack, atRisk, offTrack)

Example: go-linear initiative status-update-create --initiative=<uuid> --body="On track for Q1 release" --health=onTrack

Related: initiative_status-update-list, initiative_status-update-get, initiative_status-update-archive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runStatusUpdateCreate(cmd, client)
		},
	}

	cmd.Flags().String("initiative", "", "Initiative name or UUID (required)")
	_ = cmd.MarkFlagRequired("initiative")
	cmd.Flags().String("body", "", "Status update body (markdown) (required)")
	_ = cmd.MarkFlagRequired("body")
	cmd.Flags().String("health", "", "Initiative health: onTrack, atRisk, offTrack")

	return cmd
}

func runStatusUpdateCreate(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	// Resolve initiative
	initiativeInput, _ := cmd.Flags().GetString("initiative")
	initiativeID, err := res.ResolveInitiative(ctx, initiativeInput)
	if err != nil {
		return fmt.Errorf("failed to resolve initiative: %w", err)
	}

	body, _ := cmd.Flags().GetString("body")
	input := intgraphql.InitiativeUpdateCreateInput{
		InitiativeID: initiativeID,
		Body:         &body,
	}

	if health, _ := cmd.Flags().GetString("health"); health != "" {
		healthType := intgraphql.InitiativeUpdateHealthType(health)
		input.Health = &healthType
	}

	result, err := client.InitiativeUpdateCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create initiative status update: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
}
