package lexers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------------------
func TestVBCLexer1(t *testing.T) {
	lexer := NewVBCLexer("")

	token := lexer.Scan()
	assert.True(t, token.IsEOF())
}

// ----------------------------------------------------------------
func TestVBCLexer2(t *testing.T) {
	lexer := NewVBCLexer("and OR Not foo9 (_bar)")

	token := lexer.Scan()
	assert.Equal(t, "and", token.LexemeText())
	assert.Equal(t, VBCLexerTypeAnd, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "OR", token.LexemeText())
	assert.Equal(t, VBCLexerTypeOr, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "Not", token.LexemeText())
	assert.Equal(t, VBCLexerTypeNot, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "foo9", token.LexemeText())
	assert.Equal(t, VBCLexerTypeIdentifier, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "(", token.LexemeText())
	assert.Equal(t, VBCLexerTypeLParen, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "_bar", token.LexemeText())
	assert.Equal(t, VBCLexerTypeIdentifier, token.Type)

	token = lexer.Scan()
	assert.Equal(t, ")", token.LexemeText())
	assert.Equal(t, VBCLexerTypeRParen, token.Type)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

// ----------------------------------------------------------------
func TestVBCLexer3(t *testing.T) {
	lexer := NewVBCLexer("x&y")

	token := lexer.Scan()
	assert.Equal(t, "x", token.LexemeText())
	assert.Equal(t, VBCLexerTypeIdentifier, token.Type)

	token = lexer.Scan()
	assert.True(t, token.IsError())
}
