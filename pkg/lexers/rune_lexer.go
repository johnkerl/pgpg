package lexers

import (
	"github.com/johnkerl/pgpg/pkg/tokens"
	"unicode/utf8"
)

// RuneLexer is primarily for unit-test purposes. Every rune is its own token.
type RuneLexer struct {
	inputText       string
	inputLength     int
	currentPosition int
	tokenLocation   *tokens.TokenLocation
}

func NewRuneLexer(inputText string) AbstractLexer {
	return &RuneLexer{
		inputText:       inputText,
		inputLength:     len(inputText),
		currentPosition: 0,
		tokenLocation:   tokens.NewTokenLocation(1, 1),
	}
}

func (lexer *RuneLexer) Scan() (token *tokens.Token, err error) {
	if lexer.currentPosition >= lexer.inputLength {
		// TODO: define and return EOF token
		return nil, nil
	}

	r, runeWidth := utf8.DecodeRuneInString(lexer.inputText[lexer.currentPosition:])
	lexer.currentPosition += runeWidth

	retval := tokens.NewToken([]rune{r}, lexer.tokenLocation)

	if r == '\n' {
		lexer.tokenLocation.LineNumber++
		lexer.tokenLocation.ColumnNumber = 1
	} else {
		lexer.tokenLocation.ColumnNumber++
	}

	return retval, nil
}
