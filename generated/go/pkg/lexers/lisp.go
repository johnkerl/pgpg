package lexers

import (
	"fmt"
	"strings"
	"unicode/utf8"

	manuallexers "github.com/johnkerl/pgpg/manual/go/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/go/pkg/tokens"
)

type LISPLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
}

var _ manuallexers.AbstractLexer = (*LISPLexer)(nil)

func NewLISPLexer(inputText string) manuallexers.AbstractLexer {
	return &LISPLexer{
		inputText:     inputText,
		inputLength:   len(inputText),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

func (lexer *LISPLexer) Scan() *tokens.Token {
	for {
		if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
			return tokens.NewEOFToken(lexer.tokenLocation)
		}

		startLocation := *lexer.tokenLocation
		scanLocation := *lexer.tokenLocation
		state := LISPLexerStartState
		lastAcceptState := -1
		lastAcceptLocation := scanLocation

		for {
			if scanLocation.ByteOffset >= lexer.inputLength {
				break
			}
			r, width := lexer.peekRuneAt(scanLocation.ByteOffset)
			nextState, ok := LISPLexerLookupTransition(state, r)
			if !ok {
				break
			}
			scanLocation.LocateRune(r, width)
			state = nextState
			if _, ok := LISPLexerActions[state]; ok {
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
		tokenType := LISPLexerActions[lastAcceptState]
		if LISPLexerIsIgnoredToken(tokenType) {
			continue
		}
		return tokens.NewToken(lexeme, tokenType, &startLocation)
	}
}

func (lexer *LISPLexer) peekRuneAt(byteOffset int) (rune, int) {
	r, width := utf8.DecodeRuneInString(lexer.inputText[byteOffset:])
	return r, width
}

func LISPLexerLookupTransition(state int, r rune) (int, bool) {
	transitionsForState, ok := LISPLexerTransitions[state]
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
func LISPLexerIsIgnoredToken(tokenType tokens.TokenType) bool {
	return strings.HasPrefix(string(tokenType), "!")
}

const LISPLexerStartState = 0

type LISPLexerRangeTransition struct {
	from rune
	to   rune
	next int
}

var LISPLexerTransitions = map[int][]LISPLexerRangeTransition{
	0: {
		{from: '\t', to: '\t', next: 1},
		{from: '\n', to: '\n', next: 2},
		{from: '\r', to: '\r', next: 3},
		{from: ' ', to: ' ', next: 4},
		{from: '(', to: '(', next: 5},
		{from: ')', to: ')', next: 6},
		{from: '*', to: '*', next: 7},
		{from: '+', to: '+', next: 8},
		{from: '-', to: '-', next: 9},
		{from: '.', to: '.', next: 10},
		{from: '/', to: '/', next: 11},
		{from: '0', to: '9', next: 12},
		{from: ';', to: ';', next: 13},
		{from: 'A', to: 'Z', next: 14},
		{from: '_', to: '_', next: 15},
		{from: 'a', to: 'z', next: 16},
	},
	7: {
		{from: '*', to: '*', next: 17},
		{from: '+', to: '+', next: 18},
		{from: '-', to: '-', next: 19},
		{from: '.', to: '.', next: 20},
		{from: '/', to: '/', next: 21},
		{from: '0', to: '9', next: 22},
		{from: 'A', to: 'Z', next: 23},
		{from: '_', to: '_', next: 24},
		{from: 'a', to: 'z', next: 25},
	},
	8: {
		{from: '*', to: '*', next: 17},
		{from: '+', to: '+', next: 18},
		{from: '-', to: '-', next: 19},
		{from: '.', to: '.', next: 20},
		{from: '/', to: '/', next: 21},
		{from: '0', to: '9', next: 22},
		{from: 'A', to: 'Z', next: 23},
		{from: '_', to: '_', next: 24},
		{from: 'a', to: 'z', next: 25},
	},
	9: {
		{from: '*', to: '*', next: 17},
		{from: '+', to: '+', next: 18},
		{from: '-', to: '-', next: 19},
		{from: '.', to: '.', next: 20},
		{from: '/', to: '/', next: 21},
		{from: '0', to: '9', next: 22},
		{from: 'A', to: 'Z', next: 23},
		{from: '_', to: '_', next: 24},
		{from: 'a', to: 'z', next: 25},
	},
	10: {
		{from: '*', to: '*', next: 17},
		{from: '+', to: '+', next: 18},
		{from: '-', to: '-', next: 19},
		{from: '.', to: '.', next: 20},
		{from: '/', to: '/', next: 21},
		{from: '0', to: '9', next: 22},
		{from: 'A', to: 'Z', next: 23},
		{from: '_', to: '_', next: 24},
		{from: 'a', to: 'z', next: 25},
	},
	11: {
		{from: '*', to: '*', next: 17},
		{from: '+', to: '+', next: 18},
		{from: '-', to: '-', next: 19},
		{from: '.', to: '.', next: 20},
		{from: '/', to: '/', next: 21},
		{from: '0', to: '9', next: 22},
		{from: 'A', to: 'Z', next: 23},
		{from: '_', to: '_', next: 24},
		{from: 'a', to: 'z', next: 25},
	},
	12: {
		{from: '*', to: '*', next: 17},
		{from: '+', to: '+', next: 18},
		{from: '-', to: '-', next: 19},
		{from: '.', to: '.', next: 20},
		{from: '/', to: '/', next: 21},
		{from: '0', to: '9', next: 22},
		{from: 'A', to: 'Z', next: 23},
		{from: '_', to: '_', next: 24},
		{from: 'a', to: 'z', next: 25},
	},
	13: {
		{from: '\x00', to: '\t', next: 26},
		{from: '\n', to: '\n', next: 27},
		{from: '\v', to: '\f', next: 28},
		{from: '\x0e', to: '\U0010ffff', next: 29},
	},
	14: {
		{from: '*', to: '*', next: 17},
		{from: '+', to: '+', next: 18},
		{from: '-', to: '-', next: 19},
		{from: '.', to: '.', next: 20},
		{from: '/', to: '/', next: 21},
		{from: '0', to: '9', next: 22},
		{from: 'A', to: 'Z', next: 23},
		{from: '_', to: '_', next: 24},
		{from: 'a', to: 'z', next: 25},
	},
	15: {
		{from: '*', to: '*', next: 17},
		{from: '+', to: '+', next: 18},
		{from: '-', to: '-', next: 19},
		{from: '.', to: '.', next: 20},
		{from: '/', to: '/', next: 21},
		{from: '0', to: '9', next: 22},
		{from: 'A', to: 'Z', next: 23},
		{from: '_', to: '_', next: 24},
		{from: 'a', to: 'z', next: 25},
	},
	16: {
		{from: '*', to: '*', next: 17},
		{from: '+', to: '+', next: 18},
		{from: '-', to: '-', next: 19},
		{from: '.', to: '.', next: 20},
		{from: '/', to: '/', next: 21},
		{from: '0', to: '9', next: 22},
		{from: 'A', to: 'Z', next: 23},
		{from: '_', to: '_', next: 24},
		{from: 'a', to: 'z', next: 25},
	},
	17: {
		{from: '*', to: '*', next: 17},
		{from: '+', to: '+', next: 18},
		{from: '-', to: '-', next: 19},
		{from: '.', to: '.', next: 20},
		{from: '/', to: '/', next: 21},
		{from: '0', to: '9', next: 22},
		{from: 'A', to: 'Z', next: 23},
		{from: '_', to: '_', next: 24},
		{from: 'a', to: 'z', next: 25},
	},
	18: {
		{from: '*', to: '*', next: 17},
		{from: '+', to: '+', next: 18},
		{from: '-', to: '-', next: 19},
		{from: '.', to: '.', next: 20},
		{from: '/', to: '/', next: 21},
		{from: '0', to: '9', next: 22},
		{from: 'A', to: 'Z', next: 23},
		{from: '_', to: '_', next: 24},
		{from: 'a', to: 'z', next: 25},
	},
	19: {
		{from: '*', to: '*', next: 17},
		{from: '+', to: '+', next: 18},
		{from: '-', to: '-', next: 19},
		{from: '.', to: '.', next: 20},
		{from: '/', to: '/', next: 21},
		{from: '0', to: '9', next: 22},
		{from: 'A', to: 'Z', next: 23},
		{from: '_', to: '_', next: 24},
		{from: 'a', to: 'z', next: 25},
	},
	20: {
		{from: '*', to: '*', next: 17},
		{from: '+', to: '+', next: 18},
		{from: '-', to: '-', next: 19},
		{from: '.', to: '.', next: 20},
		{from: '/', to: '/', next: 21},
		{from: '0', to: '9', next: 22},
		{from: 'A', to: 'Z', next: 23},
		{from: '_', to: '_', next: 24},
		{from: 'a', to: 'z', next: 25},
	},
	21: {
		{from: '*', to: '*', next: 17},
		{from: '+', to: '+', next: 18},
		{from: '-', to: '-', next: 19},
		{from: '.', to: '.', next: 20},
		{from: '/', to: '/', next: 21},
		{from: '0', to: '9', next: 22},
		{from: 'A', to: 'Z', next: 23},
		{from: '_', to: '_', next: 24},
		{from: 'a', to: 'z', next: 25},
	},
	22: {
		{from: '*', to: '*', next: 17},
		{from: '+', to: '+', next: 18},
		{from: '-', to: '-', next: 19},
		{from: '.', to: '.', next: 20},
		{from: '/', to: '/', next: 21},
		{from: '0', to: '9', next: 22},
		{from: 'A', to: 'Z', next: 23},
		{from: '_', to: '_', next: 24},
		{from: 'a', to: 'z', next: 25},
	},
	23: {
		{from: '*', to: '*', next: 17},
		{from: '+', to: '+', next: 18},
		{from: '-', to: '-', next: 19},
		{from: '.', to: '.', next: 20},
		{from: '/', to: '/', next: 21},
		{from: '0', to: '9', next: 22},
		{from: 'A', to: 'Z', next: 23},
		{from: '_', to: '_', next: 24},
		{from: 'a', to: 'z', next: 25},
	},
	24: {
		{from: '*', to: '*', next: 17},
		{from: '+', to: '+', next: 18},
		{from: '-', to: '-', next: 19},
		{from: '.', to: '.', next: 20},
		{from: '/', to: '/', next: 21},
		{from: '0', to: '9', next: 22},
		{from: 'A', to: 'Z', next: 23},
		{from: '_', to: '_', next: 24},
		{from: 'a', to: 'z', next: 25},
	},
	25: {
		{from: '*', to: '*', next: 17},
		{from: '+', to: '+', next: 18},
		{from: '-', to: '-', next: 19},
		{from: '.', to: '.', next: 20},
		{from: '/', to: '/', next: 21},
		{from: '0', to: '9', next: 22},
		{from: 'A', to: 'Z', next: 23},
		{from: '_', to: '_', next: 24},
		{from: 'a', to: 'z', next: 25},
	},
	26: {
		{from: '\x00', to: '\t', next: 26},
		{from: '\n', to: '\n', next: 27},
		{from: '\v', to: '\f', next: 28},
		{from: '\x0e', to: '\U0010ffff', next: 29},
	},
	28: {
		{from: '\x00', to: '\t', next: 26},
		{from: '\n', to: '\n', next: 27},
		{from: '\v', to: '\f', next: 28},
		{from: '\x0e', to: '\U0010ffff', next: 29},
	},
	29: {
		{from: '\x00', to: '\t', next: 26},
		{from: '\n', to: '\n', next: 27},
		{from: '\v', to: '\f', next: 28},
		{from: '\x0e', to: '\U0010ffff', next: 29},
	},
}

var LISPLexerActions = map[int]tokens.TokenType{
	1:  "!whitespace",
	2:  "!whitespace",
	3:  "!whitespace",
	4:  "!whitespace",
	5:  "lparen",
	6:  "rparen",
	7:  "identifier",
	8:  "identifier",
	9:  "identifier",
	10: "identifier",
	11: "identifier",
	12: "identifier",
	13: "!comment",
	14: "identifier",
	15: "identifier",
	16: "identifier",
	17: "identifier",
	18: "identifier",
	19: "identifier",
	20: "identifier",
	21: "identifier",
	22: "identifier",
	23: "identifier",
	24: "identifier",
	25: "identifier",
	26: "!comment",
	27: "!comment",
	28: "!comment",
	29: "!comment",
}
