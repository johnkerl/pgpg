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

const StatementsLexerBufSize = 4096

type StatementsLexer struct {
	reader        *bufio.Reader
	buf           []byte
	tokenStart    int
	tokenLocation *tokens.TokenLocation
	atEOF         bool
}

var _ liblexers.AbstractLexer = (*StatementsLexer)(nil)

func NewStatementsLexer(r io.Reader) liblexers.AbstractLexer {
	reader, ok := r.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(r)
	}
	return &StatementsLexer{
		reader:        reader,
		buf:           make([]byte, 0, StatementsLexerBufSize),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

// NewStatementsLexerFromString returns a lexer over s (convenience for tests and -e mode).
func NewStatementsLexerFromString(s string) liblexers.AbstractLexer {
	return NewStatementsLexer(strings.NewReader(s))
}

func (lexer *StatementsLexer) ensureFill(needBytes int) {
	for needBytes > len(lexer.buf) && !lexer.atEOF {
		chunk := make([]byte, StatementsLexerBufSize)
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

func (lexer *StatementsLexer) peekRuneAt(byteOffset int) (rune, int) {
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

func (lexer *StatementsLexer) Scan() *tokens.Token {
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
		state := StatementsLexerStartState
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
			nextState, ok := StatementsLexerLookupTransition(state, r)
			if !ok {
				break
			}
			scanOffset += width
			state = nextState
			if _, ok := StatementsLexerActions[state]; ok {
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
		tokenType := StatementsLexerActions[lastAcceptState]
		if StatementsLexerIsIgnoredToken(tokenType) {
			continue
		}
		return tokens.NewToken(lexeme, tokenType, &startLocation)
	}
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
