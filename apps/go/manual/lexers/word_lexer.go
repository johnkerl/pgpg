package lexers

import (
	"io"
	"unicode"
	"unicode/utf8"

	"github.com/johnkerl/pgpg/go/lib/pkg/tokens"
)

const wordLexerInitialCapacity = 1024

const (
	WordLexerTypeWord tokens.TokenType = "word"
)

// WordLexer is for unit-test purposes, as well as perhaps a layer underneath the lexer for the SENG
// reference grammar. Every word is its own token, where "word" is defined as whitespace-delimited.
type WordLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
}

func NewWordLexer(r io.Reader) AbstractLexer {
	b, _ := io.ReadAll(r)
	return NewWordLexerFromString(string(b))
}

func NewWordLexerFromString(s string) AbstractLexer {
	return &WordLexer{
		inputText:     s,
		inputLength:   len(s),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

func (lexer *WordLexer) Scan() (token *tokens.Token) {
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	lexer.ignoreNextRunesIf(unicode.IsSpace)
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	startLocation := *lexer.tokenLocation
	runes := make([]rune, 0, wordLexerInitialCapacity)

	for lexer.tokenLocation.ByteOffset < lexer.inputLength {
		r := lexer.readRune()
		if unicode.IsSpace(r) {
			break
		}
		runes = append(runes, r)
	}

	return tokens.NewToken(runes, WordLexerTypeWord, &startLocation)
}

func (lexer *WordLexer) ignoreNextRuneIf(predicate RunePredicateFunc) bool {
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return false
	}
	r, runeWidth := lexer.peekRune()
	if runeWidth == 0 {
		return false
	}

	if predicate(r) {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return true
	}
	return false
}

func (lexer *WordLexer) ignoreNextRunesIf(predicate RunePredicateFunc) {
	for lexer.ignoreNextRuneIf(predicate) {
	}
}

func (lexer *WordLexer) peekRune() (rune, int) {
	r, runeWidth := utf8.DecodeRuneInString(lexer.inputText[lexer.tokenLocation.ByteOffset:])
	return r, runeWidth
}

func (lexer *WordLexer) readRune() rune {
	r, runeWidth := lexer.peekRune()
	lexer.tokenLocation.LocateRune(r, runeWidth)
	return r
}
