package lexers

import (
	"github.com/johnkerl/pgpg/pkg/tokens"
	"unicode/utf8"
)

// RuneLexer is primarily for unit-test purposes. Every rune is its own token.
type RuneLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
}

func NewRuneLexer(inputText string) AbstractLexer {
	return &RuneLexer{
		inputText:     inputText,
		inputLength:   len(inputText),
		tokenLocation: tokens.NewDefaultTokenLocation(),
	}
}

func (lexer *RuneLexer) Scan() (token *tokens.Token, err error) {
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		// TODO: define and return EOF token
		return nil, nil
	}

	r, runeWidth := utf8.DecodeRuneInString(lexer.inputText[lexer.tokenLocation.ByteOffset:])
	lexer.tokenLocation.ByteOffset += runeWidth

	retval := tokens.NewToken([]rune{r}, lexer.tokenLocation)

	if r == '\n' {
		lexer.tokenLocation.LineNumber++
		lexer.tokenLocation.ColumnNumber = 1
	} else {
		lexer.tokenLocation.ColumnNumber++
	}

	return retval, nil
}
