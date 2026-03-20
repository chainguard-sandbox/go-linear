package initiative

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewListSubCommand creates the initiative list-sub command.
func NewListSubCommand(clientFactory cli.ClientFactory) *cobra.Command {
	paginationFlags := &cli.PaginationFlags{}

	cmd := &cobra.Command{
		Use:   "list-sub <initiative>",
		Short: "List sub-initiatives of an initiative",
		Long: `List sub-initiatives of a parent initiative. Returns id, name, status, health, targetDate.

Required: initiative (name or UUID)

Example: go-linear initiative list-sub "Company OKRs"

Related: initiative_list, initiative_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runListSub(cmd, client, args[0], paginationFlags)
		},
	}

	paginationFlags.Bind(cmd, 50)

	return cmd
}

func runListSub(cmd *cobra.Command, client *linear.Client, initiativeArg string, paginationFlags *cli.PaginationFlags) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	// Resolve initiative ID
	initiativeID, err := res.ResolveInitiative(ctx, initiativeArg)
	if err != nil {
		return fmt.Errorf("failed to resolve initiative: %w", err)
	}

	// Get parent initiative name for display
	parentInit, err := client.Initiative(ctx, initiativeID)
	if err != nil {
		return fmt.Errorf("failed to get initiative: %w", err)
	}

	subInits, err := client.SubInitiatives(ctx, initiativeID, paginationFlags.LimitPtr(), paginationFlags.AfterPtr())
	if err != nil {
		return fmt.Errorf("failed to list sub-initiatives: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
		"parentInitiative": map[string]any{
			"id":   parentInit.ID,
			"name": parentInit.Name,
		},
		"subInitiatives": subInits.Nodes,
		"pageInfo":       subInits.PageInfo,
	}, true)
}
