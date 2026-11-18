package issue

import (
	"strings"
	"testing"
	"time"
)

func TestParseSLODuration(t *testing.T) {
	tests := []struct {
		in      string
		want    time.Duration
		wantErr bool
	}{
		{in: "2d", want: 48 * time.Hour},
		{in: "14d", want: 14 * 24 * time.Hour},
		{in: "1w", want: 7 * 24 * time.Hour},
		{in: "36h", want: 36 * time.Hour},
		{in: "90m", want: 90 * time.Minute},
		{in: "30s", want: 30 * time.Second},
		{in: "1d12h", want: 36 * time.Hour},
		{in: "1w2d", want: 9 * 24 * time.Hour},
		{in: "1.5d", want: 36 * time.Hour},
		{in: " 2d ", want: 48 * time.Hour},
		{in: "", wantErr: true},
		{in: "2", wantErr: true},
		{in: "d", wantErr: true},
		{in: "2y", wantErr: true},
		{in: "2D", wantErr: true},
		{in: "0d", wantErr: true},
		{in: "-1d", wantErr: true},
		{in: "1..5d", wantErr: true},
		{in: "2d3", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := parseSLODuration(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseSLODuration(%q) = %v, want error", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseSLODuration(%q) error = %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("parseSLODuration(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestFormatSLODuration(t *testing.T) {
	tests := []struct {
		in   time.Duration
		want string
	}{
		{0, "0m"},
		{-time.Hour, "0m"},
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m"},
		{2 * time.Hour, "2h"},
		{48 * time.Hour, "2d"},
		{3*24*time.Hour + 4*time.Hour + 12*time.Minute, "3d4h12m"},
		{24*time.Hour + 5*time.Minute, "1d5m"},
	}
	for _, tt := range tests {
		if got := formatSLODuration(tt.in); got != tt.want {
			t.Errorf("formatSLODuration(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParseSLOPolicy(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    sloPolicy
		wantErr string
	}{
		{
			name: "basic",
			in:   "p1=2d,p2=14d,p3=60d",
			want: sloPolicy{{"p1", 48 * time.Hour}, {"p2", 14 * 24 * time.Hour}, {"p3", 60 * 24 * time.Hour}},
		},
		{
			name: "whitespace and trailing comma",
			in:   " p1 = 2d , Security Bug=1w,",
			want: sloPolicy{{"p1", 48 * time.Hour}, {"Security Bug", 7 * 24 * time.Hour}},
		},
		{name: "empty", in: "", wantErr: "policy is empty"},
		{name: "missing equals", in: "p1", wantErr: "want label=duration"},
		{name: "empty label", in: "=2d", wantErr: "empty label"},
		{name: "bad duration", in: "p1=soon", wantErr: "invalid policy entry"},
		{name: "duplicate case-insensitive", in: "p1=2d,P1=3d", wantErr: "duplicate policy label"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSLOPolicy(tt.in)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("parseSLOPolicy(%q) error = %v, want containing %q", tt.in, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseSLOPolicy(%q) error = %v", tt.in, err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d entries, want %d: %+v", len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("entry %d = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseSLOPolicyFile(t *testing.T) {
	in := `# support tiers
p1=2d

Tier 1, Enterprise=4h
key=with=equals=1d
`
	got, err := parseSLOPolicyFile(strings.NewReader(in))
	if err != nil {
		t.Fatalf("parseSLOPolicyFile error = %v", err)
	}
	want := sloPolicy{
		{"p1", 48 * time.Hour},
		{"Tier 1, Enterprise", 4 * time.Hour},
		{"key=with=equals", 24 * time.Hour},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("entry %d = %+v, want %+v", i, got[i], want[i])
		}
	}

	if _, err := parseSLOPolicyFile(strings.NewReader("# only comments\n")); err == nil {
		t.Error("expected error for empty policy file")
	}
}

func TestSLOPolicyMatch(t *testing.T) {
	policy := sloPolicy{{"p2", 14 * 24 * time.Hour}, {"p1", 48 * time.Hour}, {"sev", 48 * time.Hour}}
	tests := []struct {
		name   string
		labels []string
		want   string
		found  bool
	}{
		{name: "no labels", labels: nil},
		{name: "no policy labels", labels: []string{"bug"}},
		{name: "single", labels: []string{"bug", "p2"}, want: "p2", found: true},
		{name: "case-insensitive", labels: []string{"P2"}, want: "p2", found: true},
		{name: "strictest wins", labels: []string{"p2", "p1"}, want: "p1", found: true},
		{name: "tie goes to policy order", labels: []string{"sev", "p1"}, want: "p1", found: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := policy.match(tt.labels)
			if ok != tt.found {
				t.Fatalf("found = %v, want %v", ok, tt.found)
			}
			if ok && got.Label != tt.want {
				t.Errorf("label = %q, want %q", got.Label, tt.want)
			}
		})
	}
}

func TestEvaluateSLOIssue(t *testing.T) {
	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	entry := sloPolicyEntry{Label: "p1", Target: 10 * 24 * time.Hour}
	at := func(daysBeforeNow float64) time.Time {
		return now.Add(-time.Duration(daysBeforeNow * float64(24*time.Hour)))
	}
	ptr := func(t time.Time) *time.Time { return &t }

	tests := []struct {
		name        string
		created     time.Time
		completed   *time.Time
		wantStatus  sloStatus
		wantElapsed time.Duration
	}{
		{name: "completed within", created: at(20), completed: ptr(at(15)), wantStatus: sloMet, wantElapsed: 5 * 24 * time.Hour},
		{name: "completed exactly at target", created: at(20), completed: ptr(at(10)), wantStatus: sloMet, wantElapsed: 10 * 24 * time.Hour},
		{name: "completed late", created: at(20), completed: ptr(at(5)), wantStatus: sloMissed, wantElapsed: 15 * 24 * time.Hour},
		{name: "open breached", created: at(11), wantStatus: sloMissed, wantElapsed: 11 * 24 * time.Hour},
		{name: "open at risk boundary", created: at(8), wantStatus: sloAtRisk, wantElapsed: 8 * 24 * time.Hour},
		{name: "open at risk", created: at(9), wantStatus: sloAtRisk, wantElapsed: 9 * 24 * time.Hour},
		{name: "open within", created: at(2), wantStatus: sloOpenWithinTarget, wantElapsed: 2 * 24 * time.Hour},
		{name: "created in future clamps to zero", created: now.Add(time.Hour), wantStatus: sloOpenWithinTarget, wantElapsed: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := evaluateSLOIssue(sloIssue{CreatedAt: tt.created, CompletedAt: tt.completed}, entry, now, 0.8)
			if ev.Status != tt.wantStatus {
				t.Errorf("status = %s, want %s", ev.Status, tt.wantStatus)
			}
			if ev.Elapsed != tt.wantElapsed {
				t.Errorf("elapsed = %v, want %v", ev.Elapsed, tt.wantElapsed)
			}
		})
	}
}

func TestMedianDuration(t *testing.T) {
	h := time.Hour
	tests := []struct {
		name string
		in   []time.Duration
		want time.Duration
	}{
		{name: "empty", in: nil, want: 0},
		{name: "one", in: []time.Duration{5 * h}, want: 5 * h},
		{name: "odd unsorted", in: []time.Duration{9 * h, 1 * h, 5 * h}, want: 5 * h},
		{name: "even", in: []time.Duration{4 * h, 1 * h, 2 * h, 3 * h}, want: 150 * time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := append([]time.Duration(nil), tt.in...)
			if got := medianDuration(tt.in); got != tt.want {
				t.Errorf("median = %v, want %v", got, tt.want)
			}
			for i := range orig {
				if orig[i] != tt.in[i] {
					t.Fatal("medianDuration mutated its input")
				}
			}
		})
	}
}

func TestBuildSLOReport(t *testing.T) {
	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	day := 24 * time.Hour
	ago := func(d time.Duration) time.Time { return now.Add(-d) }
	ptr := func(t time.Time) *time.Time { return &t }

	policy := sloPolicy{{"p1", 2 * day}, {"p2", 14 * day}}
	issues := []sloIssue{
		// ENG
		{Identifier: "ENG-1", Team: "ENG", Assignee: "alice", Labels: []string{"p1"}, CreatedAt: ago(10 * day), CompletedAt: ptr(ago(9 * day))},       // met, 1d
		{Identifier: "ENG-2", Team: "ENG", Assignee: "alice", Labels: []string{"p1", "p2"}, CreatedAt: ago(10 * day), CompletedAt: ptr(ago(5 * day))}, // p1 governs: missed, 5d
		{Identifier: "ENG-3", Team: "ENG", Labels: []string{"p2"}, CreatedAt: ago(12 * day)},                                                          // open at risk, 12d
		// OPS
		{Identifier: "OPS-1", Team: "OPS", Assignee: "bob", Labels: []string{"P2"}, CreatedAt: ago(1 * day)},  // open within, 1d
		{Identifier: "OPS-2", Team: "OPS", Assignee: "bob", Labels: []string{"bug"}, CreatedAt: ago(1 * day)}, // unmatched
	}

	t.Run("group by team", func(t *testing.T) {
		r := buildSLOReport(issues, policy, sloGroupTeam, 80, now)
		if r.Fetched != 5 || r.Unmatched != 1 {
			t.Errorf("fetched=%d unmatched=%d, want 5 and 1", r.Fetched, r.Unmatched)
		}
		if r.GroupBy != "team" || r.AtRiskThreshold != 80 {
			t.Errorf("groupBy=%q threshold=%d", r.GroupBy, r.AtRiskThreshold)
		}
		if len(r.Groups) != 2 {
			t.Fatalf("got %d groups, want 2: %+v", len(r.Groups), r.Groups)
		}
		eng, ops := r.Groups[0], r.Groups[1]
		wantEng := sloGroupStats{
			Key: "ENG", Total: 3, Met: 1, Missed: 1, AtRisk: 1,
			MedianElapsed: "5d", MedianElapsedSeconds: int64(5 * day / time.Second),
			MaxElapsed: "12d", MaxElapsedSeconds: int64(12 * day / time.Second),
		}
		if eng != wantEng {
			t.Errorf("ENG = %+v\nwant  %+v", eng, wantEng)
		}
		if ops.Key != "OPS" || ops.Total != 1 || ops.OpenWithinTarget != 1 || ops.MedianElapsed != "1d" {
			t.Errorf("OPS = %+v", ops)
		}
		o := r.Overall
		if o.Key != "all" || o.Total != 4 || o.Met != 1 || o.Missed != 1 || o.AtRisk != 1 || o.OpenWithinTarget != 1 {
			t.Errorf("overall = %+v", o)
		}
		// Elapsed: 1d, 5d, 12d, 1d -> sorted 1,1,5,12 -> median 3d.
		if o.MedianElapsed != "3d" || o.MaxElapsed != "12d" {
			t.Errorf("overall median/max = %s/%s, want 3d/12d", o.MedianElapsed, o.MaxElapsed)
		}
		if len(r.Policy) != 2 || r.Policy[0].Target != "2d" || r.Policy[0].TargetSeconds != 172800 {
			t.Errorf("policy = %+v", r.Policy)
		}
	})

	t.Run("group by assignee", func(t *testing.T) {
		r := buildSLOReport(issues, policy, sloGroupAssignee, 80, now)
		keys := groupKeys(r.Groups)
		if strings.Join(keys, ",") != "alice,bob,unassigned" {
			t.Errorf("keys = %v", keys)
		}
	})

	t.Run("group by label uses governing label", func(t *testing.T) {
		r := buildSLOReport(issues, policy, sloGroupLabel, 80, now)
		if got := strings.Join(groupKeys(r.Groups), ","); got != "p1,p2" {
			t.Fatalf("keys = %s", got)
		}
		if r.Groups[0].Total != 2 || r.Groups[1].Total != 2 {
			t.Errorf("p1 total=%d p2 total=%d, want 2 and 2", r.Groups[0].Total, r.Groups[1].Total)
		}
	})

	t.Run("no grouping", func(t *testing.T) {
		r := buildSLOReport(issues, policy, sloGroupNone, 80, now)
		if r.GroupBy != "none" || len(r.Groups) != 1 || r.Groups[0].Key != "all" || r.Groups[0].Total != 4 {
			t.Errorf("groupBy=%q groups=%+v", r.GroupBy, r.Groups)
		}
	})

	t.Run("threshold changes at-risk", func(t *testing.T) {
		// ENG-3 is at 12/14 = 86%; at a 90% threshold it is within target.
		r := buildSLOReport(issues, policy, sloGroupNone, 90, now)
		if r.Overall.AtRisk != 0 || r.Overall.OpenWithinTarget != 2 {
			t.Errorf("overall = %+v", r.Overall)
		}
	})

	t.Run("empty", func(t *testing.T) {
		r := buildSLOReport(nil, policy, sloGroupTeam, 80, now)
		if r.Overall.Total != 0 || len(r.Groups) != 0 || r.Overall.MedianElapsed != "0m" {
			t.Errorf("report = %+v", r)
		}
		if r.Groups == nil {
			t.Error("groups should be an empty slice, not nil (JSON [])")
		}
	})
}

func TestValidateSLOGroupBy(t *testing.T) {
	for _, ok := range []string{"", "team", "assignee", "label"} {
		if err := validateSLOGroupBy(ok); err != nil {
			t.Errorf("validateSLOGroupBy(%q) = %v", ok, err)
		}
	}
	if err := validateSLOGroupBy("project"); err == nil {
		t.Error("expected error for project")
	}
}

func groupKeys(gs []sloGroupStats) []string {
	out := make([]string, len(gs))
	for i, g := range gs {
		out[i] = g.Key
	}
	return out
}
