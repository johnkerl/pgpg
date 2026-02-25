package lexgen

import (
	"strings"
	"testing"
)

func TestGenerateGoLexerCodeFormats(t *testing.T) {
	tables := &Tables{
		StartState:  0,
		Transitions: map[int][]RangeTransition{},
		Actions:     map[int]string{},
	}
	code, err := GenerateCode(tables, LexCodegenOptions{Package: "lexers", Type: "TestLexer", Format: true})
	if err != nil {
		t.Fatalf("GenerateCode() error: %v", err)
	}
	if len(code) == 0 {
		t.Fatalf("GenerateCode() returned empty code")
	}
	if !strings.Contains(string(code), "package lexers") {
		t.Fatalf("generated code missing package declaration")
	}
}
