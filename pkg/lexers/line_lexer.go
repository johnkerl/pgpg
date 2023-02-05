package lexers

import (
	"github.com/johnkerl/pgpg/pkg/tokens"
	"unicode/utf8"
)

const lineLexerInitialCapacity = 1024

// LineLexer is primarily for unit-test purposes. Every line is its own token.
type LineLexer struct {
	inputText       string
	inputLength     int
	currentPosition int
	tokenLocation   *tokens.TokenLocation
}

func NewLineLexer(inputText string) AbstractLexer {
	return &LineLexer{
		inputText:       inputText,
		inputLength:     len(inputText),
		currentPosition: 0,
		tokenLocation:   tokens.NewTokenLocation(1, 1),
	}
}

func (lexer *LineLexer) Scan() (token *tokens.Token, err error) {
	if lexer.currentPosition >= lexer.inputLength {
		// TODO: define and return EOF token
		return nil, nil
	}

	startLocation := *lexer.tokenLocation
	runes := make([]rune, 0, lineLexerInitialCapacity)

	for lexer.currentPosition < lexer.inputLength {
		r, runeWidth := utf8.DecodeRuneInString(lexer.inputText[lexer.currentPosition:])
		lexer.currentPosition += runeWidth
		if r == '\n' {
			lexer.tokenLocation.LineNumber++
			lexer.tokenLocation.ColumnNumber = 1
			break
		} else {
			lexer.tokenLocation.ColumnNumber++
			runes = append(runes, r)
		}
	}

	retval := tokens.NewToken(runes, &startLocation)

	return retval, nil
}
