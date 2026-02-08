package lexers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------------------
func TestVICLexer1(t *testing.T) {
	lexer := NewVICLexer("")

	token := lexer.Scan()
	assert.True(t, token.IsEOF())
}

// ----------------------------------------------------------------
func TestVICLexer2(t *testing.T) {
	lexer := NewVICLexer("123")

	token := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "123")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)
	assert.Equal(t, token.Type, VICLexerTypeNumber)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

// ----------------------------------------------------------------
func TestVICLexer3(t *testing.T) {
	lexer := NewVICLexer("+-*/**()=foo9")

	token := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "+")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)
	assert.Equal(t, token.Type, VICLexerTypePlus)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "-")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 2)
	assert.Equal(t, token.Type, VICLexerTypeMinus)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "*")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 3)
	assert.Equal(t, token.Type, VICLexerTypeTimes)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "/")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 4)
	assert.Equal(t, token.Type, VICLexerTypeDivide)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "**")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 5)
	assert.Equal(t, token.Type, VICLexerTypePower)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "(")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 7)
	assert.Equal(t, token.Type, VICLexerTypeLParen)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), ")")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 8)
	assert.Equal(t, token.Type, VICLexerTypeRParen)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "=")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 9)
	assert.Equal(t, token.Type, VICLexerTypeAssign)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "foo9")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 10)
	assert.Equal(t, token.Type, VICLexerTypeIdentifier)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

// ----------------------------------------------------------------
func TestVICLexerIdentifiers(t *testing.T) {
	lexer := NewVICLexer("_aZ9 __x0")

	token := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "_aZ9")
	assert.Equal(t, token.Type, VICLexerTypeIdentifier)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "__x0")
	assert.Equal(t, token.Type, VICLexerTypeIdentifier)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

// ----------------------------------------------------------------
func TestVICLexer4(t *testing.T) {
	lexer := NewVICLexer("123&456")

	token := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "123")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)
	assert.Equal(t, token.Type, VICLexerTypeNumber)

	token = lexer.Scan()
	assert.True(t, token.IsError())
}
