package lexers

import (
	"fmt"

	"github.com/johnkerl/pgpg/pkg/tokens"
	"unicode/utf8"
)

const lineLexerInitialCapacity = 1024

const (
	LineLexerTypeLine = 1
)

// LineLexer is primarily for unit-test purposes. Every line is its own token.
type LineLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
}

func NewLineLexer(inputText string) AbstractLexer {
	return &LineLexer{
		inputText:     inputText,
		inputLength:   len(inputText),
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
		} else {
			runes = append(runes, r)
		}
	}

	retval := tokens.NewToken(
		runes,
		LineLexerTypeLine,
		&startLocation,
	)

	return retval
}

func (lexer *LineLexer) DecodeType(tokenType tokens.TokenType) (string, error) {
	switch tokenType {
	case tokens.TokenTypeEOF:
		return "EOF", nil
	case tokens.TokenTypeError:
		return "error", nil
	case LineLexerTypeLine:
		return "line", nil
	default:
		return "", fmt.Errorf("unrecognized token type %d", int(tokenType))
	}
}
