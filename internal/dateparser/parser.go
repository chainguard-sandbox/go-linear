// Package dateparser provides date parsing utilities for the Linear CLI.
//
// Supports multiple date formats for AI-friendly and human-friendly input:
// - ISO 8601: "2025-12-10", "2025-12-10T15:04:05Z"
// - Named dates: "today", "yesterday", "tomorrow"
// - Duration offsets: "7d", "2w", "3m" (days, weeks, months ago)
package dateparser

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var durationRegex = regexp.MustCompile(`^(\d+)([hdwm])$`)

// relativeDuration converts a parsed amount and unit (h/d/w/m) into a
// time.Duration magnitude. It rejects amounts large enough to overflow int64
// nanoseconds, which would otherwise silently wrap and invert the offset
// direction (e.g. a huge "Nd" making Parse return a future time).
func relativeDuration(amount int, unit string) (time.Duration, error) {
	var unitDur time.Duration
	switch unit {
	case "h":
		unitDur = time.Hour
	case "d":
		unitDur = 24 * time.Hour
	case "w":
		unitDur = 7 * 24 * time.Hour
	case "m":
		unitDur = 30 * 24 * time.Hour // approximate: 30 days per month
	default:
		return 0, fmt.Errorf("invalid duration unit: %s", unit)
	}
	// unitDur is always > 0 here (the default branch returns before this point),
	// so the integer division is safe; dividing MaxInt64 by unitDur first yields
	// the largest amount that won't overflow the subsequent multiplication.
	if time.Duration(amount) > time.Duration(math.MaxInt64)/unitDur {
		return 0, fmt.Errorf("duration amount too large: %d%s", amount, unit)
	}
	return time.Duration(amount) * unitDur, nil
}

// Parser parses date strings in various formats.
type Parser struct{}

// New creates a new date parser.
func New() Parser {
	return Parser{}
}

// Parse parses a date string and returns a time.Time.
//
// Supported formats:
//   - ISO 8601: "2025-12-10", "2025-12-10T15:04:05Z"
//   - RFC3339: "2025-12-10T15:04:05-07:00"
//   - Named: "today", "yesterday", "tomorrow"
//   - Duration: "7d" (7 days ago), "2w" (2 weeks ago), "3m" (3 months ago)
func (p Parser) Parse(input string) (time.Time, error) {
	if input == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	// Try ISO 8601 date only
	if t, err := time.Parse("2006-01-02", input); err == nil {
		return t.UTC(), nil
	}

	// Try ISO 8601 with time
	if t, err := time.Parse(time.RFC3339, input); err == nil {
		return t.UTC(), nil
	}

	// Try named dates
	now := time.Now().UTC()
	switch strings.ToLower(input) {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), nil
	case "yesterday":
		yesterday := now.Add(-24 * time.Hour)
		return time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC), nil
	case "tomorrow":
		tomorrow := now.Add(24 * time.Hour)
		return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, time.UTC), nil
	}

	// Try duration format (e.g., "7d", "2w", "3m")
	if matches := durationRegex.FindStringSubmatch(input); matches != nil {
		amount, err := strconv.Atoi(matches[1])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid duration amount: %s", matches[1])
		}

		unit := matches[2]
		duration, err := relativeDuration(amount, unit)
		if err != nil {
			return time.Time{}, err
		}

		if unit == "h" {
			// Hours preserve wall-clock time; no midnight truncation.
			return now.Add(-duration), nil
		}

		result := now.Add(-duration) // Subtract duration from now
		return time.Date(result.Year(), result.Month(), result.Day(), 0, 0, 0, 0, time.UTC), nil
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s (supported: ISO8601, 'today', 'yesterday', '4h', '7d', '2w', '3m')", input)
}

// ParseFuture parses a date string treating durations as future offsets.
//
// For duration formats ("7d", "2w", "3m"), the duration is added to now.
// Use this for snooze/deadline inputs where "3d" means "3 days from now".
// Unlike Parse, "today", "yesterday", and past ISO 8601 dates are rejected.
func (p Parser) ParseFuture(input string) (time.Time, error) {
	if input == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	// Try ISO 8601 date only
	if t, err := time.Parse("2006-01-02", input); err == nil {
		t = t.UTC()
		if !t.After(time.Now().UTC()) {
			return time.Time{}, fmt.Errorf("date %s is not in the future", input)
		}
		return t, nil
	}

	// Try ISO 8601 with time
	if t, err := time.Parse(time.RFC3339, input); err == nil {
		t = t.UTC()
		if !t.After(time.Now().UTC()) {
			return time.Time{}, fmt.Errorf("date %s is not in the future", input)
		}
		return t, nil
	}

	// Try named dates
	now := time.Now().UTC()
	switch strings.ToLower(input) {
	case "today":
		return time.Time{}, fmt.Errorf("'today' is ambiguous as a future date; use 'tomorrow' or a duration like '1d'")
	case "yesterday":
		return time.Time{}, fmt.Errorf("'yesterday' is not a future date")
	case "tomorrow":
		return now.Add(24 * time.Hour), nil
	}

	// Try duration format (e.g., "7d", "2w", "3m") — future direction
	if matches := durationRegex.FindStringSubmatch(input); matches != nil {
		amount, err := strconv.Atoi(matches[1])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid duration amount: %s", matches[1])
		}

		unit := matches[2]
		duration, err := relativeDuration(amount, unit)
		if err != nil {
			return time.Time{}, err
		}

		return now.Add(duration), nil
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s (supported: ISO8601, 'tomorrow', '4h', '3d', '2w', '3m')", input)
}

// ParseDeadline parses a date string for a deadline that may lie in either
// the past or the future, such as an SLA breach time.
//
// Relative inputs point forward: durations ("4h", "3d", "2w", "3m") are added
// to now, and "tomorrow" means tomorrow, as in ParseFuture. Absolute inputs
// (ISO 8601 dates, RFC3339 timestamps) and "today"/"yesterday" are accepted
// as-is, past or future, as in Parse.
func (p Parser) ParseDeadline(input string) (time.Time, error) {
	if durationRegex.MatchString(input) || strings.EqualFold(input, "tomorrow") {
		return p.ParseFuture(input)
	}
	return p.Parse(input)
}

// MustParse parses a date string and panics on error.
// Useful for testing and initialization.
func (p Parser) MustParse(input string) time.Time {
	t, err := p.Parse(input)
	if err != nil {
		panic(fmt.Sprintf("failed to parse date %q: %v", input, err))
	}
	return t
}
