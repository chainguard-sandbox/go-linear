package issue

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/dateparser"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// slaFlagNames names the three SLA setter flags of a command. issue update
// uses --sla-*, issue batch-update uses --set-sla-* (its --set-* convention
// keeps update flags apart from filter flags).
type slaFlagNames struct {
	breachesAt string
	startedAt  string
	slaType    string
}

var (
	updateSLAFlags = slaFlagNames{
		breachesAt: "sla-breaches-at",
		startedAt:  "sla-started-at",
		slaType:    "sla-type",
	}
	batchUpdateSLAFlags = slaFlagNames{
		breachesAt: "set-sla-breaches-at",
		startedAt:  "set-sla-started-at",
		slaType:    "set-sla-type",
	}
)

// bind registers the SLA setter flags on cmd.
func (n slaFlagNames) bind(cmd *cobra.Command) {
	cmd.Flags().String(n.breachesAt, "", "SLA breach time (ISO8601, or relative from now: '24h', '3d', '1w'; empty or 'none' removes the SLA; requires SLAs enabled)")
	cmd.Flags().String(n.startedAt, "", "SLA start time (ISO8601, 'today', or relative ago: '2d'; empty or 'none' clears)")
	cmd.Flags().String(n.slaType, "", "SLA day count: all (calendar days) or onlyBusinessDays (empty or 'none' clears)")
}

// changed reports whether any SLA flag was passed, including with an empty
// value (an explicit clear).
func (n slaFlagNames) changed(cmd *cobra.Command) bool {
	return cmd.Flags().Changed(n.breachesAt) || cmd.Flags().Changed(n.startedAt) || cmd.Flags().Changed(n.slaType)
}

// isSLAClear reports whether a flag value means "clear this field".
func isSLAClear(v string) bool {
	return v == "" || v == "none"
}

// applySLAFlags parses the SLA flags that were passed and sets the matching
// fields on input. An empty value or 'none' sends an explicit null.
//
// Breach times are deadlines, so relative values point forward ("3d" = three
// days from now). Start times point back ("2d" = two days ago), like other
// date flags. Absolute dates are accepted as given.
func applySLAFlags(cmd *cobra.Command, names slaFlagNames, input *linear.IssueUpdateNullableInput) error {
	parser := dateparser.New()

	if cmd.Flags().Changed(names.breachesAt) {
		v, _ := cmd.Flags().GetString(names.breachesAt)
		if isSLAClear(v) {
			input.SLABreachesAt = linear.NewNull[time.Time]()
		} else {
			t, err := parser.ParseDeadline(v)
			if err != nil {
				return fmt.Errorf("invalid --%s: %w", names.breachesAt, err)
			}
			input.SLABreachesAt = linear.NewValue(t)
		}
	}

	if cmd.Flags().Changed(names.startedAt) {
		v, _ := cmd.Flags().GetString(names.startedAt)
		if isSLAClear(v) {
			input.SLAStartedAt = linear.NewNull[time.Time]()
		} else {
			t, err := parser.Parse(v)
			if err != nil {
				return fmt.Errorf("invalid --%s: %w", names.startedAt, err)
			}
			input.SLAStartedAt = linear.NewValue(t)
		}
	}

	if cmd.Flags().Changed(names.slaType) {
		v, _ := cmd.Flags().GetString(names.slaType)
		if isSLAClear(v) {
			input.SLAType = linear.NewNull[string]()
		} else {
			if !intgraphql.SLADayCountType(v).IsValid() {
				return fmt.Errorf("invalid --%s %q: must be 'all' or 'onlyBusinessDays'", names.slaType, v)
			}
			input.SLAType = linear.NewValue(v)
		}
	}

	return nil
}

// slaErrorPattern matches API error text that refers to SLAs: the acronym
// itself or one of the SLA input field names. It is case-sensitive so that
// words like "Slack" do not match.
var slaErrorPattern = regexp.MustCompile(`SLA|sla(BreachesAt|StartedAt|Type)`)

// slaHint is appended to API errors that mention SLAs.
const slaHint = "hint: native SLA fields need Linear's SLA feature (Business and Enterprise plans), enabled by a workspace admin in Settings > Issues > SLAs"

// withSLAHint adds slaHint to err when the error text mentions SLAs. Other
// errors are returned unchanged, so the hint only shows when the API itself
// points at SLAs.
func withSLAHint(err error) error {
	if err == nil || !slaErrorPattern.MatchString(err.Error()) {
		return err
	}
	return fmt.Errorf("%w\n%s", err, slaHint)
}

// parseIssueSort parses --sort. Supported: "sla-status", "sla-status:asc" and
// "sla-status:desc". Issues without an SLA sort last. Empty returns nil
// (server default order).
func parseIssueSort(spec string) ([]*intgraphql.IssueSortInput, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, nil
	}
	field, dir, _ := strings.Cut(spec, ":")
	if field != "sla-status" {
		return nil, fmt.Errorf("invalid --sort %q: supported: sla-status[:asc|:desc]", spec)
	}
	order := intgraphql.PaginationSortOrderAscending
	switch dir {
	case "", "asc":
	case "desc":
		order = intgraphql.PaginationSortOrderDescending
	default:
		return nil, fmt.Errorf("invalid --sort direction %q: want asc or desc", dir)
	}
	nulls := intgraphql.PaginationNullsLast
	return []*intgraphql.IssueSortInput{{
		SLAStatus: &intgraphql.SLAStatusSort{Order: &order, Nulls: &nulls},
	}}, nil
}
