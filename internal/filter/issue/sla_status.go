package issue

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplySLAStatus handles --sla-status flag.
func ApplySLAStatus(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	slaStatus, _ := cmd.Flags().GetString("sla-status")
	if slaStatus == "" {
		return nil
	}

	b.Filter().SLAStatus = &intgraphql.SLAStatusComparator{
		Eq: (*intgraphql.SLAStatus)(&slaStatus),
	}
	return nil
}
