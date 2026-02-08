package lexers

import (
	"fmt"
	"strings"
	"unicode/utf8"

	manuallexers "github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type SENGLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
}

var _ manuallexers.AbstractLexer = (*SENGLexer)(nil)

func NewSENGLexer(inputText string) manuallexers.AbstractLexer {
	return &SENGLexer{
		inputText:     inputText,
		inputLength:   len(inputText),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

func (lexer *SENGLexer) Scan() *tokens.Token {
	for {
		if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
			return tokens.NewEOFToken(lexer.tokenLocation)
		}

		startLocation := *lexer.tokenLocation
		scanLocation := *lexer.tokenLocation
		state := SENGLexerStartState
		lastAcceptState := -1
		lastAcceptLocation := scanLocation

		for {
			if scanLocation.ByteOffset >= lexer.inputLength {
				break
			}
			r, width := lexer.peekRuneAt(scanLocation.ByteOffset)
			nextState, ok := SENGLexerLookupTransition(state, r)
			if !ok {
				break
			}
			scanLocation.LocateRune(r, width)
			state = nextState
			if _, ok := SENGLexerActions[state]; ok {
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
		tokenType := SENGLexerActions[lastAcceptState]
		if SENGLexerIsIgnoredToken(tokenType) {
			continue
		}
		return tokens.NewToken(lexeme, tokenType, &startLocation)
	}
}

func (lexer *SENGLexer) peekRuneAt(byteOffset int) (rune, int) {
	r, width := utf8.DecodeRuneInString(lexer.inputText[byteOffset:])
	return r, width
}

func SENGLexerLookupTransition(state int, r rune) (int, bool) {
	transitionsForState, ok := SENGLexerTransitions[state]
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
func SENGLexerIsIgnoredToken(tokenType tokens.TokenType) bool {
	return strings.HasPrefix(string(tokenType), "!")
}

const SENGLexerStartState = 0

type SENGLexerRangeTransition struct {
	from rune
	to   rune
	next int
}

var SENGLexerTransitions = map[int][]SENGLexerRangeTransition{
	0: {
		{from: '\t', to: '\t', next: 1},
		{from: '\n', to: '\n', next: 2},
		{from: '\r', to: '\r', next: 3},
		{from: ' ', to: ' ', next: 4},
		{from: 'a', to: 'a', next: 5},
		{from: 'b', to: 'b', next: 6},
		{from: 'c', to: 'c', next: 7},
		{from: 'd', to: 'd', next: 8},
		{from: 'e', to: 'e', next: 9},
		{from: 'f', to: 'f', next: 10},
		{from: 'g', to: 'g', next: 11},
		{from: 'j', to: 'j', next: 12},
		{from: 'l', to: 'l', next: 13},
		{from: 'm', to: 'm', next: 14},
		{from: 'o', to: 'o', next: 15},
		{from: 'p', to: 'p', next: 16},
		{from: 'q', to: 'q', next: 17},
		{from: 'r', to: 'r', next: 18},
		{from: 's', to: 's', next: 19},
		{from: 't', to: 't', next: 20},
		{from: 'u', to: 'u', next: 21},
		{from: 'w', to: 'w', next: 22},
	},
	6: {
		{from: 'o', to: 'o', next: 23},
		{from: 'r', to: 'r', next: 24},
	},
	7: {
		{from: 'a', to: 'a', next: 25},
	},
	8: {
		{from: 'o', to: 'o', next: 26},
	},
	9: {
		{from: 'a', to: 'a', next: 27},
	},
	10: {
		{from: 'o', to: 'o', next: 28},
	},
	11: {
		{from: 'o', to: 'o', next: 29},
		{from: 'r', to: 'r', next: 30},
	},
	12: {
		{from: 'u', to: 'u', next: 31},
	},
	13: {
		{from: 'a', to: 'a', next: 32},
	},
	14: {
		{from: 'o', to: 'o', next: 33},
	},
	15: {
		{from: 'v', to: 'v', next: 34},
	},
	16: {
		{from: 'u', to: 'u', next: 35},
	},
	17: {
		{from: 'u', to: 'u', next: 36},
	},
	18: {
		{from: 'e', to: 'e', next: 37},
		{from: 'u', to: 'u', next: 38},
	},
	19: {
		{from: 'l', to: 'l', next: 39},
	},
	20: {
		{from: 'h', to: 'h', next: 40},
	},
	21: {
		{from: 'n', to: 'n', next: 41},
	},
	22: {
		{from: 'a', to: 'a', next: 42},
	},
	23: {
		{from: 'o', to: 'o', next: 43},
	},
	24: {
		{from: 'o', to: 'o', next: 44},
	},
	25: {
		{from: 't', to: 't', next: 45},
	},
	26: {
		{from: 'g', to: 'g', next: 46},
	},
	27: {
		{from: 't', to: 't', next: 47},
	},
	28: {
		{from: 'o', to: 'o', next: 48},
		{from: 'x', to: 'x', next: 49},
	},
	29: {
		{from: 'e', to: 'e', next: 50},
	},
	30: {
		{from: 'e', to: 'e', next: 51},
	},
	31: {
		{from: 'm', to: 'm', next: 52},
	},
	32: {
		{from: 'z', to: 'z', next: 53},
	},
	33: {
		{from: 'u', to: 'u', next: 54},
	},
	34: {
		{from: 'e', to: 'e', next: 55},
	},
	35: {
		{from: 't', to: 't', next: 56},
	},
	36: {
		{from: 'i', to: 'i', next: 57},
	},
	37: {
		{from: 'a', to: 'a', next: 58},
		{from: 'd', to: 'd', next: 59},
	},
	38: {
		{from: 'n', to: 'n', next: 60},
	},
	39: {
		{from: 'e', to: 'e', next: 61},
		{from: 'o', to: 'o', next: 62},
	},
	40: {
		{from: 'e', to: 'e', next: 63},
	},
	41: {
		{from: 'd', to: 'd', next: 64},
	},
	42: {
		{from: 'l', to: 'l', next: 65},
	},
	43: {
		{from: 'k', to: 'k', next: 66},
	},
	44: {
		{from: 'w', to: 'w', next: 67},
	},
	47: {
		{from: 's', to: 's', next: 68},
	},
	48: {
		{from: 'd', to: 'd', next: 69},
	},
	50: {
		{from: 's', to: 's', next: 70},
	},
	51: {
		{from: 'e', to: 'e', next: 71},
	},
	52: {
		{from: 'p', to: 'p', next: 72},
	},
	53: {
		{from: 'y', to: 'y', next: 73},
	},
	54: {
		{from: 's', to: 's', next: 74},
	},
	55: {
		{from: 'r', to: 'r', next: 75},
	},
	56: {
		{from: 's', to: 's', next: 76},
	},
	57: {
		{from: 'c', to: 'c', next: 77},
	},
	58: {
		{from: 'd', to: 'd', next: 78},
	},
	60: {
		{from: 's', to: 's', next: 79},
	},
	61: {
		{from: 'e', to: 'e', next: 80},
	},
	62: {
		{from: 'w', to: 'w', next: 81},
	},
	64: {
		{from: 'e', to: 'e', next: 82},
	},
	65: {
		{from: 'k', to: 'k', next: 83},
	},
	67: {
		{from: 'n', to: 'n', next: 84},
	},
	71: {
		{from: 'n', to: 'n', next: 85},
	},
	72: {
		{from: 's', to: 's', next: 86},
	},
	74: {
		{from: 'e', to: 'e', next: 87},
	},
	77: {
		{from: 'k', to: 'k', next: 88},
	},
	80: {
		{from: 'p', to: 'p', next: 89},
	},
	81: {
		{from: 'l', to: 'l', next: 90},
	},
	82: {
		{from: 'r', to: 'r', next: 91},
	},
	83: {
		{from: 's', to: 's', next: 92},
	},
	88: {
		{from: 'l', to: 'l', next: 93},
	},
	89: {
		{from: 's', to: 's', next: 94},
	},
	90: {
		{from: 'y', to: 'y', next: 95},
	},
	93: {
		{from: 'y', to: 'y', next: 96},
	},
}

var SENGLexerActions = map[int]tokens.TokenType{
	1:  "!whitespace",
	2:  "!whitespace",
	3:  "!whitespace",
	4:  "!whitespace",
	5:  "article",
	29: "intransitiveImperativeVerb",
	45: "noun",
	46: "noun",
	47: "transitiveImperativeVerb",
	49: "noun",
	56: "transitiveImperativeVerb",
	59: "adjective",
	63: "article",
	66: "noun",
	68: "transitiveVerb",
	69: "noun",
	70: "intransitiveVerb",
	72: "intransitiveImperativeVerb",
	73: "adjective",
	75: "preposition",
	76: "transitiveVerb",
	78: "transitiveImperativeVerb",
	79: "intransitiveVerb",
	84: "adjective",
	85: "adjective",
	86: "intransitiveVerb",
	87: "noun",
	88: "adjective",
	91: "preposition",
	92: "intransitiveVerb",
	94: "intransitiveVerb",
	95: "adverb",
	96: "adverb",
}
