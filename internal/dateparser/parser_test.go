package dateparser

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	p := New()
	now := time.Now().UTC()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, result time.Time)
	}{
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "ISO 8601 date",
			input:   "2025-12-10",
			wantErr: false,
			check: func(t *testing.T, result time.Time) {
				if result.Year() != 2025 || result.Month() != 12 || result.Day() != 10 {
					t.Errorf("Parse() = %v, want 2025-12-10", result)
				}
			},
		},
		{
			name:    "RFC3339",
			input:   "2025-12-10T15:04:05Z",
			wantErr: false,
			check: func(t *testing.T, result time.Time) {
				if result.Year() != 2025 || result.Month() != 12 || result.Day() != 10 {
					t.Errorf("Parse() date = %v, want 2025-12-10", result)
				}
				if result.Hour() != 15 || result.Minute() != 4 || result.Second() != 5 {
					t.Errorf("Parse() time = %v, want 15:04:05", result)
				}
			},
		},
		{
			name:    "today",
			input:   "today",
			wantErr: false,
			check: func(t *testing.T, result time.Time) {
				if result.Year() != now.Year() || result.Month() != now.Month() || result.Day() != now.Day() {
					t.Errorf("Parse('today') = %v, want today", result)
				}
				if result.Hour() != 0 || result.Minute() != 0 || result.Second() != 0 {
					t.Errorf("Parse('today') time = %v, want 00:00:00", result)
				}
			},
		},
		{
			name:    "yesterday",
			input:   "yesterday",
			wantErr: false,
			check: func(t *testing.T, result time.Time) {
				yesterday := now.Add(-24 * time.Hour)
				if result.Year() != yesterday.Year() || result.Month() != yesterday.Month() || result.Day() != yesterday.Day() {
					t.Errorf("Parse('yesterday') = %v, want yesterday", result)
				}
			},
		},
		{
			name:    "tomorrow",
			input:   "tomorrow",
			wantErr: false,
			check: func(t *testing.T, result time.Time) {
				tomorrow := now.Add(24 * time.Hour)
				if result.Year() != tomorrow.Year() || result.Month() != tomorrow.Month() || result.Day() != tomorrow.Day() {
					t.Errorf("Parse('tomorrow') = %v, want tomorrow", result)
				}
			},
		},
		{
			name:    "7 days ago",
			input:   "7d",
			wantErr: false,
			check: func(t *testing.T, result time.Time) {
				sevenDaysAgo := now.Add(-7 * 24 * time.Hour)
				if result.Year() != sevenDaysAgo.Year() || result.Month() != sevenDaysAgo.Month() || result.Day() != sevenDaysAgo.Day() {
					t.Errorf("Parse('7d') = %v, want 7 days ago", result)
				}
			},
		},
		{
			name:    "2 weeks ago",
			input:   "2w",
			wantErr: false,
			check: func(t *testing.T, result time.Time) {
				twoWeeksAgo := now.Add(-14 * 24 * time.Hour)
				if result.Year() != twoWeeksAgo.Year() || result.Month() != twoWeeksAgo.Month() || result.Day() != twoWeeksAgo.Day() {
					t.Errorf("Parse('2w') = %v, want 2 weeks ago", result)
				}
			},
		},
		{
			name:    "3 months ago",
			input:   "3m",
			wantErr: false,
			check: func(t *testing.T, result time.Time) {
				threeMonthsAgo := now.Add(-90 * 24 * time.Hour)
				// Allow some variance for month approximation
				diff := threeMonthsAgo.Sub(result)
				if diff < 0 {
					diff = -diff
				}
				if diff > 2*24*time.Hour {
					t.Errorf("Parse('3m') = %v, want approximately 3 months ago", result)
				}
			},
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "invalid duration",
			input:   "abc123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	p := New()

	// Should not panic
	result := p.MustParse("2025-12-10")
	if result.Year() != 2025 {
		t.Errorf("MustParse() = %v, want 2025", result)
	}

	// Should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustParse() should panic on invalid input")
		}
	}()
	p.MustParse("invalid")
}
