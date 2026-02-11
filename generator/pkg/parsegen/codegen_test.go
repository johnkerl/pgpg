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

func TestGenerateGoParserCodeHintMode(t *testing.T) {
	tables := &Tables{
		StartSymbol: "Root",
		Actions:     map[int]map[string]Action{},
		Gotos:       map[int]map[string]int{},
		Productions: []Production{
			{
				LHS: "Root",
				RHS: []Symbol{
					{Name: "A", Terminal: false},
					{Name: "plus", Terminal: true},
					{Name: "B", Terminal: false},
				},
				Hint: &ASTHint{
					ParentIndex:  1,
					ChildIndices: []int{0, 2},
					NodeType:     "operator",
				},
			},
		},
		HintMode: "hints",
	}
	code, err := GenerateGoParserCode(tables, "parsers", "HintTestParser")
	if err != nil {
		t.Fatalf("GenerateGoParserCode() error: %v", err)
	}
	codeStr := string(code)
	if !strings.Contains(codeStr, "hasHint") {
		t.Fatalf("hint-mode generated code should contain hasHint field")
	}
	if !strings.Contains(codeStr, "parentIndex") {
		t.Fatalf("hint-mode generated code should contain parentIndex field")
	}
	if !strings.Contains(codeStr, "childIndices") {
		t.Fatalf("hint-mode generated code should contain childIndices field")
	}
	if !strings.Contains(codeStr, `"operator"`) {
		t.Fatalf("hint-mode generated code should contain node type \"operator\"")
	}
}
