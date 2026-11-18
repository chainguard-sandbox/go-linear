package issue

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	issuefilter "github.com/chainguard-sandbox/go-linear/v2/internal/filter/issue"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// sloReportPageSize is the page size used when paginating through issues.
const sloReportPageSize = int64(100)

// NewSLOReportCommand creates the issue slo-report command.
func NewSLOReportCommand(clientFactory cli.ClientFactory) *cobra.Command {
	return newSLOReportCommand(clientFactory, time.Now)
}

// newSLOReportCommand takes a clock so tests can pin "now".
func newSLOReportCommand(clientFactory cli.ClientFactory, now func() time.Time) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slo-report",
		Short: "Report time-to-resolution against a label-based policy",
		Long: `Evaluate issues against a label -> target-duration policy and report compliance.
Runs client-side over the normal issues query; Linear's SLA feature is not required.

Policy: --policy "p1=2d,p2=14d,p3=60d" or --policy-file (one label=duration per line, # comments).
Durations: w (7d), d (24h), h, m, s; combinable (1d12h) or fractional (1.5d).
Business days are NOT supported: 1d is 24 calendar hours, weekends included.
Label names match case-insensitively. An issue with several policy labels uses the strictest (shortest) target.
Issues with no policy label are skipped (counted in "unmatched").

Elapsed: completedAt - createdAt, or now - createdAt for open issues.
Status per issue (each issue counts in exactly one):
  met               completed within target
  missed            elapsed past target (completed late, or open and already breached)
  atRisk            open, not breached, elapsed >= --at-risk-threshold % of target
  openWithinTarget  open, below the at-risk threshold
Median/max elapsed cover every evaluated issue in the group, open or completed.

Canceled issues are excluded unless --include-canceled (they have no completedAt and would read as open).
Unless --label is given, only issues carrying a policy label are fetched.

Filters: --team, --assignee, --state, --priority, --label, --project, --cycle, --creator,
  --created-after/before, --completed-after/before, --updated-after/before (same syntax as issue list)
Grouping: --group-by team|assignee|label (label = governing policy label). Omit for one "all" group.
Output: -o json (default) or -o table

Example: go-linear issue slo-report --policy "p1=2d,p2=14d" --team=ENG --created-after=2026-01-01 --group-by=team

Returns: {generatedAt, policy, groupBy, atRiskThreshold, fetched, unmatched, overall, groups: [{key, total, met, missed, atRisk, openWithinTarget, medianElapsed, medianElapsedSeconds, maxElapsed, maxElapsedSeconds}]}
Related: issue_list`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := sloOptionsFromFlags(cmd)
			if err != nil {
				return err
			}

			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runSLOReport(cmd, client, opts, now())
		},
	}

	cmd.Flags().String("policy", "", `Label targets, e.g. "p1=2d,p2=14d"`)
	cmd.Flags().String("policy-file", "", "File with one label=duration per line")
	cmd.MarkFlagsMutuallyExclusive("policy", "policy-file")
	cmd.MarkFlagsOneRequired("policy", "policy-file")
	cmd.Flags().String("group-by", "", "Group results: team, assignee, or label")
	cmd.Flags().Int("at-risk-threshold", 80, "Percent of target (1-100) at which an open issue counts as at risk")
	cmd.Flags().Bool("include-canceled", false, "Include canceled issues (evaluated as open)")
	cmd.Flags().StringP("output", "o", "json", "Output format: json or table")

	// Filters (handled by the shared issue filter builder)
	cmd.Flags().String("team", "", "Team name, key, or ID (e.g., 'ENG')")
	cmd.Flags().String("assignee", "", "Assignee name, email, or ID ('me' for current user)")
	cmd.Flags().String("state", "", "State name or ID")
	cmd.Flags().Int("priority", -1, "Priority: 0=none, 1=urgent, 2=high, 3=normal, 4=low")
	cmd.Flags().StringArray("label", []string{}, "Only issues with these labels (repeatable, OR logic)")
	cmd.Flags().String("project", "", "Project UUID")
	cmd.Flags().String("cycle", "", "Cycle UUID")
	cmd.Flags().String("creator", "", "Creator name, email, or 'me'")
	cmd.Flags().String("created-after", "", "Created after date (ISO8601, 'yesterday', '7d')")
	cmd.Flags().String("created-before", "", "Created before date")
	cmd.Flags().String("completed-after", "", "Completed after date")
	cmd.Flags().String("completed-before", "", "Completed before date")
	cmd.Flags().String("updated-after", "", "Updated after date")
	cmd.Flags().String("updated-before", "", "Updated before date")

	return cmd
}

// sloOptions are the validated non-filter flags.
type sloOptions struct {
	policy          sloPolicy
	groupBy         string
	atRiskThreshold int
	includeCanceled bool
	output          string
}

