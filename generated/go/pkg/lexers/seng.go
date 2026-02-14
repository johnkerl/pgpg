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
		{from: '#', to: '#', next: 5},
		{from: 'a', to: 'a', next: 6},
		{from: 'b', to: 'b', next: 7},
		{from: 'c', to: 'c', next: 8},
		{from: 'd', to: 'd', next: 9},
		{from: 'e', to: 'e', next: 10},
		{from: 'f', to: 'f', next: 11},
		{from: 'g', to: 'g', next: 12},
		{from: 'j', to: 'j', next: 13},
		{from: 'l', to: 'l', next: 14},
		{from: 'm', to: 'm', next: 15},
		{from: 'o', to: 'o', next: 16},
		{from: 'p', to: 'p', next: 17},
		{from: 'q', to: 'q', next: 18},
		{from: 'r', to: 'r', next: 19},
		{from: 's', to: 's', next: 20},
		{from: 't', to: 't', next: 21},
		{from: 'u', to: 'u', next: 22},
		{from: 'w', to: 'w', next: 23},
	},
	5: {
		{from: '\x00', to: '\t', next: 24},
		{from: '\n', to: '\n', next: 25},
		{from: '\v', to: '\f', next: 26},
		{from: '\x0e', to: '\U0010ffff', next: 27},
	},
	7: {
		{from: 'o', to: 'o', next: 28},
		{from: 'r', to: 'r', next: 29},
	},
	8: {
		{from: 'a', to: 'a', next: 30},
	},
	9: {
		{from: 'o', to: 'o', next: 31},
	},
	10: {
		{from: 'a', to: 'a', next: 32},
	},
	11: {
		{from: 'o', to: 'o', next: 33},
	},
	12: {
		{from: 'o', to: 'o', next: 34},
		{from: 'r', to: 'r', next: 35},
	},
	13: {
		{from: 'u', to: 'u', next: 36},
	},
	14: {
		{from: 'a', to: 'a', next: 37},
	},
	15: {
		{from: 'o', to: 'o', next: 38},
	},
	16: {
		{from: 'v', to: 'v', next: 39},
	},
	17: {
		{from: 'u', to: 'u', next: 40},
	},
	18: {
		{from: 'u', to: 'u', next: 41},
	},
	19: {
		{from: 'e', to: 'e', next: 42},
		{from: 'u', to: 'u', next: 43},
	},
	20: {
		{from: 'l', to: 'l', next: 44},
	},
	21: {
		{from: 'h', to: 'h', next: 45},
	},
	22: {
		{from: 'n', to: 'n', next: 46},
	},
	23: {
		{from: 'a', to: 'a', next: 47},
	},
	24: {
		{from: '\x00', to: '\t', next: 24},
		{from: '\n', to: '\n', next: 25},
		{from: '\v', to: '\f', next: 26},
		{from: '\x0e', to: '\U0010ffff', next: 27},
	},
	26: {
		{from: '\x00', to: '\t', next: 24},
		{from: '\n', to: '\n', next: 25},
		{from: '\v', to: '\f', next: 26},
		{from: '\x0e', to: '\U0010ffff', next: 27},
	},
	27: {
		{from: '\x00', to: '\t', next: 24},
		{from: '\n', to: '\n', next: 25},
		{from: '\v', to: '\f', next: 26},
		{from: '\x0e', to: '\U0010ffff', next: 27},
	},
	28: {
		{from: 'o', to: 'o', next: 48},
	},
	29: {
		{from: 'o', to: 'o', next: 49},
	},
	30: {
		{from: 't', to: 't', next: 50},
	},
	31: {
		{from: 'g', to: 'g', next: 51},
	},
	32: {
		{from: 't', to: 't', next: 52},
	},
	33: {
		{from: 'o', to: 'o', next: 53},
		{from: 'x', to: 'x', next: 54},
	},
	34: {
		{from: 'e', to: 'e', next: 55},
	},
	35: {
		{from: 'e', to: 'e', next: 56},
	},
	36: {
		{from: 'm', to: 'm', next: 57},
	},
	37: {
		{from: 'z', to: 'z', next: 58},
	},
	38: {
		{from: 'u', to: 'u', next: 59},
	},
	39: {
		{from: 'e', to: 'e', next: 60},
	},
	40: {
		{from: 't', to: 't', next: 61},
	},
	41: {
		{from: 'i', to: 'i', next: 62},
	},
	42: {
		{from: 'a', to: 'a', next: 63},
		{from: 'd', to: 'd', next: 64},
	},
	43: {
		{from: 'n', to: 'n', next: 65},
	},
	44: {
		{from: 'e', to: 'e', next: 66},
		{from: 'o', to: 'o', next: 67},
	},
	45: {
		{from: 'e', to: 'e', next: 68},
	},
	46: {
		{from: 'd', to: 'd', next: 69},
	},
	47: {
		{from: 'l', to: 'l', next: 70},
	},
	48: {
		{from: 'k', to: 'k', next: 71},
	},
	49: {
		{from: 'w', to: 'w', next: 72},
	},
	52: {
		{from: 's', to: 's', next: 73},
	},
	53: {
		{from: 'd', to: 'd', next: 74},
	},
	55: {
		{from: 's', to: 's', next: 75},
	},
	56: {
		{from: 'e', to: 'e', next: 76},
	},
	57: {
		{from: 'p', to: 'p', next: 77},
	},
	58: {
		{from: 'y', to: 'y', next: 78},
	},
	59: {
		{from: 's', to: 's', next: 79},
	},
	60: {
		{from: 'r', to: 'r', next: 80},
	},
	61: {
		{from: 's', to: 's', next: 81},
	},
	62: {
		{from: 'c', to: 'c', next: 82},
	},
	63: {
		{from: 'd', to: 'd', next: 83},
	},
	65: {
		{from: 's', to: 's', next: 84},
	},
	66: {
		{from: 'e', to: 'e', next: 85},
	},
	67: {
		{from: 'w', to: 'w', next: 86},
	},
	69: {
		{from: 'e', to: 'e', next: 87},
	},
	70: {
		{from: 'k', to: 'k', next: 88},
	},
	72: {
		{from: 'n', to: 'n', next: 89},
	},
	76: {
		{from: 'n', to: 'n', next: 90},
	},
	77: {
		{from: 's', to: 's', next: 91},
	},
	79: {
		{from: 'e', to: 'e', next: 92},
	},
	82: {
		{from: 'k', to: 'k', next: 93},
	},
	85: {
		{from: 'p', to: 'p', next: 94},
	},
	86: {
		{from: 'l', to: 'l', next: 95},
	},
	87: {
		{from: 'r', to: 'r', next: 96},
	},
	88: {
		{from: 's', to: 's', next: 97},
	},
	93: {
		{from: 'l', to: 'l', next: 98},
	},
	94: {
		{from: 's', to: 's', next: 99},
	},
	95: {
		{from: 'y', to: 'y', next: 100},
	},
	98: {
		{from: 'y', to: 'y', next: 101},
	},
}

var SENGLexerActions = map[int]tokens.TokenType{
	1:   "!whitespace",
	2:   "!whitespace",
	3:   "!whitespace",
	4:   "!whitespace",
	6:   "article",
	25:  "!comment",
	34:  "intransitiveImperativeVerb",
	50:  "noun",
	51:  "noun",
	52:  "transitiveImperativeVerb",
	54:  "noun",
	61:  "transitiveImperativeVerb",
	64:  "adjective",
	68:  "article",
	71:  "noun",
	73:  "transitiveVerb",
	74:  "noun",
	75:  "intransitiveVerb",
	77:  "intransitiveImperativeVerb",
	78:  "adjective",
	80:  "preposition",
	81:  "transitiveVerb",
	83:  "transitiveImperativeVerb",
	84:  "intransitiveVerb",
	89:  "adjective",
	90:  "adjective",
	91:  "intransitiveVerb",
	92:  "noun",
	93:  "adjective",
	96:  "preposition",
	97:  "intransitiveVerb",
	99:  "intransitiveVerb",
	100: "adverb",
	101: "adverb",
}
