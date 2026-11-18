package issue

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

// sloPolicyEntry maps one label to its resolution target.
type sloPolicyEntry struct {
	Label  string
	Target time.Duration
}

// sloPolicy is an ordered set of label targets. Label matching is
// case-insensitive.
type sloPolicy []sloPolicyEntry

// durationUnits are the units accepted by parseSLODuration. Days and weeks are
// calendar days (24h), not business days.
var durationUnits = map[byte]time.Duration{
	'w': 7 * 24 * time.Hour,
	'd': 24 * time.Hour,
	'h': time.Hour,
	'm': time.Minute,
	's': time.Second,
}

// parseSLODuration parses durations such as "2d", "1w", "36h", "1d12h" or
// "1.5d". Units: w (7d), d (24h), h, m, s. The result must be positive.
func parseSLODuration(s string) (time.Duration, error) {
	in := strings.TrimSpace(s)
	if in == "" {
		return 0, fmt.Errorf("empty duration")
	}

	var total time.Duration
	rest := in
	for rest != "" {
		i := 0
		for i < len(rest) && (rest[i] >= '0' && rest[i] <= '9' || rest[i] == '.') {
			i++
		}
		if i == 0 || i == len(rest) {
			return 0, fmt.Errorf("invalid duration %q (want e.g. 2d, 36h, 1w, 1d12h)", in)
		}
		unit, ok := durationUnits[rest[i]]
		if !ok {
			return 0, fmt.Errorf("invalid duration %q: unknown unit %q (use w, d, h, m, s)", in, rest[i])
		}
		n, err := strconv.ParseFloat(rest[:i], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q: %w", in, err)
		}
		total += time.Duration(n * float64(unit))
		rest = rest[i+1:]
	}

	if total <= 0 {
		return 0, fmt.Errorf("invalid duration %q: must be greater than zero", in)
	}
	return total, nil
}

// formatSLODuration renders a duration as days/hours/minutes, e.g. "3d4h12m".
// Seconds are dropped unless the duration is under a minute; zero renders as
// "0m".
func formatSLODuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d > 0 && d < time.Minute {
		return fmt.Sprintf("%ds", d/time.Second)
	}
	totalMinutes := int64(d / time.Minute)
	days := totalMinutes / (24 * 60)
	hours := totalMinutes / 60 % 24
	minutes := totalMinutes % 60

	var b strings.Builder
	if days > 0 {
		fmt.Fprintf(&b, "%dd", days)
	}
	if hours > 0 {
		fmt.Fprintf(&b, "%dh", hours)
	}
	if minutes > 0 || b.Len() == 0 {
		fmt.Fprintf(&b, "%dm", minutes)
	}
	return b.String()
}

// parseSLOPolicy parses a comma-separated "label=duration" list, as given to
// --policy.
func parseSLOPolicy(s string) (sloPolicy, error) {
	return buildSLOPolicy(strings.Split(s, ","))
}

// parseSLOPolicyFile parses a policy file: one "label=duration" entry per
// line. Blank lines and lines starting with '#' are ignored. Because entries
// are split on newlines only, label names may contain commas.
func parseSLOPolicyFile(r io.Reader) (sloPolicy, error) {
	var entries []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entries = append(entries, line)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("failed to read policy file: %w", err)
	}
	return buildSLOPolicy(entries)
}

func buildSLOPolicy(entries []string) (sloPolicy, error) {
	var policy sloPolicy
	seen := map[string]bool{}
	for _, raw := range entries {
		entry := strings.TrimSpace(raw)
		if entry == "" {
			continue
		}
		// Split on the last '=' so label names may contain '='.
		rawLabel, rawTarget, ok := strings.CutLast(entry, "=")
		if !ok {
			return nil, fmt.Errorf("invalid policy entry %q: want label=duration", entry)
		}
		label := strings.TrimSpace(rawLabel)
		if label == "" {
			return nil, fmt.Errorf("invalid policy entry %q: empty label", entry)
		}
		target, err := parseSLODuration(rawTarget)
		if err != nil {
			return nil, fmt.Errorf("invalid policy entry %q: %w", entry, err)
		}
		key := strings.ToLower(label)
		if seen[key] {
			return nil, fmt.Errorf("duplicate policy label %q", label)
		}
		seen[key] = true
		policy = append(policy, sloPolicyEntry{Label: label, Target: target})
	}
	if len(policy) == 0 {
		return nil, fmt.Errorf("policy is empty")
	}
	return policy, nil
}

