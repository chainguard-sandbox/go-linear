package attachment

import (
	"context"
	"testing"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/filter/common"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

func TestNewFilterBuilder(t *testing.T) {
	b := NewFilterBuilder(nil)
	if b == nil {
		t.Fatal("NewFilterBuilder() returned nil")
	}
	if b.Filter() == nil {
		t.Fatal("Filter() returned nil")
	}
}

func TestFilterBuilder_Build_Empty(t *testing.T) {
	b := NewFilterBuilder(nil)
	filter := b.Build()
	if filter != nil {
		t.Errorf("Build() = %v, want nil for empty filter", filter)
	}
}

func TestApplyCreatedAt(t *testing.T) {
	tests := []struct {
		name       string
		after      string
		before     string
		wantFilter bool
		wantErr    bool
	}{
		{
			name:       "no flags",
			wantFilter: false,
		},
		{
			name:       "after only",
			after:      "2025-01-01",
			wantFilter: true,
		},
		{
			name:       "before only",
			before:     "2025-12-31",
			wantFilter: true,
		},
		{
			name:       "both flags",
			after:      "2025-01-01",
			before:     "2025-12-31",
			wantFilter: true,
		},
		{
			name:    "invalid date",
			after:   "not-a-date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("created-after", tt.after, "")
			cmd.Flags().String("created-before", tt.before, "")

			b := NewFilterBuilder(nil)
			err := common.ApplyCreatedAt(context.Background(), cmd, b)

			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyCreatedAt() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				hasFilter := b.filter.CreatedAt != nil
				if hasFilter != tt.wantFilter {
					t.Errorf("ApplyCreatedAt() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
				}
			}
		})
	}
}

func TestApplyUpdatedAt(t *testing.T) {
	tests := []struct {
		name       string
		after      string
		before     string
		wantFilter bool
		wantErr    bool
	}{
		{
			name:       "no flags",
			wantFilter: false,
		},
		{
			name:       "after only",
			after:      "2025-01-01",
			wantFilter: true,
		},
		{
			name:       "before only",
			before:     "2025-12-31",
			wantFilter: true,
		},
		{
			name:    "invalid date",
			before:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("updated-after", tt.after, "")
			cmd.Flags().String("updated-before", tt.before, "")

			b := NewFilterBuilder(nil)
			err := common.ApplyUpdatedAt(context.Background(), cmd, b)

			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyUpdatedAt() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				hasFilter := b.filter.UpdatedAt != nil
				if hasFilter != tt.wantFilter {
					t.Errorf("ApplyUpdatedAt() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
				}
			}
		})
	}
}

func TestApplyID(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		wantFilter bool
	}{
		{
			name:       "no id",
			wantFilter: false,
		},
		{
			name:       "with id",
			id:         "abc-123",
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("id", tt.id, "")

			b := NewFilterBuilder(nil)
			err := common.ApplyID(context.Background(), cmd, b)

			if err != nil {
				t.Fatalf("ApplyID() error = %v", err)
			}

			hasFilter := b.filter.ID != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyID() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplyTitle(t *testing.T) {
	tests := []struct {
		name       string
		title      string
		wantFilter bool
	}{
		{
			name:       "no title",
			wantFilter: false,
		},
		{
			name:       "with title",
			title:      "design",
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("title", tt.title, "")

			b := NewFilterBuilder(nil)
			err := ApplyTitle(context.Background(), cmd, b)

			if err != nil {
				t.Fatalf("ApplyTitle() error = %v", err)
			}

			hasFilter := b.filter.Title != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyTitle() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplySubtitle(t *testing.T) {
	tests := []struct {
		name       string
		subtitle   string
		wantFilter bool
	}{
		{
			name:       "no subtitle",
			wantFilter: false,
		},
		{
			name:       "with subtitle",
			subtitle:   "doc",
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("subtitle", tt.subtitle, "")

			b := NewFilterBuilder(nil)
			err := ApplySubtitle(context.Background(), cmd, b)

			if err != nil {
				t.Fatalf("ApplySubtitle() error = %v", err)
			}

			hasFilter := b.filter.Subtitle != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplySubtitle() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplyURL(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantFilter bool
	}{
		{
			name:       "no url",
			wantFilter: false,
		},
		{
			name:       "with url",
			url:        "github.com",
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("url", tt.url, "")

			b := NewFilterBuilder(nil)
			err := ApplyURL(context.Background(), cmd, b)

			if err != nil {
				t.Fatalf("ApplyURL() error = %v", err)
			}

			hasFilter := b.filter.URL != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplyURL() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestApplySourceType(t *testing.T) {
	tests := []struct {
		name       string
		sourceType string
		wantFilter bool
	}{
		{
			name:       "no source type",
			wantFilter: false,
		},
		{
			name:       "github",
			sourceType: "github",
			wantFilter: true,
		},
		{
			name:       "slack",
			sourceType: "slack",
			wantFilter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("source-type", tt.sourceType, "")

			b := NewFilterBuilder(nil)
			err := ApplySourceType(context.Background(), cmd, b)

			if err != nil {
				t.Fatalf("ApplySourceType() error = %v", err)
			}

			hasFilter := b.filter.SourceType != nil
			if hasFilter != tt.wantFilter {
				t.Errorf("ApplySourceType() hasFilter = %v, want %v", hasFilter, tt.wantFilter)
			}
		})
	}
}

func TestFilterBuilder_FromFlags(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("created-after", "2025-01-01", "")
	cmd.Flags().String("created-before", "", "")
	cmd.Flags().String("updated-after", "", "")
	cmd.Flags().String("updated-before", "", "")
	cmd.Flags().String("id", "", "")
	cmd.Flags().String("creator", "", "")
	cmd.Flags().String("source-type", "", "")
	cmd.Flags().String("title", "test", "")
	cmd.Flags().String("subtitle", "", "")
	cmd.Flags().String("url", "", "")

	b := NewFilterBuilder(nil)
	err := b.FromFlags(context.Background(), cmd)
	if err != nil {
		t.Fatalf("FromFlags() error = %v", err)
	}

	filter := b.Build()
	if filter == nil {
		t.Fatal("Build() returned nil, expected filter")
	}
	if filter.CreatedAt == nil {
		t.Error("expected CreatedAt filter to be set")
	}
	if filter.Title == nil {
		t.Error("expected Title filter to be set")
	}
}

func TestFilterBuilder_Build_WithFilters(t *testing.T) {
	b := NewFilterBuilder(nil)
	title := "test"
	b.filter.Title = &intgraphql.StringComparator{ContainsIgnoreCase: &title}

	filter := b.Build()
	if filter == nil {
		t.Fatal("Build() returned nil, expected filter")
	}
}
