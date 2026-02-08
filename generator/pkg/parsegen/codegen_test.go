package parsegen

import (
	"strings"
	"testing"
)

func TestGenerateGoParserCodeFormats(t *testing.T) {
	tables := &Tables{
		StartSymbol: "Root",
		Actions:     map[int]map[string]Action{},
		Gotos:       map[int]map[string]int{},
		Productions: []Production{{LHS: "Root", RHS: []Symbol{{Name: "EOF", Terminal: true}}}},
	}
	code, err := GenerateGoParserCode(tables, "parsers", "TestParser")
	if err != nil {
		t.Fatalf("GenerateGoParserCode() error: %v", err)
	}
	if len(code) == 0 {
		t.Fatalf("GenerateGoParserCode() returned empty code")
	}
	if !strings.Contains(string(code), "package parsers") {
		t.Fatalf("generated code missing package declaration")
	}
}
