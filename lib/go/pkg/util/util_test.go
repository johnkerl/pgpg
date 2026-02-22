package util

import (
	"testing"
)

func TestSplitString(t *testing.T) {
	tests := []struct {
		input     string
		separator string
		want      []string
	}{
		{"", ",", nil},
		{"a", ",", []string{"a"}},
		{"a,b,c", ",", []string{"a", "b", "c"}},
		{"a b c", " ", []string{"a", "b", "c"}},
		{"x", ":", []string{"x"}},
		{"a:b", ":", []string{"a", "b"}},
	}
	for _, tt := range tests {
		got := SplitString(tt.input, tt.separator)
		if tt.want == nil {
			if got != nil {
				t.Errorf("SplitString(%q, %q) = %v, want nil", tt.input, tt.separator, got)
			}
			continue
		}
		if len(got) != len(tt.want) {
			t.Errorf("SplitString(%q, %q) length = %d, want %d", tt.input, tt.separator, len(got), len(tt.want))
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("SplitString(%q, %q)[%d] = %q, want %q", tt.input, tt.separator, i, got[i], tt.want[i])
			}
		}
	}
}
