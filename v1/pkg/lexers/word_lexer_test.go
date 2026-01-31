package lexers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------------------
func TestWordLexer1(t *testing.T) {
	lexer := NewWordLexer("")

	token := lexer.Scan()
	assert.True(t, token.IsEOF())
}

// ----------------------------------------------------------------
func TestWordLexer2(t *testing.T) {
	lexer := NewWordLexer("abc")

	token := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "abc")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

// ----------------------------------------------------------------
func TestWordLexer3(t *testing.T) {
	lexer := NewWordLexer(" abc  def   \n ghi ")

	token := lexer.Scan()
	assert.NotNil(t, token)
	assert.Equal(t, string(token.Lexeme), "abc")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 2)

	token = lexer.Scan()
	assert.NotNil(t, token)
	assert.Equal(t, string(token.Lexeme), "def")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 7)

	token = lexer.Scan()
	assert.NotNil(t, token)
	assert.Equal(t, string(token.Lexeme), "ghi")
	assert.Equal(t, token.Location.LineNumber, 2)
	assert.Equal(t, token.Location.ColumnNumber, 2)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}
