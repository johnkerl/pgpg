package lexers

import (
	"fmt"
	"unicode/utf8"

	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type GeneratedLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
}

func NewGeneratedLexer(inputText string) *GeneratedLexer {
	return &GeneratedLexer{
		inputText:     inputText,
		inputLength:   len(inputText),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

func (lexer *GeneratedLexer) Scan() *tokens.Token {
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	startLocation := *lexer.tokenLocation
	scanLocation := *lexer.tokenLocation
	state := startState
	lastAcceptState := -1
	lastAcceptLocation := scanLocation

	for {
		if scanLocation.ByteOffset >= lexer.inputLength {
			break
		}
		r, width := lexer.peekRuneAt(scanLocation.ByteOffset)
		nextState, ok := transitions[state][r]
		if !ok {
			break
		}
		scanLocation.LocateRune(r, width)
		state = nextState
		if _, ok := actions[state]; ok {
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
	return tokens.NewToken(lexeme, actions[lastAcceptState], &startLocation)
}

func (lexer *GeneratedLexer) peekRuneAt(byteOffset int) (rune, int) {
	r, width := utf8.DecodeRuneInString(lexer.inputText[byteOffset:])
	return r, width
}

const startState = 0

var transitions = map[int]map[rune]int{
	0: {
		'+': 1,
		'-': 2,
		'0': 3,
		'1': 4,
		'2': 5,
		'3': 6,
		'4': 7,
		'5': 8,
		'6': 9,
		'7': 10,
		'8': 11,
		'9': 12,
	},
}

var actions = map[int]tokens.TokenType{
	1: "sign",
	2: "sign",
	3: "digit",
	4: "digit",
	5: "digit",
	6: "digit",
	7: "digit",
	8: "digit",
	9: "digit",
	10: "digit",
	11: "digit",
	12: "digit",
}
