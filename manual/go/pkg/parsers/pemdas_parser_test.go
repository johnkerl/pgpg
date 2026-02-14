package parsers

import (
	"testing"

	"github.com/johnkerl/pgpg/manual/go/pkg/asts"
	"github.com/johnkerl/pgpg/manual/go/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/go/pkg/tokens"
	"github.com/stretchr/testify/assert"
)

func assertOpNode(
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

func assertNumberLeaf(
	t *testing.T,
	node *asts.ASTNode,
	expectedLexeme string,
) {
	t.Helper()
	if assert.NotNil(t, node) && assert.NotNil(t, node.Token) {
		assert.Equal(t, lexers.PEMDASLexerTypeNumber, node.Token.Type)
		assert.Equal(t, expectedLexeme, node.Token.LexemeText())
		assert.Nil(t, node.Children)
	}
}

func TestPEMDASParserPowerRightAssociative(t *testing.T) {
	parser := NewPEMDASParser()
	ast, err := parser.Parse("2**3**4")
	assert.NoError(t, err)

	root := ast.RootNode
	assertOpNode(t, root, lexers.PEMDASLexerTypePower, 2)
	assertNumberLeaf(t, root.Children[0], "2")

	right := root.Children[1]
	assertOpNode(t, right, lexers.PEMDASLexerTypePower, 2)
	assertNumberLeaf(t, right.Children[0], "3")
	assertNumberLeaf(t, right.Children[1], "4")
}

func TestPEMDASParserPrecedenceTimesOverPlus(t *testing.T) {
	parser := NewPEMDASParser()
	ast, err := parser.Parse("2+3*4")
	assert.NoError(t, err)

	root := ast.RootNode
	assertOpNode(t, root, lexers.PEMDASLexerTypePlus, 2)
	assertNumberLeaf(t, root.Children[0], "2")

	right := root.Children[1]
	assertOpNode(t, right, lexers.PEMDASLexerTypeTimes, 2)
	assertNumberLeaf(t, right.Children[0], "3")
	assertNumberLeaf(t, right.Children[1], "4")
}

func TestPEMDASParserParentheses(t *testing.T) {
	parser := NewPEMDASParser()
	ast, err := parser.Parse("2*(3+4)")
	assert.NoError(t, err)

	root := ast.RootNode
	assertOpNode(t, root, lexers.PEMDASLexerTypeTimes, 2)
	assertNumberLeaf(t, root.Children[0], "2")

	right := root.Children[1]
	assertOpNode(t, right, lexers.PEMDASLexerTypePlus, 2)
	assertNumberLeaf(t, right.Children[0], "3")
	assertNumberLeaf(t, right.Children[1], "4")
}

func TestPEMDASParserUnaryMinus(t *testing.T) {
	parser := NewPEMDASParser()
	ast, err := parser.Parse("-2**3")
	assert.NoError(t, err)

	root := ast.RootNode
	assertOpNode(t, root, lexers.PEMDASLexerTypePower, 2)

	left := root.Children[0]
	assertOpNode(t, left, lexers.PEMDASLexerTypeMinus, 1)
	assertNumberLeaf(t, left.Children[0], "2")

	assertNumberLeaf(t, root.Children[1], "3")
}
