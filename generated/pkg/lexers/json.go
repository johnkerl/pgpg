package lexers

import (
	"fmt"
	"strings"
	"unicode/utf8"

	manuallexers "github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type JSONLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
}

var _ manuallexers.AbstractLexer = (*JSONLexer)(nil)

func NewJSONLexer(inputText string) manuallexers.AbstractLexer {
	return &JSONLexer{
		inputText:     inputText,
		inputLength:   len(inputText),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

func (lexer *JSONLexer) Scan() *tokens.Token {
	for {
		if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
			return tokens.NewEOFToken(lexer.tokenLocation)
		}

		startLocation := *lexer.tokenLocation
		scanLocation := *lexer.tokenLocation
		state := JSONLexerStartState
		lastAcceptState := -1
		lastAcceptLocation := scanLocation

		for {
			if scanLocation.ByteOffset >= lexer.inputLength {
				break
			}
			r, width := lexer.peekRuneAt(scanLocation.ByteOffset)
			nextState, ok := JSONLexerLookupTransition(state, r)
			if !ok {
				break
			}
			scanLocation.LocateRune(r, width)
			state = nextState
			if _, ok := JSONLexerActions[state]; ok {
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
		tokenType := JSONLexerActions[lastAcceptState]
		if JSONLexerIsIgnoredToken(tokenType) {
			continue
		}
		return tokens.NewToken(lexeme, tokenType, &startLocation)
	}
}

func (lexer *JSONLexer) peekRuneAt(byteOffset int) (rune, int) {
	r, width := utf8.DecodeRuneInString(lexer.inputText[byteOffset:])
	return r, width
}

func JSONLexerLookupTransition(state int, r rune) (int, bool) {
	transitionsForState, ok := JSONLexerTransitions[state]
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
func JSONLexerIsIgnoredToken(tokenType tokens.TokenType) bool {
	return strings.HasPrefix(string(tokenType), "!")
}

const JSONLexerStartState = 0

type JSONLexerRangeTransition struct {
	from rune
	to   rune
	next int
}

var JSONLexerTransitions = map[int][]JSONLexerRangeTransition{
	0: {
		{from: '\t', to: '\t', next: 1},
		{from: '\n', to: '\n', next: 2},
		{from: '\r', to: '\r', next: 3},
		{from: ' ', to: ' ', next: 4},
		{from: '"', to: '"', next: 5},
		{from: ',', to: ',', next: 6},
		{from: '0', to: '9', next: 7},
		{from: ':', to: ':', next: 8},
		{from: '[', to: '[', next: 9},
		{from: ']', to: ']', next: 10},
		{from: 'f', to: 'f', next: 11},
		{from: 'n', to: 'n', next: 12},
		{from: 't', to: 't', next: 13},
	},
	5: {
		{from: '\x00', to: '\t', next: 14},
		{from: '\v', to: '\f', next: 15},
		{from: '\x0e', to: '!', next: 16},
		{from: '"', to: '"', next: 17},
		{from: '#', to: '\U0010ffff', next: 16},
	},
	7: {
		{from: '0', to: '9', next: 18},
	},
	11: {
		{from: 'a', to: 'a', next: 19},
	},
	12: {
		{from: 'u', to: 'u', next: 20},
	},
	13: {
		{from: 'r', to: 'r', next: 21},
	},
	14: {
		{from: '\x00', to: '\t', next: 14},
		{from: '\v', to: '\f', next: 15},
		{from: '\x0e', to: '!', next: 16},
		{from: '"', to: '"', next: 17},
		{from: '#', to: '\U0010ffff', next: 16},
	},
	15: {
		{from: '\x00', to: '\t', next: 14},
		{from: '\v', to: '\f', next: 15},
		{from: '\x0e', to: '!', next: 16},
		{from: '"', to: '"', next: 17},
		{from: '#', to: '\U0010ffff', next: 16},
	},
	16: {
		{from: '\x00', to: '\t', next: 14},
		{from: '\v', to: '\f', next: 15},
		{from: '\x0e', to: '!', next: 16},
		{from: '"', to: '"', next: 17},
		{from: '#', to: '\U0010ffff', next: 16},
	},
	17: {
		{from: '\x00', to: '\t', next: 14},
		{from: '\v', to: '\f', next: 15},
		{from: '\x0e', to: '!', next: 16},
		{from: '"', to: '"', next: 17},
		{from: '#', to: '\U0010ffff', next: 16},
	},
	18: {
		{from: '0', to: '9', next: 18},
	},
	19: {
		{from: 'l', to: 'l', next: 22},
	},
	20: {
		{from: 'l', to: 'l', next: 23},
	},
	21: {
		{from: 'u', to: 'u', next: 24},
	},
	22: {
		{from: 's', to: 's', next: 25},
	},
	23: {
		{from: 'l', to: 'l', next: 26},
	},
	24: {
		{from: 'e', to: 'e', next: 27},
	},
	25: {
		{from: 'e', to: 'e', next: 28},
	},
}

var JSONLexerActions = map[int]tokens.TokenType{
	1:  "!whitespace",
	2:  "!whitespace",
	3:  "!whitespace",
	4:  "!whitespace",
	6:  "comma",
	7:  "number",
	8:  "colon",
	9:  "lcurly",
	10: "rcurly",
	17: "string",
	18: "number",
	26: "null",
	27: "true",
	28: "false",
}
