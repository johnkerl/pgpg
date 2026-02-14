package lexers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCannedTextLexer1(t *testing.T) {
	lexer := NewCannedTextLexer("")

	token := lexer.Scan()
	assert.NotNil(t, token)
	assert.True(t, token.IsEOF())
}

func TestCannedTextLexer2(t *testing.T) {
	lexer := NewCannedTextLexer("a b c")

	token := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "a")

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "b")

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "c")

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}
