package lexers

import (
	"fmt"
	"strings"
	"unicode/utf8"

	manuallexers "github.com/johnkerl/pgpg/manual/go/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/go/pkg/tokens"
)

type PEMDASPlainLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
}

var _ manuallexers.AbstractLexer = (*PEMDASPlainLexer)(nil)

func NewPEMDASPlainLexer(inputText string) manuallexers.AbstractLexer {
	return &PEMDASPlainLexer{
		inputText:     inputText,
		inputLength:   len(inputText),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

func (lexer *PEMDASPlainLexer) Scan() *tokens.Token {
	for {
		if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
			return tokens.NewEOFToken(lexer.tokenLocation)
		}

		startLocation := *lexer.tokenLocation
		scanLocation := *lexer.tokenLocation
		state := PEMDASPlainLexerStartState
		lastAcceptState := -1
		lastAcceptLocation := scanLocation

		for {
			if scanLocation.ByteOffset >= lexer.inputLength {
				break
			}
			r, width := lexer.peekRuneAt(scanLocation.ByteOffset)
			nextState, ok := PEMDASPlainLexerLookupTransition(state, r)
			if !ok {
				break
			}
			scanLocation.LocateRune(r, width)
			state = nextState
			if _, ok := PEMDASPlainLexerActions[state]; ok {
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
		tokenType := PEMDASPlainLexerActions[lastAcceptState]
		if PEMDASPlainLexerIsIgnoredToken(tokenType) {
			continue
		}
		return tokens.NewToken(lexeme, tokenType, &startLocation)
	}
}

func (lexer *PEMDASPlainLexer) peekRuneAt(byteOffset int) (rune, int) {
	r, width := utf8.DecodeRuneInString(lexer.inputText[byteOffset:])
	return r, width
}

func PEMDASPlainLexerLookupTransition(state int, r rune) (int, bool) {
	transitionsForState, ok := PEMDASPlainLexerTransitions[state]
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
func PEMDASPlainLexerIsIgnoredToken(tokenType tokens.TokenType) bool {
	return strings.HasPrefix(string(tokenType), "!")
}

const PEMDASPlainLexerStartState = 0

type PEMDASPlainLexerRangeTransition struct {
	from rune
	to   rune
	next int
}

var PEMDASPlainLexerTransitions = map[int][]PEMDASPlainLexerRangeTransition{
	0: {
		{from: '\t', to: '\t', next: 1},
		{from: '\n', to: '\n', next: 2},
		{from: '\r', to: '\r', next: 3},
		{from: ' ', to: ' ', next: 4},
		{from: '%', to: '%', next: 5},
		{from: '(', to: '(', next: 6},
		{from: ')', to: ')', next: 7},
		{from: '*', to: '*', next: 8},
		{from: '+', to: '+', next: 9},
		{from: '-', to: '-', next: 10},
		{from: '/', to: '/', next: 11},
		{from: '0', to: '0', next: 12},
		{from: '1', to: '1', next: 13},
		{from: '2', to: '2', next: 14},
		{from: '3', to: '3', next: 15},
		{from: '4', to: '4', next: 16},
		{from: '5', to: '5', next: 17},
		{from: '6', to: '6', next: 18},
		{from: '7', to: '7', next: 19},
		{from: '8', to: '8', next: 20},
		{from: '9', to: '9', next: 21},
	},
	8: {
		{from: '*', to: '*', next: 22},
	},
	12: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	13: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	14: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	15: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	16: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	17: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	18: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	19: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	20: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	21: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	23: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	24: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	25: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	26: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	27: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	28: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	29: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	30: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	31: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
	32: {
		{from: '0', to: '0', next: 23},
		{from: '1', to: '1', next: 24},
		{from: '2', to: '2', next: 25},
		{from: '3', to: '3', next: 26},
		{from: '4', to: '4', next: 27},
		{from: '5', to: '5', next: 28},
		{from: '6', to: '6', next: 29},
		{from: '7', to: '7', next: 30},
		{from: '8', to: '8', next: 31},
		{from: '9', to: '9', next: 32},
	},
}

var PEMDASPlainLexerActions = map[int]tokens.TokenType{
	1:  "!whitespace",
	2:  "!whitespace",
	3:  "!whitespace",
	4:  "!whitespace",
	5:  "modulo",
	6:  "lparen",
	7:  "rparen",
	8:  "times",
	9:  "plus",
	10: "minus",
	11: "divide",
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
	22: "exponentiation",
	23: "int_literal",
	24: "int_literal",
	25: "int_literal",
	26: "int_literal",
	27: "int_literal",
	28: "int_literal",
	29: "int_literal",
	30: "int_literal",
	31: "int_literal",
	32: "int_literal",
}
