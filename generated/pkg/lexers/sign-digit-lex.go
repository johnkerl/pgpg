package lexers

import (
	"fmt"
	"unicode/utf8"

	manuallexers "github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type SignDigitLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
}

var _ manuallexers.AbstractLexer = (*SignDigitLexer)(nil)

func NewSignDigitLexer(inputText string) manuallexers.AbstractLexer {
	return &SignDigitLexer{
		inputText:     inputText,
		inputLength:   len(inputText),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

func (lexer *SignDigitLexer) Scan() *tokens.Token {
	for {
		if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
			return tokens.NewEOFToken(lexer.tokenLocation)
		}

		startLocation := *lexer.tokenLocation
		scanLocation := *lexer.tokenLocation
		state := SignDigitLexerStartState
		lastAcceptState := -1
		lastAcceptLocation := scanLocation

		for {
			if scanLocation.ByteOffset >= lexer.inputLength {
				break
			}
			r, width := lexer.peekRuneAt(scanLocation.ByteOffset)
			nextState, ok := SignDigitLexerLookupTransition(state, r)
			if !ok {
				break
			}
			scanLocation.LocateRune(r, width)
			state = nextState
			if _, ok := SignDigitLexerActions[state]; ok {
				lastAcceptState = state
				lastAcceptLocation = scanLocation
			}
		}

		if lastAcceptState < 0 {
			r, _ := lexer.peekRuneAt(lexer.tokenLocation.ByteOffset)
			return tokens.NewErrorToken(fmt.Sprintf("lexer: unrecognized input %q", r), lexer.tokenLocation)
		}

		lexemeText := lexer.inputText[lexer.tokenLocation.ByteOffset:lastAcceptLocation.ByteOffset]
		lexeme := []rune(lexemeText)
		*lexer.tokenLocation = lastAcceptLocation
		tokenType := SignDigitLexerActions[lastAcceptState]
		return tokens.NewToken(lexeme, tokenType, &startLocation)
	}
}

func (lexer *SignDigitLexer) peekRuneAt(byteOffset int) (rune, int) {
	r, width := utf8.DecodeRuneInString(lexer.inputText[byteOffset:])
	return r, width
}

func SignDigitLexerLookupTransition(state int, r rune) (int, bool) {
	transitionsForState, ok := SignDigitLexerTransitions[state]
	if !ok {
		return 0, false
	}
	for _, tr := range transitionsForState {
		if r < tr.from {
			return 0, false
		}
		if r >= tr.from && r <= tr.to {
			return tr.next, true
		}
	}
	return 0, false
}

const SignDigitLexerStartState = 0

type SignDigitLexerRangeTransition struct {
	from rune
	to   rune
	next int
}

var SignDigitLexerTransitions = map[int][]SignDigitLexerRangeTransition{
	0: {
		{from: '+', to: '+', next: 1},
		{from: '-', to: '-', next: 2},
		{from: '0', to: '0', next: 3},
		{from: '1', to: '1', next: 4},
		{from: '2', to: '2', next: 5},
		{from: '3', to: '3', next: 6},
		{from: '4', to: '4', next: 7},
		{from: '5', to: '5', next: 8},
		{from: '6', to: '6', next: 9},
		{from: '7', to: '7', next: 10},
		{from: '8', to: '8', next: 11},
		{from: '9', to: '9', next: 12},
	},
}

var SignDigitLexerActions = map[int]tokens.TokenType{
	1: "sign",
	2: "sign",
	3: "digit",
	4: "digit",
	5: "digit",
	6: "digit",
	7: "digit",
	8: "digit",
	9: "digit",
	10: "digit",
	11: "digit",
	12: "digit",
}
