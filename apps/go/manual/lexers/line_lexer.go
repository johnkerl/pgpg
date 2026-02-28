package lexers

import (
	"io"
	"unicode/utf8"

	"github.com/johnkerl/pgpg/go/lib/pkg/tokens"
)

const lineLexerInitialCapacity = 1024

const (
	LineLexerTypeLine tokens.TokenType = "line"
)

// LineLexer is primarily for unit-test purposes. Every line is its own token.
type LineLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
}

func NewLineLexer(r io.Reader) AbstractLexer {
	b, _ := io.ReadAll(r)
	return NewLineLexerFromString(string(b))
}

func NewLineLexerFromString(s string) AbstractLexer {
	return &LineLexer{
		inputText:     s,
		inputLength:   len(s),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

func (lexer *LineLexer) Scan() (token *tokens.Token) {
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	startLocation := *lexer.tokenLocation
	runes := make([]rune, 0, lineLexerInitialCapacity)

	for lexer.tokenLocation.ByteOffset < lexer.inputLength {
		r, runeWidth := utf8.DecodeRuneInString(lexer.inputText[lexer.tokenLocation.ByteOffset:])
		lexer.tokenLocation.LocateRune(r, runeWidth)
		if r == '\n' {
			break
		}
		runes = append(runes, r)
	}

	retval := tokens.NewToken(
		runes,
		LineLexerTypeLine,
		&startLocation,
	)

	return retval
}
