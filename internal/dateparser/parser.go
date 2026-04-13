// Package dateparser provides date parsing utilities for the Linear CLI.
//
// Supports multiple date formats for AI-friendly and human-friendly input:
// - ISO 8601: "2025-12-10", "2025-12-10T15:04:05Z"
// - Named dates: "today", "yesterday", "tomorrow"
// - Duration offsets: "7d", "2w", "3m" (days, weeks, months ago)
package dateparser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var durationRegex = regexp.MustCompile(`^(\d+)([dwm])$`)

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
		var duration time.Duration

		switch unit {
		case "d":
			duration = time.Duration(amount) * 24 * time.Hour
		case "w":
			duration = time.Duration(amount) * 7 * 24 * time.Hour
		case "m":
			// Approximate: 30 days per month
			duration = time.Duration(amount) * 30 * 24 * time.Hour
		default:
			return time.Time{}, fmt.Errorf("invalid duration unit: %s", unit)
		}

		result := now.Add(-duration) // Subtract duration from now
		return time.Date(result.Year(), result.Month(), result.Day(), 0, 0, 0, 0, time.UTC), nil
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s (supported: ISO8601, 'today', 'yesterday', '7d', '2w', '3m')", input)
}

// ParseFuture parses a date string treating durations as future offsets.
//
// Identical to Parse for absolute dates and named dates. For duration formats
// ("7d", "2w", "3m"), the duration is added to now instead of subtracted.
// Use this for snooze/deadline inputs where "3d" means "3 days from now".
func (p Parser) ParseFuture(input string) (time.Time, error) {
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

	// Try duration format (e.g., "7d", "2w", "3m") — future direction
	if matches := durationRegex.FindStringSubmatch(input); matches != nil {
		amount, err := strconv.Atoi(matches[1])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid duration amount: %s", matches[1])
		}

		unit := matches[2]
		var duration time.Duration

		switch unit {
		case "d":
			duration = time.Duration(amount) * 24 * time.Hour
		case "w":
			duration = time.Duration(amount) * 7 * 24 * time.Hour
		case "m":
			duration = time.Duration(amount) * 30 * 24 * time.Hour
		default:
			return time.Time{}, fmt.Errorf("invalid duration unit: %s", unit)
		}

		result := now.Add(duration)
		return time.Date(result.Year(), result.Month(), result.Day(), 0, 0, 0, 0, time.UTC), nil
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s (supported: ISO8601, 'today', 'yesterday', '7d', '2w', '3m')", input)
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
