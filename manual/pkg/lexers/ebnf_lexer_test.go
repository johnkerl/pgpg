package lexers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEBNFLexer1(t *testing.T) {
	lexer := NewEBNFLexer("")

	token := lexer.Scan()
	assert.True(t, token.IsEOF())
}

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

func TestEBNFLexer4(t *testing.T) {
	// Standalone colon is now valid (used in AST hint blocks)
	lexer := NewEBNFLexer("x :=")

	token := lexer.Scan()
	assert.Equal(t, "x", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeIdentifier, token.Type)

	token = lexer.Scan()
	assert.Equal(t, ":", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeColon, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "=", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeAssign, token.Type)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

func TestEBNFLexerArrow(t *testing.T) {
	lexer := NewEBNFLexer(`A ::= B -> { "parent" : 0 }`)

	token := lexer.Scan()
	assert.Equal(t, "A", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeIdentifier, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "::=", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeAssign, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "B", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeIdentifier, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "->", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeArrow, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "{", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeLBrace, token.Type)

	token = lexer.Scan()
	assert.Equal(t, `"parent"`, token.LexemeText())
	assert.Equal(t, EBNFLexerTypeString, token.Type)

	token = lexer.Scan()
	assert.Equal(t, ":", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeColon, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "0", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeInteger, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "}", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeRBrace, token.Type)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

func TestEBNFLexerDashStillWorks(t *testing.T) {
	// Dash without > should still produce dash token (used in ranges)
	lexer := NewEBNFLexer(`"a"-"z"`)

	token := lexer.Scan()
	assert.Equal(t, `"a"`, token.LexemeText())
	assert.Equal(t, EBNFLexerTypeString, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "-", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeDash, token.Type)

	token = lexer.Scan()
	assert.Equal(t, `"z"`, token.LexemeText())
	assert.Equal(t, EBNFLexerTypeString, token.Type)
}

func TestEBNFLexerCommaAndIntegers(t *testing.T) {
	lexer := NewEBNFLexer("0, 12, 345")

	token := lexer.Scan()
	assert.Equal(t, "0", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeInteger, token.Type)

	token = lexer.Scan()
	assert.Equal(t, ",", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeComma, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "12", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeInteger, token.Type)

	token = lexer.Scan()
	assert.Equal(t, ",", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeComma, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "345", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeInteger, token.Type)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}
