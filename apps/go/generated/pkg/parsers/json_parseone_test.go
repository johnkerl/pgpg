package parsers

import (
	"strings"
	"testing"

	"github.com/johnkerl/pgpg/apps/go/generated/pkg/lexers"
)

// TestJSONParseOneMixedType verifies that multi-object input with mixed value types
// (e.g. [] {} or {} [] or 1 2 3 [] 5) parses correctly via ParseOne. This depends on
// the parse table generator adding reduce for all of First(Value) in value-complete states.
func TestJSONParseOneMixedType(t *testing.T) {
	tests := []struct {
		input   string
		wantN   int
		wantErr bool
	}{
		{"[] {}", 2, false},
		{"{} []", 2, false},
		{"1 2 3 [] 5", 5, false},
		{"[] []", 2, false},
		{"{} {}", 2, false},
		{"1 2 3", 3, false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lex := lexers.NewJSONLexer(strings.NewReader(tt.input))
			parser := NewJSONParser()
			var n int
			for {
				ast, done, err := parser.ParseOne(lex, "")
				if err != nil {
					if !tt.wantErr {
						t.Errorf("ParseOne (object %d) error: %v", n+1, err)
					}
					return
				}
				if ast != nil {
					n++
				}
				if done {
					break
				}
			}
			if n != tt.wantN {
				t.Errorf("got %d top-level values, want %d", n, tt.wantN)
			}
		})
	}
}