func sloOptionsFromFlags(cmd *cobra.Command) (sloOptions, error) {
	var opts sloOptions

	policyStr, _ := cmd.Flags().GetString("policy")
	policyFile, _ := cmd.Flags().GetString("policy-file")
	var err error
	if policyFile != "" {
		f, ferr := os.Open(policyFile) // #nosec G304 - path is supplied by the user running the CLI
		if ferr != nil {
			return opts, fmt.Errorf("failed to open policy file: %w", ferr)
		}
		defer f.Close()
		opts.policy, err = parseSLOPolicyFile(f)
	} else {
		opts.policy, err = parseSLOPolicy(policyStr)
	}
	if err != nil {
		return opts, err
	}

	opts.groupBy, _ = cmd.Flags().GetString("group-by")
	if err := validateSLOGroupBy(opts.groupBy); err != nil {
		return opts, err
	}

	opts.atRiskThreshold, _ = cmd.Flags().GetInt("at-risk-threshold")
	if opts.atRiskThreshold < 1 || opts.atRiskThreshold > 100 {
		return opts, fmt.Errorf("invalid --at-risk-threshold %d: must be between 1 and 100", opts.atRiskThreshold)
	}

	opts.includeCanceled, _ = cmd.Flags().GetBool("include-canceled")

	opts.output, _ = cmd.Flags().GetString("output")
	if opts.output != "json" && opts.output != "table" {
		return opts, fmt.Errorf("invalid --output %q: want json or table", opts.output)
	}
	return opts, nil
}

func runSLOReport(cmd *cobra.Command, client *linear.Client, opts sloOptions, now time.Time) error {
	ctx := cmd.Context()

	filterBuilder := issuefilter.NewIssueFilterBuilder(resolver.New(client))
	if err := filterBuilder.FromFlags(ctx, cmd); err != nil {
		return err
	}
	f := filterBuilder.Filter()
	if !opts.includeCanceled {
		null := true
		f.CanceledAt = &intgraphql.NullableDateComparator{Null: &null}
	}
	// Narrow the fetch to issues with a policy label, unless the caller chose
	// labels explicitly. Matching is still re-checked client-side.
	if f.Labels == nil {
		f.Labels = policyLabelFilter(opts.policy)
	}
	issueFilter := filterBuilder.Build()

	issues, err := fetchSLOIssues(cmd, client, issueFilter)
	if err != nil {
		return err
	}

	report := buildSLOReport(issues, opts.policy, opts.groupBy, opts.atRiskThreshold, now)

	if opts.output == "table" {
		return writeSLOTable(cmd.OutOrStdout(), report)
	}
	return formatter.FormatJSON(cmd.OutOrStdout(), report, true)
}

// policyLabelFilter matches issues carrying any policy label (case-insensitive).
func policyLabelFilter(policy sloPolicy) *intgraphql.IssueLabelCollectionFilter {
	ors := make([]*intgraphql.IssueLabelFilter, 0, len(policy))
	for _, name := range policy.labels() {
		n := name
		ors = append(ors, &intgraphql.IssueLabelFilter{
			Name: &intgraphql.StringComparator{EqIgnoreCase: &n},
		})
	}
	return &intgraphql.IssueLabelCollectionFilter{
		Some: &intgraphql.IssueLabelFilter{Or: ors},
	}
}

// fetchSLOIssues pages through every issue matching the filter.
func fetchSLOIssues(cmd *cobra.Command, client *linear.Client, issueFilter *intgraphql.IssueFilter) ([]sloIssue, error) {
	ctx := cmd.Context()
	pageSize := sloReportPageSize
	var after *string
	var out []sloIssue

	for {
		page, err := client.IssuesFiltered(ctx, &pageSize, after, issueFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to list issues: %w", err)
		}
		for _, n := range page.Nodes {
			if n != nil {
				out = append(out, sloIssueFromFiltered(n))
			}
		}
		if !page.PageInfo.HasNextPage || page.PageInfo.EndCursor == nil {
			break
		}
		after = page.PageInfo.EndCursor
	}
	return out, nil
}

func sloIssueFromFiltered(n *intgraphql.ListIssuesFiltered_Issues_Nodes) sloIssue {
	is := sloIssue{
		Identifier:  n.Identifier,
		Team:        n.Team.Key,
		CreatedAt:   n.CreatedAt,
		CompletedAt: n.CompletedAt,
	}
	if is.Team == "" {
		is.Team = n.Team.Name
	}
	if a := n.Assignee; a != nil {
		switch {
		case a.Name != "":
			is.Assignee = a.Name
		case a.DisplayName != "":
			is.Assignee = a.DisplayName
		default:
			is.Assignee = a.Email
		}
	}
	for _, l := range n.Labels.Nodes {
		if l != nil {
			is.Labels = append(is.Labels, l.Name)
		}
	}
	return is
}

// writeSLOTable renders the report as an aligned text table.
func writeSLOTable(w io.Writer, r sloReport) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "GROUP\tTOTAL\tMET\tMISSED\tAT-RISK\tOPEN-OK\tMEDIAN\tMAX")
	row := func(s sloGroupStats) {
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\t%d\t%s\t%s\n",
			s.Key, s.Total, s.Met, s.Missed, s.AtRisk, s.OpenWithinTarget, s.MedianElapsed, s.MaxElapsed)
	}
	for _, g := range r.Groups {
		row(g)
	}
	if r.GroupBy != "none" && len(r.Groups) > 0 {
		overall := r.Overall
		overall.Key = "TOTAL"
		row(overall)
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "\n%d issues evaluated, %d without a policy label skipped. At-risk threshold: %d%%. Group by: %s.\n",
		r.Overall.Total, r.Unmatched, r.AtRiskThreshold, r.GroupBy)
	return err
}
