package lexers

import (
	"fmt"
	"strings"
	"unicode/utf8"

	manuallexers "github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type ArithLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
}

var _ manuallexers.AbstractLexer = (*ArithLexer)(nil)

func NewArithLexer(inputText string) manuallexers.AbstractLexer {
	return &ArithLexer{
		inputText:     inputText,
		inputLength:   len(inputText),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

func (lexer *ArithLexer) Scan() *tokens.Token {
	for {
		if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
			return tokens.NewEOFToken(lexer.tokenLocation)
		}

		startLocation := *lexer.tokenLocation
		scanLocation := *lexer.tokenLocation
		state := ArithLexerStartState
		lastAcceptState := -1
		lastAcceptLocation := scanLocation

		for {
			if scanLocation.ByteOffset >= lexer.inputLength {
				break
			}
			r, width := lexer.peekRuneAt(scanLocation.ByteOffset)
			nextState, ok := ArithLexerLookupTransition(state, r)
			if !ok {
				break
			}
			scanLocation.LocateRune(r, width)
			state = nextState
			if _, ok := ArithLexerActions[state]; ok {
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
		tokenType := ArithLexerActions[lastAcceptState]
		if ArithLexerIsIgnoredToken(tokenType) {
			continue
		}
		return tokens.NewToken(lexeme, tokenType, &startLocation)
	}
}

func (lexer *ArithLexer) peekRuneAt(byteOffset int) (rune, int) {
	r, width := utf8.DecodeRuneInString(lexer.inputText[byteOffset:])
	return r, width
}

func ArithLexerLookupTransition(state int, r rune) (int, bool) {
	transitionsForState, ok := ArithLexerTransitions[state]
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

func ArithLexerIsIgnoredToken(tokenType tokens.TokenType) bool {
	return strings.HasPrefix(string(tokenType), "!")
}

const ArithLexerStartState = 0

type ArithLexerRangeTransition struct {
	from rune
	to   rune
	next int
}

var ArithLexerTransitions = map[int][]ArithLexerRangeTransition{
	0: {
		{from: '\t', to: '\t', next: 1},
		{from: '\n', to: '\n', next: 2},
		{from: '\r', to: '\r', next: 3},
		{from: ' ', to: ' ', next: 4},
		{from: '%', to: '%', next: 5},
		{from: '*', to: '*', next: 6},
		{from: '+', to: '+', next: 7},
		{from: '-', to: '-', next: 8},
		{from: '/', to: '/', next: 9},
		{from: '0', to: '0', next: 10},
		{from: '1', to: '1', next: 11},
		{from: '2', to: '2', next: 12},
		{from: '3', to: '3', next: 13},
		{from: '4', to: '4', next: 14},
		{from: '5', to: '5', next: 15},
		{from: '6', to: '6', next: 16},
		{from: '7', to: '7', next: 17},
		{from: '8', to: '8', next: 18},
		{from: '9', to: '9', next: 19},
	},
	10: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	11: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	12: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	13: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	14: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	15: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	16: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	17: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	18: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	19: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	20: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	21: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	22: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	23: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	24: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	25: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	26: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	27: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	28: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
	29: {
		{from: '0', to: '0', next: 20},
		{from: '1', to: '1', next: 21},
		{from: '2', to: '2', next: 22},
		{from: '3', to: '3', next: 23},
		{from: '4', to: '4', next: 24},
		{from: '5', to: '5', next: 25},
		{from: '6', to: '6', next: 26},
		{from: '7', to: '7', next: 27},
		{from: '8', to: '8', next: 28},
		{from: '9', to: '9', next: 29},
	},
}

var ArithLexerActions = map[int]tokens.TokenType{
	1:  "!whitespace",
	2:  "!whitespace",
	3:  "!whitespace",
	4:  "!whitespace",
	5:  "modulo",
	6:  "times",
	7:  "plus",
	8:  "minus",
	9:  "divide",
	10: "int_literal",
	11: "int_literal",
	12: "int_literal",
	13: "int_literal",
	14: "int_literal",
	15: "int_literal",
	16: "int_literal",
	17: "int_literal",
	18: "int_literal",
	19: "int_literal",
	20: "int_literal",
	21: "int_literal",
	22: "int_literal",
	23: "int_literal",
	24: "int_literal",
	25: "int_literal",
	26: "int_literal",
	27: "int_literal",
	28: "int_literal",
	29: "int_literal",
}
