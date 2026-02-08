package parsers

import (
	"testing"

	"github.com/johnkerl/pgpg/pkg/asts"
	"github.com/johnkerl/pgpg/pkg/lexers"
	"github.com/johnkerl/pgpg/pkg/tokens"
	"github.com/stretchr/testify/assert"
)

func assertVICOpNode(
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

func assertVICNumberLeaf(
	t *testing.T,
	node *asts.ASTNode,
	expectedLexeme string,
) {
	t.Helper()
	if assert.NotNil(t, node) && assert.NotNil(t, node.Token) {
		assert.Equal(t, lexers.VICLexerTypeNumber, node.Token.Type)
		assert.Equal(t, expectedLexeme, node.Token.LexemeText())
		assert.Nil(t, node.Children)
	}
}

func assertVICIdentifierLeaf(
	t *testing.T,
	node *asts.ASTNode,
	expectedLexeme string,
) {
	t.Helper()
	if assert.NotNil(t, node) && assert.NotNil(t, node.Token) {
		assert.Equal(t, lexers.VICLexerTypeIdentifier, node.Token.Type)
		assert.Equal(t, expectedLexeme, node.Token.LexemeText())
		assert.Nil(t, node.Children)
	}
}

// ----------------------------------------------------------------
func TestVICParserPowerRightAssociative(t *testing.T) {
	parser := NewVICParser()
	ast, err := parser.Parse("2**3**4")
	assert.NoError(t, err)

	root := ast.RootNode
	assertVICOpNode(t, root, lexers.VICLexerTypePower, 2)
	assertVICNumberLeaf(t, root.Children[0], "2")

	right := root.Children[1]
	assertVICOpNode(t, right, lexers.VICLexerTypePower, 2)
	assertVICNumberLeaf(t, right.Children[0], "3")
	assertVICNumberLeaf(t, right.Children[1], "4")
}

// ----------------------------------------------------------------
func TestVICParserPrecedenceTimesOverPlus(t *testing.T) {
	parser := NewVICParser()
	ast, err := parser.Parse("2+3*4")
	assert.NoError(t, err)

	root := ast.RootNode
	assertVICOpNode(t, root, lexers.VICLexerTypePlus, 2)
	assertVICNumberLeaf(t, root.Children[0], "2")

	right := root.Children[1]
	assertVICOpNode(t, right, lexers.VICLexerTypeTimes, 2)
	assertVICNumberLeaf(t, right.Children[0], "3")
	assertVICNumberLeaf(t, right.Children[1], "4")
}

// ----------------------------------------------------------------
func TestVICParserParentheses(t *testing.T) {
	parser := NewVICParser()
	ast, err := parser.Parse("2*(3+4)")
	assert.NoError(t, err)

	root := ast.RootNode
	assertVICOpNode(t, root, lexers.VICLexerTypeTimes, 2)
	assertVICNumberLeaf(t, root.Children[0], "2")

	right := root.Children[1]
	assertVICOpNode(t, right, lexers.VICLexerTypePlus, 2)
	assertVICNumberLeaf(t, right.Children[0], "3")
	assertVICNumberLeaf(t, right.Children[1], "4")
}

// ----------------------------------------------------------------
func TestVICParserUnaryMinus(t *testing.T) {
	parser := NewVICParser()
	ast, err := parser.Parse("-2**3")
	assert.NoError(t, err)

	root := ast.RootNode
	assertVICOpNode(t, root, lexers.VICLexerTypePower, 2)

	left := root.Children[0]
	assertVICOpNode(t, left, lexers.VICLexerTypeMinus, 1)
	assertVICNumberLeaf(t, left.Children[0], "2")

	assertVICNumberLeaf(t, root.Children[1], "3")
}

// ----------------------------------------------------------------
func TestVICParserIdentifiers(t *testing.T) {
	parser := NewVICParser()
	ast, err := parser.Parse("x+1")
	assert.NoError(t, err)

	root := ast.RootNode
	assertVICOpNode(t, root, lexers.VICLexerTypePlus, 2)
	assertVICIdentifierLeaf(t, root.Children[0], "x")
	assertVICNumberLeaf(t, root.Children[1], "1")
}

// ----------------------------------------------------------------
func TestVICParserAssignment(t *testing.T) {
	parser := NewVICParser()
	ast, err := parser.Parse("x = x + 1")
	assert.NoError(t, err)

	root := ast.RootNode
	assert.Equal(t, VICParserNodeTypeAssignment, root.Type)
	assert.Equal(t, lexers.VICLexerTypeAssign, root.Token.Type)
	assert.Len(t, root.Children, 2)
	assertVICIdentifierLeaf(t, root.Children[0], "x")

	right := root.Children[1]
	assertVICOpNode(t, right, lexers.VICLexerTypePlus, 2)
	assertVICIdentifierLeaf(t, right.Children[0], "x")
	assertVICNumberLeaf(t, right.Children[1], "1")
}

// ----------------------------------------------------------------
func TestVICParserAssignmentRequiresIdentifier(t *testing.T) {
	parser := NewVICParser()
	_, err := parser.Parse("1 = 1 + 2")
	assert.Error(t, err)
}
