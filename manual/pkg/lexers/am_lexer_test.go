package lexers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAMLexer1(t *testing.T) {
	lexer := NewAMLexer("")

	token := lexer.Scan()
	assert.True(t, token.IsEOF())
}

func TestAMLexer2(t *testing.T) {
	lexer := NewAMLexer("123")

	token := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "123")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)
	assert.Equal(t, token.Type, AMLexerTypeNumber)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

func TestAMLexer3(t *testing.T) {
	lexer := NewAMLexer("++*8888")

	token := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "+")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)
	assert.Equal(t, token.Type, AMLexerTypePlus)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "+")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 2)
	assert.Equal(t, token.Type, AMLexerTypePlus)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "*")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 3)
	assert.Equal(t, token.Type, AMLexerTypeTimes)

	token = lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "8888")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 4)
	assert.Equal(t, token.Type, AMLexerTypeNumber)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

func TestAMLexer4(t *testing.T) {
	lexer := NewAMLexer("123&456")

	token := lexer.Scan()
	assert.Equal(t, string(token.Lexeme), "123")
	assert.Equal(t, token.Location.LineNumber, 1)
	assert.Equal(t, token.Location.ColumnNumber, 1)
	assert.Equal(t, token.Type, AMLexerTypeNumber)

	token = lexer.Scan()
	assert.True(t, token.IsError())
}