// labels returns the policy label names in policy order.
func (p sloPolicy) labels() []string {
	out := make([]string, len(p))
	for i, e := range p {
		out[i] = e.Label
	}
	return out
}

// match returns the strictest (shortest-target) policy entry among the given
// issue labels. Ties go to the entry listed first in the policy.
func (p sloPolicy) match(issueLabels []string) (sloPolicyEntry, bool) {
	have := make(map[string]bool, len(issueLabels))
	for _, l := range issueLabels {
		have[strings.ToLower(l)] = true
	}
	var best sloPolicyEntry
	found := false
	for _, e := range p {
		if !have[strings.ToLower(e.Label)] {
			continue
		}
		if !found || e.Target < best.Target {
			best = e
			found = true
		}
	}
	return best, found
}

// sloIssue is the subset of issue data the report needs.
type sloIssue struct {
	Identifier  string
	Team        string
	Assignee    string // empty when unassigned
	Labels      []string
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// sloStatus is the outcome of evaluating one issue.
type sloStatus string

const (
	sloMet              sloStatus = "met"
	sloMissed           sloStatus = "missed"
	sloAtRisk           sloStatus = "atRisk"
	sloOpenWithinTarget sloStatus = "openWithinTarget"
)

// sloEvaluation is one issue evaluated against the policy.
type sloEvaluation struct {
	Issue   sloIssue
	Policy  sloPolicyEntry
	Elapsed time.Duration
	Status  sloStatus
}

// evaluateSLOIssue evaluates an issue against a single policy entry.
// Elapsed is completedAt-createdAt, or now-createdAt for open issues.
// atRiskFraction is in (0,1]; an open issue whose elapsed time is at or above
// that fraction of the target (but not past it) is at risk.
func evaluateSLOIssue(issue sloIssue, entry sloPolicyEntry, now time.Time, atRiskFraction float64) sloEvaluation {
	end := now
	if issue.CompletedAt != nil {
		end = *issue.CompletedAt
	}
	elapsed := max(end.Sub(issue.CreatedAt), 0)

	ev := sloEvaluation{Issue: issue, Policy: entry, Elapsed: elapsed}
	switch {
	case elapsed > entry.Target:
		ev.Status = sloMissed
	case issue.CompletedAt != nil:
		ev.Status = sloMet
	case float64(elapsed) >= atRiskFraction*float64(entry.Target):
		ev.Status = sloAtRisk
	default:
		ev.Status = sloOpenWithinTarget
	}
	return ev
}

// sloGroupBy values accepted by --group-by.
const (
	sloGroupNone     = ""
	sloGroupTeam     = "team"
	sloGroupAssignee = "assignee"
	sloGroupLabel    = "label"
)

const (
	sloAllGroupKey    = "all"
	sloUnassignedKey  = "unassigned"
	sloNoTeamGroupKey = "unknown"
)

func validateSLOGroupBy(g string) error {
	switch g {
	case sloGroupNone, sloGroupTeam, sloGroupAssignee, sloGroupLabel:
		return nil
	}
	return fmt.Errorf("invalid --group-by %q: want team, assignee, or label", g)
}

// sloGroupKey returns the group an evaluation belongs to. Label grouping uses
// the governing (strictest) policy label, so each issue lands in exactly one
// group.
func sloGroupKey(ev sloEvaluation, groupBy string) string {
	switch groupBy {
	case sloGroupTeam:
		if ev.Issue.Team == "" {
			return sloNoTeamGroupKey
		}
		return ev.Issue.Team
	case sloGroupAssignee:
		if ev.Issue.Assignee == "" {
			return sloUnassignedKey
		}
		return ev.Issue.Assignee
	case sloGroupLabel:
		return ev.Policy.Label
	}
	return sloAllGroupKey
}

// sloGroupStats are per-group counts. met+missed+atRisk+openWithinTarget ==
// total.
type sloGroupStats struct {
	Key                  string `json:"key"`
	Total                int    `json:"total"`
	Met                  int    `json:"met"`
	Missed               int    `json:"missed"`
	AtRisk               int    `json:"atRisk"`
	OpenWithinTarget     int    `json:"openWithinTarget"`
	MedianElapsed        string `json:"medianElapsed"`
	MedianElapsedSeconds int64  `json:"medianElapsedSeconds"`
	MaxElapsed           string `json:"maxElapsed"`
	MaxElapsedSeconds    int64  `json:"maxElapsedSeconds"`
}

// summarizeSLO computes stats over a set of evaluations.
func summarizeSLO(key string, evs []sloEvaluation) sloGroupStats {
	s := sloGroupStats{Key: key, Total: len(evs)}
	elapsed := make([]time.Duration, 0, len(evs))
	for _, ev := range evs {
		switch ev.Status {
		case sloMet:
			s.Met++
		case sloMissed:
			s.Missed++
		case sloAtRisk:
			s.AtRisk++
		case sloOpenWithinTarget:
			s.OpenWithinTarget++
		}
		elapsed = append(elapsed, ev.Elapsed)
	}
	med := medianDuration(elapsed)
	var mx time.Duration
	if len(elapsed) > 0 {
		mx = slices.Max(elapsed)
	}
	s.MedianElapsed = formatSLODuration(med)
	s.MedianElapsedSeconds = int64(med / time.Second)
	s.MaxElapsed = formatSLODuration(mx)
	s.MaxElapsedSeconds = int64(mx / time.Second)
	return s
}

// groupSLO partitions evaluations by group key and summarizes each group.
// Groups are sorted by key.
func groupSLO(evs []sloEvaluation, groupBy string) []sloGroupStats {
	byKey := map[string][]sloEvaluation{}
	for _, ev := range evs {
		k := sloGroupKey(ev, groupBy)
		byKey[k] = append(byKey[k], ev)
	}
	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]sloGroupStats, 0, len(keys))
	for _, k := range keys {
		out = append(out, summarizeSLO(k, byKey[k]))
	}
	return out
}

