package lexers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------------------
func TestPEMDASLexer1(t *testing.T) {
	lexer := NewPEMDASLexer("")

	token := lexer.Scan()
	assert.True(t, token.IsEOF())
}

// ----------------------------------------------------------------
func TestPEMDASLexer2(t *testing.T) {
	lexer := NewPEMDASLexer("123")

	token := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "123")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)
	assert.Equal(t, token.Type, PEMDASLexerTypeNumber)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

// ----------------------------------------------------------------
func TestPEMDASLexer3(t *testing.T) {
	lexer := NewPEMDASLexer("+-*/^()8888")

	token := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "+")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)
	assert.Equal(t, token.Type, PEMDASLexerTypePlus)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "-")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 2)
	assert.Equal(t, token.Type, PEMDASLexerTypeMinus)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "*")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 3)
	assert.Equal(t, token.Type, PEMDASLexerTypeTimes)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "/")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 4)
	assert.Equal(t, token.Type, PEMDASLexerTypeDivide)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "^")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 5)
	assert.Equal(t, token.Type, PEMDASLexerTypePower)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "(")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 6)
	assert.Equal(t, token.Type, PEMDASLexerTypeLParen)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), ")")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 7)
	assert.Equal(t, token.Type, PEMDASLexerTypeRParen)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "8888")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 8)
	assert.Equal(t, token.Type, PEMDASLexerTypeNumber)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

// ----------------------------------------------------------------
func TestPEMDASLexer4(t *testing.T) {
	lexer := NewPEMDASLexer("123&456")

	token := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "123")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)
	assert.Equal(t, token.Type, PEMDASLexerTypeNumber)

	token = lexer.Scan()
	assert.True(t, token.IsError())
}
