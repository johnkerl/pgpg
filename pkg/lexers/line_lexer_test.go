package lexers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------------------
func TestLineLexer1(t *testing.T) {
	lexer := NewLineLexer("")

	token, err := lexer.Scan()
	assert.Nil(t, token)
	assert.Nil(t, err)
}

// ----------------------------------------------------------------
func TestLineLexer2(t *testing.T) {
	lexer := NewLineLexer("abc")

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
func TestLineLexer3(t *testing.T) {
	lexer := NewLineLexer("a\nbc")

	token, err := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "a")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)
	assert.Nil(t, err)

	token, err = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "bc")
	assert.Equal(t, token.Location.LineNumber, 2)
	assert.Equal(t, token.Location.ColumnNumber, 1)
	assert.Nil(t, err)

	token, err = lexer.Scan()
	assert.Nil(t, token)
	assert.Nil(t, err)
}

// ----------------------------------------------------------------
func TestLineLexer4(t *testing.T) {
	lexer := NewLineLexer("\n\n\n")

	token, err := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)
	assert.Nil(t, err)

	token, err = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "")
	assert.Equal(t, token.Location.LineNumber, 2)
	assert.Equal(t, token.Location.ColumnNumber, 1)
	assert.Nil(t, err)

	token, err = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "")
	assert.Equal(t, token.Location.LineNumber, 3)
	assert.Equal(t, token.Location.ColumnNumber, 1)
	assert.Nil(t, err)

	token, err = lexer.Scan()
	assert.Nil(t, token)
	assert.Nil(t, err)
}
