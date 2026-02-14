package lexers

import (
	"fmt"
	"strings"
	"unicode/utf8"

	manuallexers "github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type StatementsLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
}

var _ manuallexers.AbstractLexer = (*StatementsLexer)(nil)

func NewStatementsLexer(inputText string) manuallexers.AbstractLexer {
	return &StatementsLexer{
		inputText:     inputText,
		inputLength:   len(inputText),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

func (lexer *StatementsLexer) Scan() *tokens.Token {
	for {
		if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
			return tokens.NewEOFToken(lexer.tokenLocation)
		}

		startLocation := *lexer.tokenLocation
		scanLocation := *lexer.tokenLocation
		state := StatementsLexerStartState
		lastAcceptState := -1
		lastAcceptLocation := scanLocation

		for {
			if scanLocation.ByteOffset >= lexer.inputLength {
				break
			}
			r, width := lexer.peekRuneAt(scanLocation.ByteOffset)
			nextState, ok := StatementsLexerLookupTransition(state, r)
			if !ok {
				break
			}
			scanLocation.LocateRune(r, width)
			state = nextState
			if _, ok := StatementsLexerActions[state]; ok {
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
		tokenType := StatementsLexerActions[lastAcceptState]
		if StatementsLexerIsIgnoredToken(tokenType) {
			continue
		}
		return tokens.NewToken(lexeme, tokenType, &startLocation)
	}
}

func (lexer *StatementsLexer) peekRuneAt(byteOffset int) (rune, int) {
	r, width := utf8.DecodeRuneInString(lexer.inputText[byteOffset:])
	return r, width
}

func StatementsLexerLookupTransition(state int, r rune) (int, bool) {
	transitionsForState, ok := StatementsLexerTransitions[state]
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
func StatementsLexerIsIgnoredToken(tokenType tokens.TokenType) bool {
	return strings.HasPrefix(string(tokenType), "!")
}

const StatementsLexerStartState = 0

type StatementsLexerRangeTransition struct {
	from rune
	to   rune
	next int
}

var StatementsLexerTransitions = map[int][]StatementsLexerRangeTransition{
	0: {
		{from: '\t', to: '\t', next: 1},
		{from: '\n', to: '\n', next: 2},
		{from: '\r', to: '\r', next: 3},
		{from: ' ', to: ' ', next: 4},
		{from: '#', to: '#', next: 5},
		{from: '(', to: '(', next: 6},
		{from: ')', to: ')', next: 7},
		{from: '0', to: '9', next: 8},
		{from: ';', to: ';', next: 9},
		{from: '=', to: '=', next: 10},
		{from: 'A', to: 'Z', next: 11},
		{from: '_', to: '_', next: 12},
		{from: 'a', to: 'h', next: 13},
		{from: 'i', to: 'i', next: 14},
		{from: 'j', to: 'o', next: 13},
		{from: 'p', to: 'p', next: 15},
		{from: 'q', to: 'z', next: 13},
	},
	5: {
		{from: '\x00', to: '\t', next: 16},
		{from: '\n', to: '\n', next: 17},
		{from: '\v', to: '\f', next: 18},
		{from: '\x0e', to: '\U0010ffff', next: 19},
	},
	8: {
		{from: '0', to: '9', next: 20},
	},
	11: {
		{from: '0', to: '9', next: 21},
		{from: 'A', to: 'Z', next: 22},
		{from: '_', to: '_', next: 23},
		{from: 'a', to: 'z', next: 24},
	},
	12: {
		{from: '0', to: '9', next: 21},
		{from: 'A', to: 'Z', next: 22},
		{from: '_', to: '_', next: 23},
		{from: 'a', to: 'z', next: 24},
	},
	13: {
		{from: '0', to: '9', next: 21},
		{from: 'A', to: 'Z', next: 22},
		{from: '_', to: '_', next: 23},
		{from: 'a', to: 'z', next: 24},
	},
	14: {
		{from: '0', to: '9', next: 21},
		{from: 'A', to: 'Z', next: 22},
		{from: '_', to: '_', next: 23},
		{from: 'a', to: 'e', next: 24},
		{from: 'f', to: 'f', next: 25},
		{from: 'g', to: 'z', next: 24},
	},
	15: {
		{from: '0', to: '9', next: 21},
		{from: 'A', to: 'Z', next: 22},
		{from: '_', to: '_', next: 23},
		{from: 'a', to: 'q', next: 24},
		{from: 'r', to: 'r', next: 26},
		{from: 's', to: 'z', next: 24},
	},
	16: {
		{from: '\x00', to: '\t', next: 16},
		{from: '\n', to: '\n', next: 17},
		{from: '\v', to: '\f', next: 18},
		{from: '\x0e', to: '\U0010ffff', next: 19},
	},
	18: {
		{from: '\x00', to: '\t', next: 16},
		{from: '\n', to: '\n', next: 17},
		{from: '\v', to: '\f', next: 18},
		{from: '\x0e', to: '\U0010ffff', next: 19},
	},
	19: {
		{from: '\x00', to: '\t', next: 16},
		{from: '\n', to: '\n', next: 17},
		{from: '\v', to: '\f', next: 18},
		{from: '\x0e', to: '\U0010ffff', next: 19},
	},
	20: {
		{from: '0', to: '9', next: 20},
	},
	21: {
		{from: '0', to: '9', next: 21},
		{from: 'A', to: 'Z', next: 22},
		{from: '_', to: '_', next: 23},
		{from: 'a', to: 'z', next: 24},
	},
	22: {
		{from: '0', to: '9', next: 21},
		{from: 'A', to: 'Z', next: 22},
		{from: '_', to: '_', next: 23},
		{from: 'a', to: 'z', next: 24},
	},
	23: {
		{from: '0', to: '9', next: 21},
		{from: 'A', to: 'Z', next: 22},
		{from: '_', to: '_', next: 23},
		{from: 'a', to: 'z', next: 24},
	},
	24: {
		{from: '0', to: '9', next: 21},
		{from: 'A', to: 'Z', next: 22},
		{from: '_', to: '_', next: 23},
		{from: 'a', to: 'z', next: 24},
	},
	25: {
		{from: '0', to: '9', next: 21},
		{from: 'A', to: 'Z', next: 22},
		{from: '_', to: '_', next: 23},
		{from: 'a', to: 'z', next: 24},
	},
	26: {
		{from: '0', to: '9', next: 21},
		{from: 'A', to: 'Z', next: 22},
		{from: '_', to: '_', next: 23},
		{from: 'a', to: 'h', next: 24},
		{from: 'i', to: 'i', next: 27},
		{from: 'j', to: 'z', next: 24},
	},
	27: {
		{from: '0', to: '9', next: 21},
		{from: 'A', to: 'Z', next: 22},
		{from: '_', to: '_', next: 23},
		{from: 'a', to: 'm', next: 24},
		{from: 'n', to: 'n', next: 28},
		{from: 'o', to: 'z', next: 24},
	},
	28: {
		{from: '0', to: '9', next: 21},
		{from: 'A', to: 'Z', next: 22},
		{from: '_', to: '_', next: 23},
		{from: 'a', to: 's', next: 24},
		{from: 't', to: 't', next: 29},
		{from: 'u', to: 'z', next: 24},
	},
	29: {
		{from: '0', to: '9', next: 21},
		{from: 'A', to: 'Z', next: 22},
		{from: '_', to: '_', next: 23},
		{from: 'a', to: 'z', next: 24},
	},
}

var StatementsLexerActions = map[int]tokens.TokenType{
	1:  "!whitespace",
	2:  "!whitespace",
	3:  "!whitespace",
	4:  "!whitespace",
	6:  "lparen",
	7:  "rparen",
	8:  "int_literal",
	9:  "semicolon",
	10: "equals",
	11: "id",
	12: "id",
	13: "id",
	14: "id",
	15: "id",
	17: "!comment",
	20: "int_literal",
	21: "id",
	22: "id",
	23: "id",
	24: "id",
	25: "if",
	26: "id",
	27: "id",
	28: "id",
	29: "print",
}
