package lexers

import (
	"fmt"

	"github.com/johnkerl/pgpg/pkg/tokens"
	"unicode/utf8"
)

const (
	RuneLexerRuneType = 1
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

func (lexed *RuneLexer) DecodeType(tokenType tokens.TokenType) (string, error) {
	switch tokenType {
	case tokens.TokenTypeEOF:
		return "EOF", nil
	case tokens.TokenTypeError:
		return "error", nil
	case RuneLexerRuneType:
		return "rune", nil
	default:
		return "", fmt.Errorf("unrecognized token type %d", int(tokenType))
	}
}
