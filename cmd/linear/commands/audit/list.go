package audit

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/dateparser"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewListCommand creates the audit list command.
func NewListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	paginationFlags := &cli.PaginationFlags{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List audit log entries",
		Long: `List audit log entries with optional filtering. Requires Admin or Owner role.

Filters: --type (entry type), --actor (user name, email, or ID), --ip, --country-code, --created-after, --created-before
Date filters accept ISO8601, relative durations ('7d', '2w'), or 'yesterday'. Comparisons are inclusive.

Example: go-linear audit list --type=issue.create --limit=20
Example: go-linear audit list --actor=jane@example.com --created-after=7d

Related: audit types`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client, paginationFlags)
		},
	}

	cmd.Flags().String("type", "", "Filter by audit entry type (see: audit types)")
	cmd.Flags().String("actor", "", "Filter by actor (name, email, or user ID)")
	cmd.Flags().String("ip", "", "Filter by IP address")
	cmd.Flags().String("country-code", "", "Filter by country code (e.g. US, DE)")
	cmd.Flags().String("created-after", "", "Created after date, inclusive (ISO8601, 'yesterday', '7d')")
	cmd.Flags().String("created-before", "", "Created before date, inclusive (ISO8601, 'yesterday', '7d')")
	paginationFlags.Bind(cmd, 50)

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client, paginationFlags *cli.PaginationFlags) error {
	ctx := cmd.Context()

	first := paginationFlags.LimitPtr()
	var afterPtr *string
	if paginationFlags.After != "" {
		afterPtr = &paginationFlags.After
	}

	// Build filter
	var filter *intgraphql.AuditEntryFilter
	typeFlag, _ := cmd.Flags().GetString("type")
	actorFlag, _ := cmd.Flags().GetString("actor")
	ipFlag, _ := cmd.Flags().GetString("ip")
	countryCodeFlag, _ := cmd.Flags().GetString("country-code")
	createdAfter, _ := cmd.Flags().GetString("created-after")
	createdBefore, _ := cmd.Flags().GetString("created-before")

	if typeFlag != "" || actorFlag != "" || ipFlag != "" || countryCodeFlag != "" || createdAfter != "" || createdBefore != "" {
		filter = &intgraphql.AuditEntryFilter{}

		if typeFlag != "" {
			filter.Type = &intgraphql.StringComparator{Eq: &typeFlag}
		}
		if actorFlag != "" {
			res := resolver.New(client)
			actorID, err := res.ResolveUser(ctx, actorFlag)
			if err != nil {
				return fmt.Errorf("failed to resolve --actor: %w", err)
			}
			filter.Actor = &intgraphql.NullableUserFilter{
				ID: &intgraphql.IDComparator{Eq: &actorID},
			}
		}
		if ipFlag != "" {
			filter.IP = &intgraphql.StringComparator{Eq: &ipFlag}
		}
		if countryCodeFlag != "" {
			filter.CountryCode = &intgraphql.StringComparator{Eq: &countryCodeFlag}
		}

		parser := dateparser.New()
		if createdAfter != "" {
			t, err := parser.Parse(createdAfter)
			if err != nil {
				return fmt.Errorf("invalid --created-after: %w", err)
			}
			if filter.CreatedAt == nil {
				filter.CreatedAt = &intgraphql.DateComparator{}
			}
			s := t.UTC().Format(time.RFC3339)
			filter.CreatedAt.Gte = &s
		}
		if createdBefore != "" {
			t, err := parser.Parse(createdBefore)
			if err != nil {
				return fmt.Errorf("invalid --created-before: %w", err)
			}
			if filter.CreatedAt == nil {
				filter.CreatedAt = &intgraphql.DateComparator{}
			}
			s := t.UTC().Format(time.RFC3339)
			filter.CreatedAt.Lte = &s
		}
		if filter.CreatedAt != nil && filter.CreatedAt.Gte != nil && filter.CreatedAt.Lte != nil {
			if *filter.CreatedAt.Gte > *filter.CreatedAt.Lte {
				return fmt.Errorf("--created-after must be earlier than --created-before")
			}
		}
	}

	entries, err := client.AuditEntries(ctx, first, afterPtr, filter)
	if err != nil {
		return fmt.Errorf("failed to list audit entries: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), entries, true)
}
