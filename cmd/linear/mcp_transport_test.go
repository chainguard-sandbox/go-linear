package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestFixDoubleEncodedFlags(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNil bool // true = no fix needed
		wantOut string
	}{
		{
			name:    "already an object - no fix",
			input:   `{"name":"go-linear_issue_update","arguments":{"flags":{"state":"Done"}}}`,
			wantNil: true,
		},
		{
			name:    "double-encoded flags - fixed",
			input:   `{"name":"go-linear_issue_update","arguments":{"flags":"{\"state\":\"Done\"}"}}`,
			wantOut: `{"name":"go-linear_issue_update","arguments":{"flags":{"state":"Done"}}}`,
		},
		{
			name:    "no flags key - no fix",
			input:   `{"name":"go-linear_issue_list","arguments":{}}`,
			wantNil: true,
		},
		{
			name:    "double-encoded with multiple fields",
			input:   `{"name":"go-linear_issue_update","arguments":{"flags":"{\"issue\":\"ENG-1\",\"body\":\"test\"}"}}`,
			wantOut: `{"name":"go-linear_issue_update","arguments":{"flags":{"issue":"ENG-1","body":"test"}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fixDoubleEncodedFlags(json.RawMessage(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %s", result)
				}
				return
			}
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			// Normalize by round-tripping through map to avoid key-order differences
			var got, want map[string]any
			if err := json.Unmarshal(result, &got); err != nil {
				t.Fatalf("result is invalid JSON: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.wantOut), &want); err != nil {
				t.Fatalf("wantOut is invalid JSON: %v", err)
			}
			gotB, _ := json.Marshal(got)
			wantB, _ := json.Marshal(want)
			if !bytes.Equal(gotB, wantB) {
				t.Errorf("got %s, want %s", gotB, wantB)
			}
		})
	}
}
