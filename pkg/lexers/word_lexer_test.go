package lexers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------------------
func TestWordLexer1(t *testing.T) {
	lexer := NewWordLexer("")

	token, err := lexer.Scan()
	assert.Nil(t, token)
	assert.Nil(t, err)
}

// ----------------------------------------------------------------
func TestWordLexer2(t *testing.T) {
	lexer := NewWordLexer("abc")

	token, err := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "abc")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)
	assert.Nil(t, err)

	token, err = lexer.Scan()
	assert.Nil(t, token)
	assert.Nil(t, err)
}

// ----------------------------------------------------------------
func TestWordLexer3(t *testing.T) {
	lexer := NewWordLexer(" abc  def   \n ghi ")

	token, err := lexer.Scan()
	assert.NotNil(t, token)
	assert.Nil(t, err)
	assert.Equal(t, string(token.Lexeme), "abc")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 2)

	token, err = lexer.Scan()
	assert.NotNil(t, token)
	assert.Nil(t, err)
	assert.Equal(t, string(token.Lexeme), "def")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 7)

	token, err = lexer.Scan()
	assert.NotNil(t, token)
	assert.Nil(t, err)
	assert.Equal(t, string(token.Lexeme), "ghi")
	assert.Equal(t, token.Location.LineNumber, 2)
	assert.Equal(t, token.Location.ColumnNumber, 2)

	token, err = lexer.Scan()
	assert.Nil(t, token)
	assert.Nil(t, err)
}
