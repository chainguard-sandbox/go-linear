package fieldfilter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		fieldSpec string
		wantErr   bool
		wantLen   int
	}{
		{
			name:      "empty spec returns nil",
			fieldSpec: "",
			wantErr:   false,
			wantLen:   0,
		},
		{
			name:      "single field",
			fieldSpec: "id",
			wantErr:   false,
			wantLen:   1,
		},
		{
			name:      "multiple fields",
			fieldSpec: "id,title,priority",
			wantErr:   false,
			wantLen:   3,
		},
		{
			name:      "nested field",
			fieldSpec: "assignee.name",
			wantErr:   false,
			wantLen:   1,
		},
		{
			name:      "fields with spaces",
			fieldSpec: "id, title, assignee.name",
			wantErr:   false,
			wantLen:   3,
		},
		{
			name:      "empty field in list",
			fieldSpec: "id,,title",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := New(tt.fieldSpec, nil) // No defaults for these tests
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.wantLen == 0 && fs != nil {
					t.Errorf("New() = %v, want nil for empty spec", fs)
				}
				if tt.wantLen > 0 && len(fs.fields) != tt.wantLen {
					t.Errorf("New() fields count = %d, want %d", len(fs.fields), tt.wantLen)
				}
			}
		})
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name      string
		fieldSpec string
		input     string
		want      string
		wantErr   bool
	}{
		{
			name:      "nil selector returns unchanged",
			fieldSpec: "",
			input:     `{"id":"123","title":"Test"}`,
			want:      `{"id":"123","title":"Test"}`,
		},
		{
			name:      "single field selection",
			fieldSpec: "id",
			input:     `{"id":"123","title":"Test","priority":1}`,
			want:      `{"id":"123"}`,
		},
		{
			name:      "multiple fields",
			fieldSpec: "id,title",
			input:     `{"id":"123","title":"Test","priority":1}`,
			want:      `{"id":"123","title":"Test"}`,
		},
		{
			name:      "nested field selection",
			fieldSpec: "id,assignee.name",
			input:     `{"id":"123","assignee":{"name":"Alice","email":"alice@example.com"},"title":"Test"}`,
			want:      `{"assignee":{"name":"Alice"},"id":"123"}`,
		},
		{
			name:      "array filtering",
			fieldSpec: "id,title",
			input:     `[{"id":"1","title":"A","priority":1},{"id":"2","title":"B","priority":2}]`,
			want:      `[{"id":"1","title":"A"},{"id":"2","title":"B"}]`,
		},
		{
			name:      "nonexistent fields filtered out",
			fieldSpec: "id,nonexistent",
			input:     `{"id":"123","title":"Test"}`,
			want:      `{"id":"123"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := New(tt.fieldSpec, nil)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			got, err := fs.Filter([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Filter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Normalize JSON for comparison (remove whitespace variations)
				gotStr := strings.TrimSpace(string(got))
				wantStr := strings.TrimSpace(tt.want)

				// Parse and re-marshal both to normalize
				var gotObj, wantObj any
				if err := json.Unmarshal([]byte(gotStr), &gotObj); err != nil {
					t.Fatalf("Failed to parse got JSON: %v", err)
				}
				if err := json.Unmarshal([]byte(wantStr), &wantObj); err != nil {
					t.Fatalf("Failed to parse want JSON: %v", err)
				}

				gotNorm, _ := json.Marshal(gotObj)
				wantNorm, _ := json.Marshal(wantObj)

				if !bytes.Equal(gotNorm, wantNorm) {
					t.Errorf("Filter() = %s, want %s", string(gotNorm), string(wantNorm))
				}
			}
		})
	}
}

func TestWithNullable(t *testing.T) {
	input := []byte(`{"nodes":[{"id":"a","slaBreachesAt":"2026-01-01T00:00:00Z"},{"id":"b"}]}`)

	t.Run("pads selected nullable fields in list items only", func(t *testing.T) {
		fs, err := NewForList("id,slaBreachesAt,completedAt", nil)
		if err != nil {
			t.Fatal(err)
		}
		// completedAt is nullable and selected; slaStartedAt is nullable but not selected.
		fs = fs.WithNullable("slaBreachesAt", "completedAt", "slaStartedAt")

		out, err := fs.Filter(input)
		if err != nil {
			t.Fatal(err)
		}
		want := `{"nodes":[{"completedAt":null,"id":"a","slaBreachesAt":"2026-01-01T00:00:00Z"},{"completedAt":null,"id":"b","slaBreachesAt":null}]}`
		if string(out) != want {
			t.Errorf("Filter() =\n%s\nwant\n%s", out, want)
		}
	})

	t.Run("without WithNullable absent fields stay omitted", func(t *testing.T) {
		fs, err := NewForList("id,slaBreachesAt", nil)
		if err != nil {
			t.Fatal(err)
		}
		out, err := fs.Filter(input)
		if err != nil {
			t.Fatal(err)
		}
		want := `{"nodes":[{"id":"a","slaBreachesAt":"2026-01-01T00:00:00Z"},{"id":"b"}]}`
		if string(out) != want {
			t.Errorf("Filter() =\n%s\nwant\n%s", out, want)
		}
	})

	t.Run("nil selector without fields stays nil", func(t *testing.T) {
		var fs *FieldSelector
		if got := fs.WithNullable(); got != nil {
			t.Errorf("WithNullable() on nil = %v, want nil", got)
		}
	})

	t.Run("nil selector keeps all fields and pads top-level nodes only", func(t *testing.T) {
		fs, err := NewForList("none", nil)
		if err != nil {
			t.Fatal(err)
		}
		fs = fs.WithNullable("completedAt", "slaBreachesAt")

		in := []byte(`{"nodes":[{"id":"a","n":12345678901234567890,"labels":{"nodes":[{"id":"l"}]}},{"id":"b","completedAt":"x"}],"pageInfo":{"hasNextPage":false}}`)
		out, err := fs.Filter(in)
		if err != nil {
			t.Fatal(err)
		}
		want := `{"nodes":[{"completedAt":null,"id":"a","labels":{"nodes":[{"id":"l"}]},"n":12345678901234567890,"slaBreachesAt":null},{"completedAt":"x","id":"b","slaBreachesAt":null}],"pageInfo":{"hasNextPage":false}}`
		if string(out) != want {
			t.Errorf("Filter() =\n%s\nwant\n%s", out, want)
		}
	})

	t.Run("nil selector keeps indented input indented", func(t *testing.T) {
		var fs *FieldSelector
		fs = fs.WithNullable("completedAt")
		out, err := fs.Filter([]byte("{\n  \"nodes\": [\n    {\n      \"id\": \"a\"\n    }\n  ]\n}"))
		if err != nil {
			t.Fatal(err)
		}
		want := "{\n  \"nodes\": [\n    {\n      \"completedAt\": null,\n      \"id\": \"a\"\n    }\n  ]\n}"
		if string(out) != want {
			t.Errorf("Filter() =\n%s\nwant\n%s", out, want)
		}
	})

	t.Run("nil selector leaves non-list output alone", func(t *testing.T) {
		var fs *FieldSelector
		fs = fs.WithNullable("completedAt")
		in := []byte(`{"id":"a"}`)
		out, err := fs.Filter(in)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(out, in) {
			t.Errorf("Filter() = %s, want %s", out, in)
		}
	})
}
