package lexers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------------------
func TestCannedTextLexer1(t *testing.T) {
	lexer := NewCannedTextLexer("")

	token, err := lexer.Scan()
	assert.Nil(t, token)
	assert.Nil(t, err)
}

// ----------------------------------------------------------------
func TestCannedTextLexer2(t *testing.T) {
	lexer := NewCannedTextLexer("a b c")

	token, err := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "a")
	assert.Nil(t, err)

	token, err = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "b")
	assert.Nil(t, err)

	token, err = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "c")
	assert.Nil(t, err)

	token, err = lexer.Scan()
	assert.Nil(t, token)
	assert.Nil(t, err)
}
