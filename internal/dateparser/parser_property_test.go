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

// mustParseFuture is the ParseFuture analog of mustParse.
func mustParseFuture(t *testing.T, p Parser, input string) time.Time {
	t.Helper()
	result, err := p.ParseFuture(input)
	if err != nil {
		t.Fatalf("ParseFuture(%q): unexpected error: %v", input, err)
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

// Property: "Nh" preserves wall-clock time — unlike d/w/m it is not truncated to midnight.
// Verified by comparing Parse("24h") (no truncation) with Parse("1d") (midnight truncation):
// they must differ unless the test runs exactly at midnight UTC.
func TestParse_HoursNotTruncated(t *testing.T) {
	p := New()
	before := time.Now().UTC()
	if before.Equal(before.Truncate(24 * time.Hour)) {
		t.Skip("running at midnight UTC; hour/day truncation indistinguishable")
	}
	day := mustParse(t, p, "1d")
	if day.Hour() != 0 || day.Minute() != 0 || day.Second() != 0 {
		t.Fatalf("1d result %v is not midnight UTC — assumption broken", day)
	}
	hour, err := p.Parse("24h")
	if err != nil {
		t.Fatalf("Parse(\"24h\"): %v", err)
	}
	after := time.Now().UTC()
	// Skip if a UTC date boundary was crossed during the two Parse calls: when
	// Parse("24h")'s internal time.Now() lands at midnight, the result equals the
	// midnight-truncated 1d result through coincidence, not a truncation bug.
	if !before.Truncate(24 * time.Hour).Equal(after.Truncate(24 * time.Hour)) {
		t.Skip("execution straddled midnight UTC; result indeterminate")
	}
	if hour.Equal(day) {
		t.Errorf("Parse(\"24h\") == Parse(\"1d\") (%v); hours should not be truncated to midnight", hour)
	}
}

// Property: all future results are after the time the test started.
// N=0 is intentionally excluded from Nd inputs: ParseFuture("0d") returns now+0=now,
// which is not strictly after the pre-call timestamp.
func TestParseFuture_RelativeDatesAreFuture(t *testing.T) {
	p := New()
	for _, n := range []int{1, 2, 7, 14, 30, 90, 365} {
		input := fmt.Sprintf("%dd", n)
		before := time.Now().UTC()
		result, err := p.ParseFuture(input)
		if err != nil {
			t.Fatalf("ParseFuture(%q): %v", input, err)
		}
		if !result.After(before) {
			t.Errorf("ParseFuture(%q) = %v is not after pre-call timestamp (%v)", input, result, before)
		}
	}
	for _, input := range []string{"1w", "2w", "1m", "3m"} {
		before := time.Now().UTC()
		result, err := p.ParseFuture(input)
		if err != nil {
			t.Fatalf("ParseFuture(%q): %v", input, err)
		}
		if !result.After(before) {
			t.Errorf("ParseFuture(%q) = %v is not after pre-call timestamp (%v)", input, result, before)
		}
	}
}

// Property: larger N → later timestamp (monotonic, opposite direction from Parse).
// Covers d/w/m units and cross-unit ordering (2w < 1m since 14d < 30d).
func TestParseFuture_RelativeDay_Monotonic(t *testing.T) {
	p := New()
	pairs := [][2]int{{1, 2}, {2, 7}, {7, 14}, {14, 30}, {30, 90}, {90, 365}}
	for _, pair := range pairs {
		smaller, larger := pair[0], pair[1]
		t1 := mustParseFuture(t, p, fmt.Sprintf("%dd", smaller))
		t2 := mustParseFuture(t, p, fmt.Sprintf("%dd", larger))
		if !t2.After(t1) {
			t.Errorf("%dd (%v) should be before %dd (%v)", smaller, t1, larger, t2)
		}
	}
	wPairs := [][2]string{{"1w", "2w"}, {"1m", "2m"}, {"2m", "3m"}, {"2w", "1m"}}
	for _, pair := range wPairs {
		t1 := mustParseFuture(t, p, pair[0])
		t2 := mustParseFuture(t, p, pair[1])
		if !t2.After(t1) {
			t.Errorf("%s (%v) should be before %s (%v)", pair[0], t1, pair[1], t2)
		}
	}
}

// Property: ParseFuture rejects "yesterday" and "today"; accepts "tomorrow".
func TestParseFuture_NamedDateRejections(t *testing.T) {
	p := New()
	for _, rejected := range []string{"yesterday", "today"} {
		if _, err := p.ParseFuture(rejected); err == nil {
			t.Errorf("ParseFuture(%q) expected error, got nil", rejected)
		}
	}
	result, err := p.ParseFuture("tomorrow")
	if err != nil {
		t.Fatalf("ParseFuture(\"tomorrow\"): unexpected error: %v", err)
	}
	if result.Location() != time.UTC {
		t.Errorf("ParseFuture(\"tomorrow\") location = %v, want UTC", result.Location())
	}
}

// Property: ParseFuture does not truncate any unit to midnight — all preserve wall-clock time.
// This contrasts with Parse where only "h" is exempt from midnight truncation.
// Each input is a t.Run subtest so that a midnight skip on one input does not
// silently discard the remaining inputs.
func TestParseFuture_NoMidnightTruncation(t *testing.T) {
	p := New()
	for _, input := range []string{"1h", "1d", "2w", "3m", "tomorrow"} {
		t.Run(input, func(t *testing.T) {
			before := time.Now().UTC()
			if before.Equal(before.Truncate(24 * time.Hour)) {
				t.Skip("running at midnight UTC; truncation indistinguishable from no truncation")
			}
			result, err := p.ParseFuture(input)
			if err != nil {
				t.Fatalf("ParseFuture(%q): %v", input, err)
			}
			after := time.Now().UTC()
			// Skip if a UTC date boundary was crossed during this ParseFuture call:
			// when the internal time.Now() lands at midnight, the result is correctly
			// midnight, indistinguishable from a truncation bug.
			if !before.Truncate(24 * time.Hour).Equal(after.Truncate(24 * time.Hour)) {
				t.Skip("execution straddled midnight UTC; result indeterminate")
			}
			// Use Truncate rather than H/M/S: 00:00:00.5 has H=M=S=0 but is not midnight.
			if result.Equal(result.Truncate(24 * time.Hour)) {
				t.Errorf("ParseFuture(%q) = %v: expected wall-clock time, not midnight", input, result)
			}
		})
	}
}

// Property: ParseFuture rejects past RFC3339 datetimes (parser.go line ~137).
func TestParseFuture_PastRFC3339Rejected(t *testing.T) {
	p := New()
	if _, err := p.ParseFuture("2000-01-01T00:00:00Z"); err == nil {
		t.Error("ParseFuture(\"2000-01-01T00:00:00Z\") expected error for past datetime, got nil")
	}
}

// Property: ParseFuture rejects a past date-only ISO 8601 string (parser.go line ~128).
func TestParseFuture_PastISODateRejected(t *testing.T) {
	p := New()
	if _, err := p.ParseFuture("2000-01-01"); err == nil {
		t.Error("ParseFuture(\"2000-01-01\") expected error for past date, got nil")
	}
}

// Property: ParseFuture rejects empty string input.
func TestParseFuture_EmptyStringRejected(t *testing.T) {
	p := New()
	if _, err := p.ParseFuture(""); err == nil {
		t.Error("ParseFuture(\"\") expected error for empty input, got nil")
	}
}

// Property: ParseFuture with a date-only ISO 8601 string returns midnight UTC —
// the same truncation behavior as Parse for absolute dates.
func TestParseFuture_AbsoluteISODate(t *testing.T) {
	p := New()
	result, err := p.ParseFuture("2099-01-01")
	if err != nil {
		t.Fatalf("ParseFuture(\"2099-01-01\"): %v", err)
	}
	if result.Location() != time.UTC {
		t.Errorf("location = %v, want UTC", result.Location())
	}
	if !result.Equal(result.Truncate(24 * time.Hour)) {
		t.Errorf("ParseFuture(\"2099-01-01\") = %v, want midnight UTC", result)
	}
}

// Property: ParseFuture accepts a future RFC3339 datetime and returns the correct UTC value.
// Covers the time.Parse(time.RFC3339, input) acceptance path (parser.go lines 135–141).
func TestParseFuture_FutureRFC3339Accepted(t *testing.T) {
	p := New()
	result, err := p.ParseFuture("2099-01-01T15:04:05Z")
	if err != nil {
		t.Fatalf("ParseFuture(\"2099-01-01T15:04:05Z\"): %v", err)
	}
	if result.Location() != time.UTC {
		t.Errorf("location = %v, want UTC", result.Location())
	}
	if result.Year() != 2099 || result.Month() != 1 || result.Day() != 1 ||
		result.Hour() != 15 || result.Minute() != 4 || result.Second() != 5 {
		t.Errorf("ParseFuture(\"2099-01-01T15:04:05Z\") = %v, want 2099-01-01T15:04:05Z", result)
	}
}

// Fuzz: ParseFuture never panics; on success result is always UTC.
func FuzzParseFuture_NoPanic(f *testing.F) {
	f.Add("7d")
	f.Add("tomorrow")
	f.Add("1h")
	f.Add("2099-01-01")
	f.Add("2099-01-01T15:04:05Z")
	f.Add("")
	f.Add("invalid")
	f.Add("0d")
	f.Add("yesterday")            // rejected: not a future date
	f.Add("2000-01-01T00:00:00Z") // rejected: past RFC3339 datetime

	f.Fuzz(func(t *testing.T, s string) {
		// Relies on time.Now() being non-decreasing within a single iteration.
		// An NTP backward step could cause result.Before(before) on a correct implementation,
		// but this is not a realistic concern in CI.
		before := time.Now().UTC()
		p := New()
		result, err := p.ParseFuture(s)
		if err != nil {
			return
		}
		if result.Location() != time.UTC {
			t.Errorf("ParseFuture(%q) succeeded but result is not UTC: %v", s, result)
		}
		// N=0 (e.g. "0d") returns now+0 which equals ParseFuture's internal now >= before.
		if result.Before(before) {
			t.Errorf("ParseFuture(%q) = %v is before test start %v", s, result, before)
		}
	})
}

// Property: Parse("0d") == today midnight; Parse("0h") is not truncated to midnight;
// ParseFuture("0d") is not before the pre-call timestamp.
// Each section is a t.Run subtest so that a midnight skip in one section does not
// silently suppress the other sections.
func TestParse_ZeroAmounts(t *testing.T) {
	p := New()

	t.Run("0d equals today", func(t *testing.T) {
		// Both calls take their own time.Now(); bracket to skip if a UTC date boundary
		// is crossed between them.
		before := time.Now().UTC()
		if before.Equal(before.Truncate(24 * time.Hour)) {
			t.Skip("running at midnight UTC; date may shift between Parse calls")
		}
		today := mustParse(t, p, "today")
		zeroDay := mustParse(t, p, "0d")
		after := time.Now().UTC()
		if !before.Truncate(24 * time.Hour).Equal(after.Truncate(24 * time.Hour)) {
			t.Skip("execution straddled midnight UTC; Parse(\"today\") and Parse(\"0d\") may differ by one day")
		}
		if !zeroDay.Equal(today) {
			t.Errorf("Parse(\"0d\") = %v, want today midnight (%v)", zeroDay, today)
		}
	})

	t.Run("0h not truncated", func(t *testing.T) {
		before := time.Now().UTC()
		if before.Equal(before.Truncate(24 * time.Hour)) {
			t.Skip("running at midnight UTC; Parse(\"0h\") == midnight is ambiguous")
		}
		zeroHour := mustParse(t, p, "0h")
		after := time.Now().UTC()
		if !before.Truncate(24 * time.Hour).Equal(after.Truncate(24 * time.Hour)) {
			t.Skip("execution straddled midnight UTC; Parse(\"0h\") result indeterminate")
		}
		if zeroHour.Equal(zeroHour.Truncate(24 * time.Hour)) {
			t.Errorf("Parse(\"0h\") = %v: should preserve wall-clock time, not truncate to midnight", zeroHour)
		}
	})

	t.Run("ParseFuture 0d", func(t *testing.T) {
		// ParseFuture("0d") adds zero duration to now; result >= the pre-call timestamp.
		before := time.Now().UTC()
		result, err := p.ParseFuture("0d")
		if err != nil {
			t.Fatalf("ParseFuture(\"0d\"): %v", err)
		}
		if result.Before(before) {
			t.Errorf("ParseFuture(\"0d\") = %v is before pre-call timestamp %v", result, before)
		}
	})
}

// Fuzz: Parse never panics; on success result is always UTC.
func FuzzParse_NoPanic(f *testing.F) {
	f.Add("7d")
	f.Add("1h")
	f.Add("today")
	f.Add("2025-12-10")
	f.Add("2025-12-10T15:04:05Z")
	f.Add("")
	f.Add("invalid")
	f.Add("0d")
	f.Add("999m")
	f.Add("9223372036855m") // triggers relativeDuration overflow guard for 'm'
	f.Add("106751992d")     // triggers relativeDuration overflow guard for 'd'

	f.Fuzz(func(t *testing.T, s string) {
		p := New()
		result, err := p.Parse(s)
		if err == nil && result.Location() != time.UTC {
			t.Errorf("Parse(%q) succeeded but result is not UTC: %v", s, result)
		}
	})
}
