package lexers

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	liblexers "github.com/johnkerl/pgpg/go/lib/pkg/lexers"
	"github.com/johnkerl/pgpg/go/lib/pkg/tokens"
)

const JSONLexerBufSize = 4096

type JSONLexer struct {
	reader        *bufio.Reader
	buf           []byte
	tokenStart    int
	tokenLocation *tokens.TokenLocation
	atEOF         bool
}

var _ liblexers.AbstractLexer = (*JSONLexer)(nil)

func NewJSONLexer(r io.Reader) liblexers.AbstractLexer {
	reader, ok := r.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(r)
	}
	return &JSONLexer{
		reader:        reader,
		buf:           make([]byte, 0, JSONLexerBufSize),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

// NewJSONLexerFromString returns a lexer over s (convenience for tests and -e mode).
func NewJSONLexerFromString(s string) liblexers.AbstractLexer {
	return NewJSONLexer(strings.NewReader(s))
}

func (lexer *JSONLexer) ensureFill(needBytes int) {
	for needBytes > len(lexer.buf) && !lexer.atEOF {
		chunk := make([]byte, JSONLexerBufSize)
		n, err := lexer.reader.Read(chunk)
		if n > 0 {
			lexer.buf = append(lexer.buf, chunk[:n]...)
		}
		if err == io.EOF {
			lexer.atEOF = true
			return
		}
		if err != nil {
			lexer.atEOF = true
			return
		}
	}
}

func (lexer *JSONLexer) peekRuneAt(byteOffset int) (rune, int) {
	lexer.ensureFill(byteOffset + utf8.UTFMax)
	if byteOffset >= len(lexer.buf) {
		return 0, 0
	}
	r, width := utf8.DecodeRune(lexer.buf[byteOffset:])
	if width == 0 {
		return 0, 0
	}
	return r, width
}

func (lexer *JSONLexer) Scan() *tokens.Token {
	lexer.ensureFill(lexer.tokenStart + 1)
	if lexer.tokenStart >= len(lexer.buf) && lexer.atEOF {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	for {
		if lexer.tokenStart >= len(lexer.buf) {
			if lexer.atEOF {
				return tokens.NewEOFToken(lexer.tokenLocation)
			}
			lexer.ensureFill(lexer.tokenStart + 1)
			if lexer.tokenStart >= len(lexer.buf) {
				return tokens.NewEOFToken(lexer.tokenLocation)
			}
		}

		startLocation := *lexer.tokenLocation
		scanOffset := lexer.tokenStart
		state := JSONLexerStartState
		lastAcceptState := -1
		lastAcceptOffset := scanOffset

		for {
			if scanOffset >= len(lexer.buf) {
				if !lexer.atEOF {
					lexer.ensureFill(scanOffset + utf8.UTFMax)
				}
				if scanOffset >= len(lexer.buf) {
					break
				}
			}
			r, width := lexer.peekRuneAt(scanOffset)
			if width == 0 {
				break
			}
			nextState, ok := JSONLexerLookupTransition(state, r)
			if !ok {
				break
			}
			scanOffset += width
			state = nextState
			if _, ok := JSONLexerActions[state]; ok {
				lastAcceptState = state
				lastAcceptOffset = scanOffset
			}
		}

		if lastAcceptState < 0 {
			r, _ := lexer.peekRuneAt(lexer.tokenStart)
			return tokens.NewErrorToken(fmt.Sprintf("lexer: unrecognized input %q", r), lexer.tokenLocation)
		}

		lexemeText := string(lexer.buf[lexer.tokenStart:lastAcceptOffset])
		lexeme := []rune(lexemeText)
		for len(lexemeText) > 0 {
			r, w := utf8.DecodeRuneInString(lexemeText)
			lexer.tokenLocation.LocateRune(r, w)
			lexemeText = lexemeText[w:]
		}
		lexer.buf = lexer.buf[lastAcceptOffset:]
		lexer.tokenStart = 0
		tokenType := JSONLexerActions[lastAcceptState]
		if JSONLexerIsIgnoredToken(tokenType) {
			continue
		}
		return tokens.NewToken(lexeme, tokenType, &startLocation)
	}
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
		{from: '-', to: '-', next: 7},
		{from: '0', to: '0', next: 8},
		{from: '1', to: '9', next: 9},
		{from: ':', to: ':', next: 10},
		{from: '[', to: '[', next: 11},
		{from: ']', to: ']', next: 12},
		{from: 'f', to: 'f', next: 13},
		{from: 'n', to: 'n', next: 14},
		{from: 't', to: 't', next: 15},
		{from: '{', to: '{', next: 16},
		{from: '}', to: '}', next: 17},
	},
	5: {
		{from: ' ', to: '!', next: 18},
		{from: '"', to: '"', next: 19},
		{from: '#', to: '[', next: 20},
		{from: '\\', to: '\\', next: 21},
		{from: ']', to: '\uffff', next: 22},
	},
	7: {
		{from: '0', to: '0', next: 8},
		{from: '1', to: '9', next: 9},
	},
	8: {
		{from: '.', to: '.', next: 23},
		{from: 'E', to: 'E', next: 24},
		{from: 'e', to: 'e', next: 25},
	},
	9: {
		{from: '.', to: '.', next: 23},
		{from: '0', to: '9', next: 26},
		{from: 'E', to: 'E', next: 24},
		{from: 'e', to: 'e', next: 25},
	},
	13: {
		{from: 'a', to: 'a', next: 27},
	},
	14: {
		{from: 'u', to: 'u', next: 28},
	},
	15: {
		{from: 'r', to: 'r', next: 29},
	},
	18: {
		{from: ' ', to: '!', next: 18},
		{from: '"', to: '"', next: 19},
		{from: '#', to: '[', next: 20},
		{from: '\\', to: '\\', next: 21},
		{from: ']', to: '\uffff', next: 22},
	},
	20: {
		{from: ' ', to: '!', next: 18},
		{from: '"', to: '"', next: 19},
		{from: '#', to: '[', next: 20},
		{from: '\\', to: '\\', next: 21},
		{from: ']', to: '\uffff', next: 22},
	},
	21: {
		{from: '"', to: '"', next: 30},
		{from: '/', to: '/', next: 31},
		{from: '\\', to: '\\', next: 32},
		{from: 'b', to: 'b', next: 33},
		{from: 'f', to: 'f', next: 34},
		{from: 'n', to: 'n', next: 35},
		{from: 'r', to: 'r', next: 36},
		{from: 't', to: 't', next: 37},
		{from: 'u', to: 'u', next: 38},
	},
	22: {
		{from: ' ', to: '!', next: 18},
		{from: '"', to: '"', next: 19},
		{from: '#', to: '[', next: 20},
		{from: '\\', to: '\\', next: 21},
		{from: ']', to: '\uffff', next: 22},
	},
	23: {
		{from: '0', to: '9', next: 39},
	},
	24: {
		{from: '+', to: '+', next: 40},
		{from: '-', to: '-', next: 41},
		{from: '0', to: '9', next: 42},
	},
	25: {
		{from: '+', to: '+', next: 40},
		{from: '-', to: '-', next: 41},
		{from: '0', to: '9', next: 42},
	},
	26: {
		{from: '.', to: '.', next: 23},
		{from: '0', to: '9', next: 26},
		{from: 'E', to: 'E', next: 24},
		{from: 'e', to: 'e', next: 25},
	},
	27: {
		{from: 'l', to: 'l', next: 43},
	},
	28: {
		{from: 'l', to: 'l', next: 44},
	},
	29: {
		{from: 'u', to: 'u', next: 45},
	},
	30: {
		{from: ' ', to: '!', next: 18},
		{from: '"', to: '"', next: 19},
		{from: '#', to: '[', next: 20},
		{from: '\\', to: '\\', next: 21},
		{from: ']', to: '\uffff', next: 22},
	},
	31: {
		{from: ' ', to: '!', next: 18},
		{from: '"', to: '"', next: 19},
		{from: '#', to: '[', next: 20},
		{from: '\\', to: '\\', next: 21},
		{from: ']', to: '\uffff', next: 22},
	},
	32: {
		{from: ' ', to: '!', next: 18},
		{from: '"', to: '"', next: 19},
		{from: '#', to: '[', next: 20},
		{from: '\\', to: '\\', next: 21},
		{from: ']', to: '\uffff', next: 22},
	},
	33: {
		{from: ' ', to: '!', next: 18},
		{from: '"', to: '"', next: 19},
		{from: '#', to: '[', next: 20},
		{from: '\\', to: '\\', next: 21},
		{from: ']', to: '\uffff', next: 22},
	},
	34: {
		{from: ' ', to: '!', next: 18},
		{from: '"', to: '"', next: 19},
		{from: '#', to: '[', next: 20},
		{from: '\\', to: '\\', next: 21},
		{from: ']', to: '\uffff', next: 22},
	},
	35: {
		{from: ' ', to: '!', next: 18},
		{from: '"', to: '"', next: 19},
		{from: '#', to: '[', next: 20},
		{from: '\\', to: '\\', next: 21},
		{from: ']', to: '\uffff', next: 22},
	},
	36: {
		{from: ' ', to: '!', next: 18},
		{from: '"', to: '"', next: 19},
		{from: '#', to: '[', next: 20},
		{from: '\\', to: '\\', next: 21},
		{from: ']', to: '\uffff', next: 22},
	},
	37: {
		{from: ' ', to: '!', next: 18},
		{from: '"', to: '"', next: 19},
		{from: '#', to: '[', next: 20},
		{from: '\\', to: '\\', next: 21},
		{from: ']', to: '\uffff', next: 22},
	},
	38: {
		{from: '0', to: '9', next: 46},
		{from: 'A', to: 'F', next: 47},
		{from: 'a', to: 'f', next: 48},
	},
	39: {
		{from: '0', to: '9', next: 49},
		{from: 'E', to: 'E', next: 24},
		{from: 'e', to: 'e', next: 25},
	},
	40: {
		{from: '0', to: '9', next: 42},
	},
	41: {
		{from: '0', to: '9', next: 42},
	},
	42: {
		{from: '0', to: '9', next: 50},
	},
	43: {
		{from: 's', to: 's', next: 51},
	},
	44: {
		{from: 'l', to: 'l', next: 52},
	},
	45: {
		{from: 'e', to: 'e', next: 53},
	},
	46: {
		{from: '0', to: '9', next: 54},
		{from: 'A', to: 'F', next: 55},
		{from: 'a', to: 'f', next: 56},
	},
	47: {
		{from: '0', to: '9', next: 54},
		{from: 'A', to: 'F', next: 55},
		{from: 'a', to: 'f', next: 56},
	},
	48: {
		{from: '0', to: '9', next: 54},
		{from: 'A', to: 'F', next: 55},
		{from: 'a', to: 'f', next: 56},
	},
	49: {
		{from: '0', to: '9', next: 49},
		{from: 'E', to: 'E', next: 24},
		{from: 'e', to: 'e', next: 25},
	},
	50: {
		{from: '0', to: '9', next: 50},
	},
	51: {
		{from: 'e', to: 'e', next: 57},
	},
	54: {
		{from: '0', to: '9', next: 58},
		{from: 'A', to: 'F', next: 59},
		{from: 'a', to: 'f', next: 60},
	},
	55: {
		{from: '0', to: '9', next: 58},
		{from: 'A', to: 'F', next: 59},
		{from: 'a', to: 'f', next: 60},
	},
	56: {
		{from: '0', to: '9', next: 58},
		{from: 'A', to: 'F', next: 59},
		{from: 'a', to: 'f', next: 60},
	},
	58: {
		{from: '0', to: '9', next: 61},
		{from: 'A', to: 'F', next: 62},
		{from: 'a', to: 'f', next: 63},
	},
	59: {
		{from: '0', to: '9', next: 61},
		{from: 'A', to: 'F', next: 62},
		{from: 'a', to: 'f', next: 63},
	},
	60: {
		{from: '0', to: '9', next: 61},
		{from: 'A', to: 'F', next: 62},
		{from: 'a', to: 'f', next: 63},
	},
	61: {
		{from: ' ', to: '!', next: 18},
		{from: '"', to: '"', next: 19},
		{from: '#', to: '[', next: 20},
		{from: '\\', to: '\\', next: 21},
		{from: ']', to: '\uffff', next: 22},
	},
	62: {
		{from: ' ', to: '!', next: 18},
		{from: '"', to: '"', next: 19},
		{from: '#', to: '[', next: 20},
		{from: '\\', to: '\\', next: 21},
		{from: ']', to: '\uffff', next: 22},
	},
	63: {
		{from: ' ', to: '!', next: 18},
		{from: '"', to: '"', next: 19},
		{from: '#', to: '[', next: 20},
		{from: '\\', to: '\\', next: 21},
		{from: ']', to: '\uffff', next: 22},
	},
}

var JSONLexerActions = map[int]tokens.TokenType{
	1:  "!whitespace",
	2:  "!whitespace",
	3:  "!whitespace",
	4:  "!whitespace",
	6:  "comma",
	8:  "number",
	9:  "number",
	10: "colon",
	11: "lbracket",
	12: "rbracket",
	16: "lcurly",
	17: "rcurly",
	19: "string",
	26: "number",
	39: "number",
	42: "number",
	49: "number",
	50: "number",
	52: "null",
	53: "true",
	57: "false",
}
