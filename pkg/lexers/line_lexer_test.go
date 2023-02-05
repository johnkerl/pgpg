package lexers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------------------
func TestLineLexer1(t *testing.T) {
	lexer := NewLineLexer("")

	token := lexer.Scan()
	assert.True(t, token.IsEOF())
}

// ----------------------------------------------------------------
func TestLineLexer2(t *testing.T) {
	lexer := NewLineLexer("abc")

	token := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "abc")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

// ----------------------------------------------------------------
func TestLineLexer3(t *testing.T) {
	lexer := NewLineLexer("a\nbc")

	token := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "a")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "bc")
	assert.Equal(t, token.Location.LineNumber, 2)
	assert.Equal(t, token.Location.ColumnNumber, 1)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

// ----------------------------------------------------------------
func TestLineLexer4(t *testing.T) {
	lexer := NewLineLexer("\n\n\n")

	token := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "")
	assert.Equal(t, token.Location.LineNumber, 2)
	assert.Equal(t, token.Location.ColumnNumber, 1)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "")
	assert.Equal(t, token.Location.LineNumber, 3)
	assert.Equal(t, token.Location.ColumnNumber, 1)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}