// medianDuration returns the median; for an even count it is the mean of the
// two middle values. Empty input returns 0.
func medianDuration(ds []time.Duration) time.Duration {
	if len(ds) == 0 {
		return 0
	}
	// make+copy rather than slices.Clone: nilaway treats Clone's result as
	// possibly nil even though ds is non-empty here.
	sorted := make([]time.Duration, len(ds))
	copy(sorted, ds)
	slices.Sort(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[mid]
	}
	return sorted[mid-1] + (sorted[mid]-sorted[mid-1])/2
}

// sloPolicyJSON is the policy as echoed in the report.
type sloPolicyJSON struct {
	Label         string `json:"label"`
	Target        string `json:"target"`
	TargetSeconds int64  `json:"targetSeconds"`
}

// sloReport is the JSON output of issue slo-report.
type sloReport struct {
	GeneratedAt     time.Time       `json:"generatedAt"`
	Policy          []sloPolicyJSON `json:"policy"`
	GroupBy         string          `json:"groupBy"`
	AtRiskThreshold int             `json:"atRiskThreshold"`
	Fetched         int             `json:"fetched"`
	Unmatched       int             `json:"unmatched"`
	Overall         sloGroupStats   `json:"overall"`
	Groups          []sloGroupStats `json:"groups"`
}

// buildSLOReport evaluates issues against the policy and builds the report.
// Issues with no policy label are counted in Unmatched and otherwise ignored.
// atRiskThreshold is a percentage in [1,100].
func buildSLOReport(issues []sloIssue, policy sloPolicy, groupBy string, atRiskThreshold int, now time.Time) sloReport {
	fraction := float64(atRiskThreshold) / 100
	evs := make([]sloEvaluation, 0, len(issues))
	unmatched := 0
	for _, is := range issues {
		entry, ok := policy.match(is.Labels)
		if !ok {
			unmatched++
			continue
		}
		evs = append(evs, evaluateSLOIssue(is, entry, now, fraction))
	}

	pj := make([]sloPolicyJSON, len(policy))
	for i, e := range policy {
		pj[i] = sloPolicyJSON{
			Label:         e.Label,
			Target:        formatSLODuration(e.Target),
			TargetSeconds: int64(math.Round(e.Target.Seconds())),
		}
	}

	gb := groupBy
	if gb == sloGroupNone {
		gb = "none"
	}

	return sloReport{
		GeneratedAt:     now.UTC(),
		Policy:          pj,
		GroupBy:         gb,
		AtRiskThreshold: atRiskThreshold,
		Fetched:         len(issues),
		Unmatched:       unmatched,
		Overall:         summarizeSLO(sloAllGroupKey, evs),
		Groups:          groupSLO(evs, groupBy),
	}
}
