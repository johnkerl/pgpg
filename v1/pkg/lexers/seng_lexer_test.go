package lexers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------------------
func TestSENGLexer1(t *testing.T) {
	lexer := NewSENGLexer("")

	token := lexer.Scan()
	assert.True(t, token.IsEOF())
}

// ----------------------------------------------------------------
func TestSENGLexer2(t *testing.T) {
	lexer := NewSENGLexer("the the \n dog")

	token := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "the")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "the")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 5)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "dog")
	assert.Equal(t, token.Location.LineNumber, 2)
	assert.Equal(t, token.Location.ColumnNumber, 2)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

// ----------------------------------------------------------------
func TestSENGLexer3(t *testing.T) {
	lexer := NewSENGLexer(" the nonesuch goes")

	token := lexer.Scan()
	assert.NotNil(t, token)
	assert.Equal(t, string(token.Lexeme), "the")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 2)

	token = lexer.Scan()
	assert.NotNil(t, token)
	assert.True(t, token.IsError())
}
