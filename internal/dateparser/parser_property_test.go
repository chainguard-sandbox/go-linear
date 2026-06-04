package dateparser

import (
	"fmt"
	"testing"
	"time"
)

// Property: all "Nd" results are in the past relative to when the test started.
func TestParse_RelativeDatesArePast(t *testing.T) {
	p := New()
	before := time.Now().UTC()
	for _, n := range []int{1, 2, 7, 14, 30, 90, 365} {
		input := fmt.Sprintf("%dd", n)
		result, err := p.Parse(input)
		if err != nil {
			t.Fatalf("Parse(%q): %v", input, err)
		}
		if !result.Before(before) {
			t.Errorf("Parse(%q) = %v is not before now (%v)", input, result, before)
		}
	}
}

// mustParse parses input, failing the test if Parse returns an error for an
// input the caller assumes is valid. Swallowing the error would otherwise
// surface as a misleading ordering assertion on a zero-value time.
func mustParse(t *testing.T, p Parser, input string) time.Time {
	t.Helper()
	result, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse(%q): unexpected error: %v", input, err)
	}
	return result
}

// Property: larger N in "Nd" → earlier timestamp (strictly monotonic).
func TestParse_RelativeDay_Monotonic(t *testing.T) {
	p := New()
	pairs := [][2]int{{1, 2}, {2, 7}, {7, 14}, {14, 30}, {30, 90}, {90, 365}}
	for _, pair := range pairs {
		smaller, larger := pair[0], pair[1]
		t1 := mustParse(t, p, fmt.Sprintf("%dd", smaller))
		t2 := mustParse(t, p, fmt.Sprintf("%dd", larger))
		if !t1.After(t2) {
			t.Errorf("%dd (%v) should be after %dd (%v)", smaller, t1, larger, t2)
		}
	}
}

// Property: 1d < 1w < 1m in terms of distance from now
// (smaller offset = more recent = larger timestamp).
func TestParse_UnitOrdering(t *testing.T) {
	p := New()
	d1 := mustParse(t, p, "1d")
	w1 := mustParse(t, p, "1w")
	m1 := mustParse(t, p, "1m")
	if !d1.After(w1) {
		t.Errorf("1d (%v) should be after 1w (%v)", d1, w1)
	}
	if !w1.After(m1) {
		t.Errorf("1w (%v) should be after 1m (%v)", w1, m1)
	}
}

// Property: yesterday < today < tomorrow.
func TestParse_NamedDateOrdering(t *testing.T) {
	p := New()
	yesterday := mustParse(t, p, "yesterday")
	today := mustParse(t, p, "today")
	tomorrow := mustParse(t, p, "tomorrow")
	if !today.After(yesterday) {
		t.Errorf("today (%v) should be after yesterday (%v)", today, yesterday)
	}
	if !tomorrow.After(today) {
		t.Errorf("tomorrow (%v) should be after today (%v)", tomorrow, today)
	}
}

// Property: all relative/named results are UTC with zero time component.
func TestParse_ResultsAreUTCMidnight(t *testing.T) {
	p := New()
	inputs := []string{"today", "yesterday", "tomorrow", "1d", "7d", "2w", "3m"}
	for _, input := range inputs {
		result, err := p.Parse(input)
		if err != nil {
			t.Fatalf("Parse(%q): %v", input, err)
		}
		if result.Location() != time.UTC {
			t.Errorf("Parse(%q) location = %v, want UTC", input, result.Location())
		}
		if h, m, s, ns := result.Hour(), result.Minute(), result.Second(), result.Nanosecond(); h != 0 || m != 0 || s != 0 || ns != 0 {
			t.Errorf("Parse(%q) = %v, want midnight UTC", input, result)
		}
	}
}

// Fuzz: Parse never panics; on success result is always UTC.
func FuzzParse_NoPanic(f *testing.F) {
	f.Add("7d")
	f.Add("today")
	f.Add("2025-12-10")
	f.Add("2025-12-10T15:04:05Z")
	f.Add("")
	f.Add("invalid")
	f.Add("0d")
	f.Add("999m")

	f.Fuzz(func(t *testing.T, s string) {
		p := New()
		result, err := p.Parse(s)
		if err == nil && result.Location() != time.UTC {
			t.Errorf("Parse(%q) succeeded but result is not UTC: %v", s, result)
		}
	})
}
