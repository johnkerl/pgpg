package lexers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------------------
func TestRuneLexer1(t *testing.T) {
	lexer := NewRuneLexer("")

	token, err := lexer.Scan()
	assert.Nil(t, token)
	assert.Nil(t, err)
}

// ----------------------------------------------------------------
func TestRuneLexer2(t *testing.T) {
	lexer := NewRuneLexer("abc")

	token, err := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "a")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)
	assert.Nil(t, err)

	token, err = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "b")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 2)
	assert.Nil(t, err)

	token, err = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "c")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 3)
	assert.Nil(t, err)

	token, err = lexer.Scan()
	assert.Nil(t, token)
	assert.Nil(t, err)
}

// ----------------------------------------------------------------
func TestRuneLexer3(t *testing.T) {
	lexer := NewRuneLexer("a\nbc")

	token, err := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "a")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)
	assert.Nil(t, err)

	token, err = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "\n")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 2)
	assert.Nil(t, err)

	token, err = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "b")
	assert.Equal(t, token.Location.LineNumber, 2)
	assert.Equal(t, token.Location.ColumnNumber, 1)
	assert.Nil(t, err)

	token, err = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "c")
	assert.Equal(t, token.Location.LineNumber, 2)
	assert.Equal(t, token.Location.ColumnNumber, 2)
	assert.Nil(t, err)

	token, err = lexer.Scan()
	assert.Nil(t, token)
	assert.Nil(t, err)
}
