package parsers

import (
	"testing"

	"github.com/johnkerl/pgpg/manual/pkg/asts"
	"github.com/stretchr/testify/assert"
)

func assertEBNFNodeType(t *testing.T, node *asts.ASTNode, nodeType asts.NodeType) {
	t.Helper()
	if assert.NotNil(t, node) {
		assert.Equal(t, nodeType, node.Type)
	}
}

func TestEBNFParserSimpleGrammar(t *testing.T) {
	parser := NewEBNFParser()
	ast, err := parser.Parse("rule = \"a\" | \"b\";")
	assert.NoError(t, err)

	root := ast.RootNode
	assertEBNFNodeType(t, root, EBNFParserNodeTypeGrammar)
	assert.Len(t, root.Children, 1)

	rule := root.Children[0]
	assertEBNFNodeType(t, rule, EBNFParserNodeTypeRule)
	assert.Len(t, rule.Children, 2)
}

func TestEBNFParserGroupingOptionalRepeat(t *testing.T) {
	parser := NewEBNFParser()
	ast, err := parser.Parse("expr ::= term { ( \"+\" | \"-\" ) term } [ \";\" ]")
	assert.NoError(t, err)

	root := ast.RootNode
	assertEBNFNodeType(t, root, EBNFParserNodeTypeGrammar)
	assert.Len(t, root.Children, 1)
}

func TestEBNFParserMultipleRules(t *testing.T) {
	parser := NewEBNFParser()
	ast, err := parser.Parse("a = \"x\"; b = \"y\";")
	assert.NoError(t, err)
	assert.Len(t, ast.RootNode.Children, 2)
}

func TestEBNFParserMissingRuleName(t *testing.T) {
	parser := NewEBNFParser()
	_, err := parser.Parse("= \"x\";")
	assert.Error(t, err)
}

func TestEBNFParserHintSimple(t *testing.T) {
	parser := NewEBNFParser()
	ast, err := parser.Parse(`A ::= B C D -> { "parent": 1, "children": [0, 2], "type": "sum" };`)
	assert.NoError(t, err)

	root := ast.RootNode
	assertEBNFNodeType(t, root, EBNFParserNodeTypeGrammar)
	assert.Len(t, root.Children, 1)

	rule := root.Children[0]
	assertEBNFNodeType(t, rule, EBNFParserNodeTypeRule)
	assert.Len(t, rule.Children, 2)

	// The expression should be a hinted_sequence
	expr := rule.Children[1]
	assertEBNFNodeType(t, expr, EBNFParserNodeTypeHintedSequence)
	assert.Len(t, expr.Children, 2)

	// Child 0 is the sequence
	seq := expr.Children[0]
	assertEBNFNodeType(t, seq, EBNFParserNodeTypeSequence)
	assert.Len(t, seq.Children, 3)

	// Child 1 is the hint
	hint := expr.Children[1]
	assertEBNFNodeType(t, hint, EBNFParserNodeTypeHint)
	assert.Len(t, hint.Children, 3) // parent, children, type
}

func TestEBNFParserHintPerAlternative(t *testing.T) {
	parser := NewEBNFParser()
	ast, err := parser.Parse(`
		Op ::= Left plus Right  -> { "parent": 1, "children": [0, 2] }
		     | Left minus Right -> { "parent": 1, "children": [0, 2] };
	`)
	assert.NoError(t, err)

	root := ast.RootNode
	rule := root.Children[0]
	expr := rule.Children[1]

	// Should be alternates with two hinted_sequence children
	assertEBNFNodeType(t, expr, EBNFParserNodeTypeAlternates)
	assert.Len(t, expr.Children, 2)

	assertEBNFNodeType(t, expr.Children[0], EBNFParserNodeTypeHintedSequence)
	assertEBNFNodeType(t, expr.Children[1], EBNFParserNodeTypeHintedSequence)
}

func TestEBNFParserHintEmptyChildren(t *testing.T) {
	parser := NewEBNFParser()
	ast, err := parser.Parse(`A ::= B -> { "parent": 0, "children": [] };`)
	assert.NoError(t, err)

	rule := ast.RootNode.Children[0]
	expr := rule.Children[1]
	assertEBNFNodeType(t, expr, EBNFParserNodeTypeHintedSequence)

	hint := expr.Children[1]
	assertEBNFNodeType(t, hint, EBNFParserNodeTypeHint)
	// children field should have an empty array
	childrenField := hint.Children[1]
	assertEBNFNodeType(t, childrenField, EBNFParserNodeTypeHintField)
	assertEBNFNodeType(t, childrenField.Children[0], EBNFParserNodeTypeHintArray)
	assert.Len(t, childrenField.Children[0].Children, 0)
}

func TestEBNFParserNoHintStillWorks(t *testing.T) {
	// A grammar without hints should parse as before
	parser := NewEBNFParser()
	ast, err := parser.Parse(`A ::= B C D;`)
	assert.NoError(t, err)

	rule := ast.RootNode.Children[0]
	expr := rule.Children[1]
	assertEBNFNodeType(t, expr, EBNFParserNodeTypeSequence)
}

func TestEBNFParserHintWithAppendedChildren(t *testing.T) {
	parser := NewEBNFParser()
	ast, err := parser.Parse(`A ::= B C D E -> { "parent": 1, "with_appended_children": [2, 3] };`)
	assert.NoError(t, err)

	root := ast.RootNode
	rule := root.Children[0]
	expr := rule.Children[1]
	assertEBNFNodeType(t, expr, EBNFParserNodeTypeHintedSequence)
	assert.Len(t, expr.Children, 2)

	hint := expr.Children[1]
	assertEBNFNodeType(t, hint, EBNFParserNodeTypeHint)
	// Hint should have two fields: parent and with_appended_children
	assert.Len(t, hint.Children, 2)
}

func TestEBNFParserHintWithPrependedChildren(t *testing.T) {
	parser := NewEBNFParser()
	ast, err := parser.Parse(`A ::= B C D E -> { "parent": 1, "with_prepended_children": [0, 2] };`)
	assert.NoError(t, err)

	root := ast.RootNode
	rule := root.Children[0]
	expr := rule.Children[1]
	assertEBNFNodeType(t, expr, EBNFParserNodeTypeHintedSequence)
	assert.Len(t, expr.Children, 2)

	hint := expr.Children[1]
	assertEBNFNodeType(t, hint, EBNFParserNodeTypeHint)
	// Hint should have two fields: parent and with_prepended_children
	assert.Len(t, hint.Children, 2)
}

func TestEBNFParserHintWithAdoptedGrandchildren(t *testing.T) {
	parser := NewEBNFParser()
	ast, err := parser.Parse(`A ::= B C D -> { "parent": 1, "with_adopted_grandchildren": [0, 2] };`)
	assert.NoError(t, err)

	root := ast.RootNode
	rule := root.Children[0]
	expr := rule.Children[1]
	assertEBNFNodeType(t, expr, EBNFParserNodeTypeHintedSequence)
	assert.Len(t, expr.Children, 2)

	hint := expr.Children[1]
	assertEBNFNodeType(t, hint, EBNFParserNodeTypeHint)
	// Hint should have two fields: parent and with_adopted_grandchildren
	assert.Len(t, hint.Children, 2)
}
