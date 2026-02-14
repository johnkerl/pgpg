package lexers

import (
	"fmt"

	"github.com/johnkerl/pgpg/manual/go/pkg/tokens"
	"unicode"
	"unicode/utf8"
)

const amLexerInitialCapacity = 1024

const (
	AMLexerTypeNumber tokens.TokenType = "number"
	AMLexerTypePlus   tokens.TokenType = "+"
	AMLexerTypeTimes  tokens.TokenType = "*"
)

// AMLexer is for the AME and AMNE grammars: addition and multiplication of integers.  At the syntax
// level, AME has equal operator precedence, while AMNE binds multiplication more tightly than
// addition. But here at the lex level, they're the same.
type AMLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
}

func NewAMLexer(inputText string) AbstractLexer {
	return &AMLexer{
		inputText:     inputText,
		inputLength:   len(inputText),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

func (lexer *AMLexer) Scan() (token *tokens.Token) {
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	lexer.ignoreNextRunesIf(unicode.IsSpace)
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	startLocation := *lexer.tokenLocation

	// Look for: single '+', single '*', or one or more consecutive digits.
	//
	// It's syntactically wrong if the expression starts with a "*", or of there are two "+" in a
	// row, etc etc etc -- we absolutely have the power to check for those here -- but strictly
	// speaking our job here is only to split the input text into tokens and let the parser
	// determine whether the syntax is acceptable or not.
	//
	// That means we don't need to track whether we're in a state of "just saw a number, next must
	// be plus, times, or EOF".

	r, runeWidth := lexer.peekRune()

	if r == '+' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, AMLexerTypePlus, &startLocation)

	} else if r == '*' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, AMLexerTypeTimes, &startLocation)

	} else if unicode.IsDigit(r) {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		runes := make([]rune, 0, amLexerInitialCapacity)
		runes = append(runes, r)

		// TODO: make a method to detect runs of digits and only do a LocateRunes once.
		for {
			r, runeWidth := lexer.peekRune()
			if unicode.IsDigit(r) {
				lexer.tokenLocation.LocateRune(r, runeWidth)
				runes = append(runes, r)
			} else {
				break
			}
		}
		return tokens.NewToken(runes, AMLexerTypeNumber, &startLocation)

	} else {
		return tokens.NewErrorToken(
			fmt.Sprintf("AM lexer: unrecognized token %q (%U)", r, r),
			lexer.tokenLocation,
		)
	}
	panic("not reached")
}

func (lexer *AMLexer) ignoreNextRuneIf(predicate runePredicateFunc) bool {
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return false
	}
	r, runeWidth := lexer.peekRune()
	if runeWidth == 0 {
		return false
	}

	if predicate(r) {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return true
	} else {
		return false
	}
}

func (lexer *AMLexer) ignoreNextRunesIf(predicate runePredicateFunc) {
	for lexer.ignoreNextRuneIf(predicate) {
	}
}

// peekRune gets the next rune from the input without updating location information.
func (lexer *AMLexer) peekRune() (rune, int) {
	r, runeWidth := utf8.DecodeRuneInString(lexer.inputText[lexer.tokenLocation.ByteOffset:])
	return r, runeWidth
}

// readRune gets the next rune from the input and updates location information.
func (lexer *AMLexer) readRune() rune {
	r, runeWidth := lexer.peekRune()
	lexer.tokenLocation.LocateRune(r, runeWidth)
	return r
}
