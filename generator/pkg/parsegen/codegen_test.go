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

func TestGenerateGoParserCodeWithAppendedChildrenHint(t *testing.T) {
	tables := &Tables{
		StartSymbol: "Root",
		Actions:     map[int]map[string]Action{},
		Gotos:       map[int]map[string]int{},
		Productions: []Production{
			{
				LHS: "Root",
				RHS: []Symbol{
					{Name: "A", Terminal: false},
					{Name: "B", Terminal: false},
					{Name: "C", Terminal: false},
					{Name: "D", Terminal: false},
				},
				Hint: &ASTHint{
					ParentIndex:          1,
					WithAppendedChildren: []int{2, 3},
				},
			},
		},
		HintMode: "hints",
	}
	code, err := GenerateGoParserCode(tables, "parsers", "AppendTestParser")
	if err != nil {
		t.Fatalf("GenerateGoParserCode() error: %v", err)
	}
	codeStr := string(code)
	if !strings.Contains(codeStr, "hasWithAppendedChildren") {
		t.Fatalf("generated code should contain hasWithAppendedChildren")
	}
	if !strings.Contains(codeStr, "withAppendedChildren") {
		t.Fatalf("generated code should contain withAppendedChildren")
	}
}

func TestGenerateGoParserCodeWithPrependedChildrenHint(t *testing.T) {
	tables := &Tables{
		StartSymbol: "Root",
		Actions:     map[int]map[string]Action{},
		Gotos:       map[int]map[string]int{},
		Productions: []Production{
			{
				LHS: "Root",
				RHS: []Symbol{
					{Name: "A", Terminal: false},
					{Name: "B", Terminal: false},
					{Name: "C", Terminal: false},
					{Name: "D", Terminal: false},
				},
				Hint: &ASTHint{
					ParentIndex:           1,
					WithPrependedChildren: []int{0, 2},
				},
			},
		},
		HintMode: "hints",
	}
	code, err := GenerateGoParserCode(tables, "parsers", "PrependTestParser")
	if err != nil {
		t.Fatalf("GenerateGoParserCode() error: %v", err)
	}
	codeStr := string(code)
	if !strings.Contains(codeStr, "hasWithPrependedChildren") {
		t.Fatalf("generated code should contain hasWithPrependedChildren")
	}
	if !strings.Contains(codeStr, "withPrependedChildren") {
		t.Fatalf("generated code should contain withPrependedChildren")
	}
}

func TestGenerateGoParserCodeWithAdoptedGrandchildrenHint(t *testing.T) {
	tables := &Tables{
		StartSymbol: "Root",
		Actions:     map[int]map[string]Action{},
		Gotos:       map[int]map[string]int{},
		Productions: []Production{
			{
				LHS: "Root",
				RHS: []Symbol{
					{Name: "A", Terminal: false},
					{Name: "B", Terminal: false},
					{Name: "C", Terminal: false},
					{Name: "D", Terminal: false},
				},
				Hint: &ASTHint{
					ParentIndex:              1,
					WithAdoptedGrandchildren: []int{0, 2},
				},
			},
		},
		HintMode: "hints",
	}
	code, err := GenerateGoParserCode(tables, "parsers", "AdoptTestParser")
	if err != nil {
		t.Fatalf("GenerateGoParserCode() error: %v", err)
	}
	codeStr := string(code)
	if !strings.Contains(codeStr, "hasWithAdoptedGrandchildren") {
		t.Fatalf("generated code should contain hasWithAdoptedGrandchildren")
	}
	if !strings.Contains(codeStr, "withAdoptedGrandchildren") {
		t.Fatalf("generated code should contain withAdoptedGrandchildren")
	}
}

// TestGenerateGoParserCodeWithAdoptedGrandchildrenRespectsType verifies that when a
// production uses with_adopted_grandchildren and specifies a "type" hint, the generated
// parser applies that type to the result node (so e.g. Array with parent_literal "[]"
// and type "array" yields node type "array", not "[]").
func TestGenerateGoParserCodeWithAdoptedGrandchildrenRespectsType(t *testing.T) {
	tables := &Tables{
		StartSymbol: "Root",
		Actions:     map[int]map[string]Action{},
		Gotos:       map[int]map[string]int{},
		Productions: []Production{
			{
				LHS: "Root",
				RHS: []Symbol{
					{Name: "lbracket", Terminal: true},
					{Name: "Elements", Terminal: false},
					{Name: "rbracket", Terminal: true},
				},
				Hint: &ASTHint{
					ParentLiteral:             strPtr("[]"),
					WithAdoptedGrandchildren: []int{1},
					NodeType:                  "array",
				},
			},
		},
		HintMode: "hints",
	}
	code, err := GenerateGoParserCode(tables, "parsers", "ArrayTypeParser")
	if err != nil {
		t.Fatalf("GenerateGoParserCode() error: %v", err)
	}
	codeStr := string(code)
	// The template must apply prod.nodeType when set, so the result node gets type "array" not "[]".
	if !strings.Contains(codeStr, "nodeType := prod.nodeType") {
		t.Fatalf("generated code should apply hint type in with_adopted_grandchildren branch (nodeType := prod.nodeType)")
	}
	if !strings.Contains(codeStr, `asts.NodeType("array")`) {
		t.Fatalf("generated production table should include nodeType for the hint")
	}
	// Result node must be built with nodeType, not parentType, when type is set.
	if !strings.Contains(codeStr, "NewASTNode(parentToken, nodeType, newChildren)") {
		t.Fatalf("generated code should build node with nodeType in with_adopted_grandchildren branch")
	}
}

func strPtr(s string) *string { return &s }
