package common

import (
	"context"
	"testing"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/dateparser"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// mockDateFilterable implements DateFilterable for testing.
type mockDateFilterable struct {
	createdAt *intgraphql.DateComparator
	parser    dateparser.Parser
}

func newMockDateFilterable() *mockDateFilterable {
	return &mockDateFilterable{
		createdAt: &intgraphql.DateComparator{},
		parser:    dateparser.New(),
	}
}

func (m *mockDateFilterable) CreatedAtComparator() *intgraphql.DateComparator {
	return m.createdAt
}

func (m *mockDateFilterable) Parser() dateparser.Parser {
	return m.parser
}

// mockUpdateDateFilterable implements UpdateDateFilterable for testing.
type mockUpdateDateFilterable struct {
	*mockDateFilterable
	updatedAt *intgraphql.DateComparator
}

func newMockUpdateDateFilterable() *mockUpdateDateFilterable {
	return &mockUpdateDateFilterable{
		mockDateFilterable: newMockDateFilterable(),
		updatedAt:          &intgraphql.DateComparator{},
	}
}

func (m *mockUpdateDateFilterable) UpdatedAtComparator() *intgraphql.DateComparator {
	return m.updatedAt
}

// mockIDFilterable implements IDFilterable for testing.
type mockIDFilterable struct {
	id *intgraphql.IDComparator
}

func (m *mockIDFilterable) SetID(comp *intgraphql.IDComparator) {
	m.id = comp
}

func TestApplyCreatedAt(t *testing.T) {
	tests := []struct {
		name       string
		after      string
		before     string
		wantAfter  bool
		wantBefore bool
		wantErr    bool
	}{
		{
			name:       "no flags set",
			after:      "",
			before:     "",
			wantAfter:  false,
			wantBefore: false,
			wantErr:    false,
		},
		{
			name:       "only after set",
			after:      "2024-01-01",
			before:     "",
			wantAfter:  true,
			wantBefore: false,
			wantErr:    false,
		},
		{
			name:       "only before set",
			after:      "",
			before:     "2024-12-31",
			wantAfter:  false,
			wantBefore: true,
			wantErr:    false,
		},
		{
			name:       "both set",
			after:      "2024-01-01",
			before:     "2024-12-31",
			wantAfter:  true,
			wantBefore: true,
			wantErr:    false,
		},
		{
			name:       "relative date yesterday",
			after:      "yesterday",
			before:     "",
			wantAfter:  true,
			wantBefore: false,
			wantErr:    false,
		},
		{
			name:       "relative date 7d",
			after:      "7d",
			before:     "",
			wantAfter:  true,
			wantBefore: false,
			wantErr:    false,
		},
		{
			name:    "invalid after date",
			after:   "not-a-date",
			before:  "",
			wantErr: true,
		},
		{
			name:    "invalid before date",
			after:   "",
			before:  "not-a-date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("created-after", "", "")
			cmd.Flags().String("created-before", "", "")

			if tt.after != "" {
				_ = cmd.Flags().Set("created-after", tt.after)
			}
			if tt.before != "" {
				_ = cmd.Flags().Set("created-before", tt.before)
			}

			mock := newMockDateFilterable()
			err := ApplyCreatedAt(context.Background(), cmd, mock)

			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyCreatedAt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.wantAfter && mock.createdAt.Gte == nil {
					t.Error("expected Gte to be set")
				}
				if !tt.wantAfter && mock.createdAt.Gte != nil {
					t.Error("expected Gte to be nil")
				}
				if tt.wantBefore && mock.createdAt.Lte == nil {
					t.Error("expected Lte to be set")
				}
				if !tt.wantBefore && mock.createdAt.Lte != nil {
					t.Error("expected Lte to be nil")
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
		wantAfter  bool
		wantBefore bool
		wantErr    bool
	}{
		{
			name:       "no flags set",
			after:      "",
			before:     "",
			wantAfter:  false,
			wantBefore: false,
			wantErr:    false,
		},
		{
			name:       "only after set",
			after:      "2024-01-01",
			before:     "",
			wantAfter:  true,
			wantBefore: false,
			wantErr:    false,
		},
		{
			name:       "only before set",
			after:      "",
			before:     "2024-12-31",
			wantAfter:  false,
			wantBefore: true,
			wantErr:    false,
		},
		{
			name:       "both set",
			after:      "2024-01-01",
			before:     "2024-12-31",
			wantAfter:  true,
			wantBefore: true,
			wantErr:    false,
		},
		{
			name:    "invalid after date",
			after:   "invalid",
			before:  "",
			wantErr: true,
		},
		{
			name:    "invalid before date",
			after:   "",
			before:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("updated-after", "", "")
			cmd.Flags().String("updated-before", "", "")

			if tt.after != "" {
				_ = cmd.Flags().Set("updated-after", tt.after)
			}
			if tt.before != "" {
				_ = cmd.Flags().Set("updated-before", tt.before)
			}

			mock := newMockUpdateDateFilterable()
			err := ApplyUpdatedAt(context.Background(), cmd, mock)

			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyUpdatedAt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.wantAfter && mock.updatedAt.Gte == nil {
					t.Error("expected Gte to be set")
				}
				if !tt.wantAfter && mock.updatedAt.Gte != nil {
					t.Error("expected Gte to be nil")
				}
				if tt.wantBefore && mock.updatedAt.Lte == nil {
					t.Error("expected Lte to be set")
				}
				if !tt.wantBefore && mock.updatedAt.Lte != nil {
					t.Error("expected Lte to be nil")
				}
			}
		})
	}
}

func TestApplyID(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		wantID bool
	}{
		{
			name:   "no id set",
			id:     "",
			wantID: false,
		},
		{
			name:   "id set",
			id:     "abc-123",
			wantID: true,
		},
		{
			name:   "uuid format",
			id:     "12345678-1234-1234-1234-123456789abc",
			wantID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("id", "", "")

			if tt.id != "" {
				_ = cmd.Flags().Set("id", tt.id)
			}

			mock := &mockIDFilterable{}
			err := ApplyID(context.Background(), cmd, mock)

			if err != nil {
				t.Errorf("ApplyID() error = %v", err)
				return
			}

			if tt.wantID {
				if mock.id == nil {
					t.Error("expected id to be set")
				} else if mock.id.Eq == nil || *mock.id.Eq != tt.id {
					t.Errorf("id = %v, want %q", mock.id.Eq, tt.id)
				}
			} else {
				if mock.id != nil {
					t.Error("expected id to be nil")
				}
			}
		})
	}
}
