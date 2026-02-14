package parsers

import (
	"testing"

	"github.com/johnkerl/pgpg/manual/go/pkg/asts"
	"github.com/johnkerl/pgpg/manual/go/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/go/pkg/tokens"
	"github.com/stretchr/testify/assert"
)

func assertVBCOpNode(
	t *testing.T,
	node *asts.ASTNode,
	expectedType tokens.TokenType,
	expectedChildren int,
) {
	t.Helper()
	if assert.NotNil(t, node) && assert.NotNil(t, node.Token) {
		assert.Equal(t, expectedType, node.Token.Type)
		assert.Len(t, node.Children, expectedChildren)
	}
}

func assertVBCIdentifierLeaf(
	t *testing.T,
	node *asts.ASTNode,
	expectedLexeme string,
) {
	t.Helper()
	if assert.NotNil(t, node) && assert.NotNil(t, node.Token) {
		assert.Equal(t, lexers.VBCLexerTypeIdentifier, node.Token.Type)
		assert.Equal(t, expectedLexeme, node.Token.LexemeText())
		assert.Nil(t, node.Children)
	}
}

func TestVBCParserPrecedenceAndOverOr(t *testing.T) {
	parser := NewVBCParser()
	ast, err := parser.Parse("a OR b AND c")
	assert.NoError(t, err)

	root := ast.RootNode
	assertVBCOpNode(t, root, lexers.VBCLexerTypeOr, 2)
	assertVBCIdentifierLeaf(t, root.Children[0], "a")

	right := root.Children[1]
	assertVBCOpNode(t, right, lexers.VBCLexerTypeAnd, 2)
	assertVBCIdentifierLeaf(t, right.Children[0], "b")
	assertVBCIdentifierLeaf(t, right.Children[1], "c")
}

func TestVBCParserNotPrecedence(t *testing.T) {
	parser := NewVBCParser()
	ast, err := parser.Parse("NOT a AND b")
	assert.NoError(t, err)

	root := ast.RootNode
	assertVBCOpNode(t, root, lexers.VBCLexerTypeAnd, 2)

	left := root.Children[0]
	assertVBCOpNode(t, left, lexers.VBCLexerTypeNot, 1)
	assertVBCIdentifierLeaf(t, left.Children[0], "a")

	assertVBCIdentifierLeaf(t, root.Children[1], "b")
}

func TestVBCParserParentheses(t *testing.T) {
	parser := NewVBCParser()
	ast, err := parser.Parse("(a OR b) AND c")
	assert.NoError(t, err)

	root := ast.RootNode
	assertVBCOpNode(t, root, lexers.VBCLexerTypeAnd, 2)

	left := root.Children[0]
	assertVBCOpNode(t, left, lexers.VBCLexerTypeOr, 2)
	assertVBCIdentifierLeaf(t, left.Children[0], "a")
	assertVBCIdentifierLeaf(t, left.Children[1], "b")

	assertVBCIdentifierLeaf(t, root.Children[1], "c")
}

func TestVBCParserIdentifierOnly(t *testing.T) {
	parser := NewVBCParser()
	ast, err := parser.Parse("_flag1")
	assert.NoError(t, err)

	root := ast.RootNode
	assertVBCIdentifierLeaf(t, root, "_flag1")
}
