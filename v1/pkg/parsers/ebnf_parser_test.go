package parsers

import (
	"testing"

	"github.com/johnkerl/pgpg/pkg/asts"
	"github.com/stretchr/testify/assert"
)

func assertEBNFNodeType(t *testing.T, node *asts.ASTNode, nodeType asts.NodeType) {
	t.Helper()
	if assert.NotNil(t, node) {
		assert.Equal(t, nodeType, node.Type)
	}
}

// ----------------------------------------------------------------
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

// ----------------------------------------------------------------
func TestEBNFParserGroupingOptionalRepeat(t *testing.T) {
	parser := NewEBNFParser()
	ast, err := parser.Parse("expr ::= term { ( \"+\" | \"-\" ) term } [ \";\" ]")
	assert.NoError(t, err)

	root := ast.RootNode
	assertEBNFNodeType(t, root, EBNFParserNodeTypeGrammar)
	assert.Len(t, root.Children, 1)
}

// ----------------------------------------------------------------
func TestEBNFParserMultipleRules(t *testing.T) {
	parser := NewEBNFParser()
	ast, err := parser.Parse("a = \"x\"; b = \"y\";")
	assert.NoError(t, err)
	assert.Len(t, ast.RootNode.Children, 2)
}

// ----------------------------------------------------------------
func TestEBNFParserMissingRuleName(t *testing.T) {
	parser := NewEBNFParser()
	_, err := parser.Parse("= \"x\";")
	assert.Error(t, err)
}
