package lexers

import (
	"io"
	"unicode/utf8"

	"github.com/johnkerl/pgpg/go/lib/pkg/tokens"
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

func NewRuneLexer(r io.Reader) AbstractLexer {
	b, _ := io.ReadAll(r)
	return NewRuneLexerFromString(string(b))
}

func NewRuneLexerFromString(s string) AbstractLexer {
	return &RuneLexer{
		inputText:     s,
		inputLength:   len(s),
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
