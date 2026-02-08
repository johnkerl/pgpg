package lexers

import (
	"github.com/johnkerl/pgpg/pkg/tokens"
	"unicode/utf8"
)

const (
	RuneLexerRuneType tokens.TokenType = "rune"
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
		tokenLocation: tokens.NewTokenLocation(),
	}
}

func (lexer *RuneLexer) Scan() (token *tokens.Token) {
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	r, runeWidth := utf8.DecodeRuneInString(lexer.inputText[lexer.tokenLocation.ByteOffset:])

	retval := tokens.NewToken(
		[]rune{r},
		RuneLexerRuneType,
		lexer.tokenLocation,
	)

	lexer.tokenLocation.LocateRune(r, runeWidth)

	return retval
}
