package lexers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------------------
func TestEBNFLexer1(t *testing.T) {
	lexer := NewEBNFLexer("")

	token := lexer.Scan()
	assert.True(t, token.IsEOF())
}

// ----------------------------------------------------------------
func TestEBNFLexer2(t *testing.T) {
	lexer := NewEBNFLexer("rule ::= \"a\" | 'b' ;")

	token := lexer.Scan()
	assert.Equal(t, "rule", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeIdentifier, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "::=", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeAssign, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "\"a\"", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeString, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "|", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeOr, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "'b'", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeString, token.Type)

	token = lexer.Scan()
	assert.Equal(t, ";", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeSemicolon, token.Type)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

// ----------------------------------------------------------------
func TestEBNFLexer3(t *testing.T) {
	lexer := NewEBNFLexer("[ { ( ) } ] =")

	token := lexer.Scan()
	assert.Equal(t, "[", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeLBracket, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "{", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeLBrace, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "(", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeLParen, token.Type)

	token = lexer.Scan()
	assert.Equal(t, ")", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeRParen, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "}", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeRBrace, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "]", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeRBracket, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "=", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeAssign, token.Type)
}

// ----------------------------------------------------------------
func TestEBNFLexer4(t *testing.T) {
	lexer := NewEBNFLexer("x :=")

	token := lexer.Scan()
	assert.Equal(t, "x", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeIdentifier, token.Type)

	token = lexer.Scan()
	assert.True(t, token.IsError())
}
