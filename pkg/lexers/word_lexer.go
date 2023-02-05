package lexers

import (
	"github.com/johnkerl/pgpg/pkg/tokens"
	"unicode"
	"unicode/utf8"
)

const wordLexerInitialCapacity = 1024

// WordLexer is for unit-test purposes, as well as perhaps a layer underneath the lexer for the SENG
// reference grammar. Every word is its own token, where "word" is defined as whitespace-delimited.
// Given this, "Hello, world!" would split to "Hello," and "world!" -- there is no special handling
// for punctuation in this lexer.
type WordLexer struct {
	inputText       string
	inputLength     int
	currentPosition int
	tokenLocation   *tokens.TokenLocation
}

func NewWordLexer(inputText string) AbstractLexer {
	return &WordLexer{
		inputText:       inputText,
		inputLength:     len(inputText),
		currentPosition: 0,
		tokenLocation:   tokens.NewTokenLocation(1, 1),
	}
}

func (lexer *WordLexer) Scan() (token *tokens.Token, err error) {
	if lexer.currentPosition >= lexer.inputLength {
		// TODO: define and return EOF token
		return nil, nil
	}

	// There are only two states: within a token or not (and OK the third state which is EOF). And
	// this lexer ignores whitespace between words -- not delivering them back to the caller -- and
	// loops over runes within a word until the word is ended. So this lexer doesn't need a state-tracker.

	// TODO: some trace-mode to optionally narrate this
	lexer.ignoreNextRunesIf(unicode.IsSpace)
	if lexer.currentPosition >= lexer.inputLength {
		// TODO: define and return EOF token
		return nil, nil
	}

	startLocation := *lexer.tokenLocation
	runes := make([]rune, 0, wordLexerInitialCapacity)

	for lexer.currentPosition < lexer.inputLength {
		r := lexer.readRune()
		if unicode.IsSpace(r) {
			break
		}
		runes = append(runes, r)
	}

	retval := tokens.NewToken(runes, &startLocation)

	return retval, nil
}

func (lexer *WordLexer) ignoreNextRuneIf(predicate runePredicateFunc) bool {
	// TODO explicit EOF handling
	r, runeWidth := utf8.DecodeRuneInString(lexer.inputText[lexer.currentPosition:])

	if predicate(r) {
		lexer.locateRune(r, runeWidth)
		return true
	} else {
		return false
	}
}

func (lexer *WordLexer) ignoreNextRunesIf(predicate runePredicateFunc) {
	// TODO explicit EOF handling
	for lexer.ignoreNextRuneIf(predicate) {
	}
}

// TODO: maybe move peekRune, readRune, acceptRune to abstract?

// peekRune gets the next rune from the input without updating location information.
func (lexer *WordLexer) peekRune() (rune, int) {
	r, runeWidth := utf8.DecodeRuneInString(lexer.inputText[lexer.currentPosition:])
	return r, runeWidth
}

// readRune gets the next rune from the input and updates location information.
func (lexer *WordLexer) readRune() rune {
	r, runeWidth := lexer.peekRune()
	lexer.locateRune(r, runeWidth)
	return r
}

// locateRune updates line/column number information for an accepted rune.
func (lexer *WordLexer) locateRune(r rune, runeWidth int) {
	if r == '\n' {
		lexer.tokenLocation.LineNumber++
		lexer.tokenLocation.ColumnNumber = 1
	} else {
		lexer.tokenLocation.ColumnNumber++
	}
	lexer.currentPosition += runeWidth
}
