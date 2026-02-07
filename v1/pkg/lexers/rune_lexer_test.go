package lexers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuneLexer1(t *testing.T) {
	lexer := NewRuneLexer("")

	token := lexer.Scan()
	assert.NotNil(t, token)
	assert.True(t, token.IsEOF())
}

func TestRuneLexer2(t *testing.T) {
	lexer := NewRuneLexer("abc")

	token := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "a")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "b")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 2)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "c")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 3)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

func TestRuneLexer3(t *testing.T) {
	lexer := NewRuneLexer("a\nbc")

	token := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "a")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "\n")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 2)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "b")
	assert.Equal(t, token.Location.LineNumber, 2)
	assert.Equal(t, token.Location.ColumnNumber, 1)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "c")
	assert.Equal(t, token.Location.LineNumber, 2)
	assert.Equal(t, token.Location.ColumnNumber, 2)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}
