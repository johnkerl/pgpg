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

const PEMDASIntLexerBufSize = 4096

type PEMDASIntLexer struct {
	reader        *bufio.Reader
	buf           []byte
	tokenStart    int
	tokenLocation *tokens.TokenLocation
	atEOF         bool
}

var _ liblexers.AbstractLexer = (*PEMDASIntLexer)(nil)

func NewPEMDASIntLexer(r io.Reader) liblexers.AbstractLexer {
	reader, ok := r.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(r)
	}
	return &PEMDASIntLexer{
		reader:        reader,
		buf:           make([]byte, 0, PEMDASIntLexerBufSize),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

// NewPEMDASIntLexerFromString returns a lexer over s (convenience for tests and -e mode).
func NewPEMDASIntLexerFromString(s string) liblexers.AbstractLexer {
	return NewPEMDASIntLexer(strings.NewReader(s))
}

func (lexer *PEMDASIntLexer) ensureFill(needBytes int) {
	for needBytes > len(lexer.buf) && !lexer.atEOF {
		chunk := make([]byte, PEMDASIntLexerBufSize)
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

func (lexer *PEMDASIntLexer) peekRuneAt(byteOffset int) (rune, int) {
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

func (lexer *PEMDASIntLexer) Scan() *tokens.Token {
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
		state := PEMDASIntLexerStartState
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
			nextState, ok := PEMDASIntLexerLookupTransition(state, r)
			if !ok {
				break
			}
			scanOffset += width
			state = nextState
			if _, ok := PEMDASIntLexerActions[state]; ok {
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
		tokenType := PEMDASIntLexerActions[lastAcceptState]
		if PEMDASIntLexerIsIgnoredToken(tokenType) {
			continue
		}
		return tokens.NewToken(lexeme, tokenType, &startLocation)
	}
}

func PEMDASIntLexerLookupTransition(state int, r rune) (int, bool) {
	transitionsForState, ok := PEMDASIntLexerTransitions[state]
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
func PEMDASIntLexerIsIgnoredToken(tokenType tokens.TokenType) bool {
	return strings.HasPrefix(string(tokenType), "!")
}

const PEMDASIntLexerStartState = 0

type PEMDASIntLexerRangeTransition struct {
	from rune
	to   rune
	next int
}

var PEMDASIntLexerTransitions = map[int][]PEMDASIntLexerRangeTransition{
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

var PEMDASIntLexerActions = map[int]tokens.TokenType{
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
