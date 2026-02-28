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

const LISPLexerBufSize = 4096

type LISPLexer struct {
	reader        *bufio.Reader
	buf           []byte
	tokenStart    int
	tokenLocation *tokens.TokenLocation
	atEOF         bool
}

var _ liblexers.AbstractLexer = (*LISPLexer)(nil)

func NewLISPLexer(r io.Reader) liblexers.AbstractLexer {
	reader, ok := r.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(r)
	}
	return &LISPLexer{
		reader:        reader,
		buf:           make([]byte, 0, LISPLexerBufSize),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

// NewLISPLexerFromString returns a lexer over s (convenience for tests and -e mode).
func NewLISPLexerFromString(s string) liblexers.AbstractLexer {
	return NewLISPLexer(strings.NewReader(s))
}

func (lexer *LISPLexer) ensureFill(needBytes int) {
	for needBytes > len(lexer.buf) && !lexer.atEOF {
		chunk := make([]byte, LISPLexerBufSize)
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

func (lexer *LISPLexer) peekRuneAt(byteOffset int) (rune, int) {
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

func (lexer *LISPLexer) Scan() *tokens.Token {
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
		state := LISPLexerStartState
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
			nextState, ok := LISPLexerLookupTransition(state, r)
			if !ok {
				break
			}
			scanOffset += width
			state = nextState
			if _, ok := LISPLexerActions[state]; ok {
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
		tokenType := LISPLexerActions[lastAcceptState]
		if LISPLexerIsIgnoredToken(tokenType) {
			continue
		}
		return tokens.NewToken(lexeme, tokenType, &startLocation)
	}
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
