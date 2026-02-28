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

const SignDigitLexerBufSize = 4096

type SignDigitLexer struct {
	reader        *bufio.Reader
	buf           []byte
	tokenStart    int
	tokenLocation *tokens.TokenLocation
	atEOF         bool
}

var _ liblexers.AbstractLexer = (*SignDigitLexer)(nil)

func NewSignDigitLexer(r io.Reader) liblexers.AbstractLexer {
	reader, ok := r.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(r)
	}
	return &SignDigitLexer{
		reader:        reader,
		buf:           make([]byte, 0, SignDigitLexerBufSize),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

// NewSignDigitLexerFromString returns a lexer over s (convenience for tests and -e mode).
func NewSignDigitLexerFromString(s string) liblexers.AbstractLexer {
	return NewSignDigitLexer(strings.NewReader(s))
}

func (lexer *SignDigitLexer) ensureFill(needBytes int) {
	for needBytes > len(lexer.buf) && !lexer.atEOF {
		chunk := make([]byte, SignDigitLexerBufSize)
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

func (lexer *SignDigitLexer) peekRuneAt(byteOffset int) (rune, int) {
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

func (lexer *SignDigitLexer) Scan() *tokens.Token {
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
		state := SignDigitLexerStartState
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
			nextState, ok := SignDigitLexerLookupTransition(state, r)
			if !ok {
				break
			}
			scanOffset += width
			state = nextState
			if _, ok := SignDigitLexerActions[state]; ok {
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
		tokenType := SignDigitLexerActions[lastAcceptState]
		return tokens.NewToken(lexeme, tokenType, &startLocation)
	}
}

func SignDigitLexerLookupTransition(state int, r rune) (int, bool) {
	transitionsForState, ok := SignDigitLexerTransitions[state]
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

const SignDigitLexerStartState = 0

type SignDigitLexerRangeTransition struct {
	from rune
	to   rune
	next int
}

var SignDigitLexerTransitions = map[int][]SignDigitLexerRangeTransition{
	0: {
		{from: '+', to: '+', next: 1},
		{from: '-', to: '-', next: 2},
		{from: '0', to: '0', next: 3},
		{from: '1', to: '1', next: 4},
		{from: '2', to: '2', next: 5},
		{from: '3', to: '3', next: 6},
		{from: '4', to: '4', next: 7},
		{from: '5', to: '5', next: 8},
		{from: '6', to: '6', next: 9},
		{from: '7', to: '7', next: 10},
		{from: '8', to: '8', next: 11},
		{from: '9', to: '9', next: 12},
	},
}

var SignDigitLexerActions = map[int]tokens.TokenType{
	1:  "sign",
	2:  "sign",
	3:  "digit",
	4:  "digit",
	5:  "digit",
	6:  "digit",
	7:  "digit",
	8:  "digit",
	9:  "digit",
	10: "digit",
	11: "digit",
	12: "digit",
}
