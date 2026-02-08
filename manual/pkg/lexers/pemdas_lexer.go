package lexers

import (
	"fmt"

	"github.com/johnkerl/pgpg/pkg/tokens"
	"unicode"
	"unicode/utf8"
)

const pemdasLexerInitialCapacity = 1024

const (
	PEMDASLexerTypeNumber tokens.TokenType = "number"
	PEMDASLexerTypePlus   tokens.TokenType = "+"
	PEMDASLexerTypeMinus  tokens.TokenType = "-"
	PEMDASLexerTypeTimes  tokens.TokenType = "*"
	PEMDASLexerTypeDivide tokens.TokenType = "/"
	PEMDASLexerTypePower  tokens.TokenType = "**"
	PEMDASLexerTypeLParen tokens.TokenType = "("
	PEMDASLexerTypeRParen tokens.TokenType = ")"
)

// PEMDASLexer is for arithmetic with parentheses, exponentiation, multiplication/division, and
// addition/subtraction. It only tokenizes; the parser enforces precedence and associativity.
type PEMDASLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
}

func NewPEMDASLexer(inputText string) AbstractLexer {
	return &PEMDASLexer{
		inputText:     inputText,
		inputLength:   len(inputText),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

func (lexer *PEMDASLexer) Scan() (token *tokens.Token) {
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	lexer.ignoreNextRunesIf(unicode.IsSpace)
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	startLocation := *lexer.tokenLocation

	// Look for: single operators/parens, or one or more consecutive digits.
	r, runeWidth := lexer.peekRune()

	if r == '+' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, PEMDASLexerTypePlus, &startLocation)

	} else if r == '-' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, PEMDASLexerTypeMinus, &startLocation)

	} else if r == '*' {
		nextRune, nextWidth := utf8.DecodeRuneInString(lexer.inputText[lexer.tokenLocation.ByteOffset+runeWidth:])
		if nextRune == '*' {
			lexer.tokenLocation.LocateRune(r, runeWidth)
			lexer.tokenLocation.LocateRune(nextRune, nextWidth)
			return tokens.NewToken([]rune{r, nextRune}, PEMDASLexerTypePower, &startLocation)
		}
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, PEMDASLexerTypeTimes, &startLocation)

	} else if r == '/' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, PEMDASLexerTypeDivide, &startLocation)

	} else if r == '(' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, PEMDASLexerTypeLParen, &startLocation)

	} else if r == ')' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, PEMDASLexerTypeRParen, &startLocation)

	} else if unicode.IsDigit(r) {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		runes := make([]rune, 0, pemdasLexerInitialCapacity)
		runes = append(runes, r)

		for {
			r, runeWidth := lexer.peekRune()
			if unicode.IsDigit(r) {
				lexer.tokenLocation.LocateRune(r, runeWidth)
				runes = append(runes, r)
			} else {
				break
			}
		}
		return tokens.NewToken(runes, PEMDASLexerTypeNumber, &startLocation)

	} else {
		return tokens.NewErrorToken(
			fmt.Sprintf("PEMDAS lexer: unrecognized token %q (%U)", r, r),
			lexer.tokenLocation,
		)
	}
}

func (lexer *PEMDASLexer) ignoreNextRuneIf(predicate runePredicateFunc) bool {
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

func (lexer *PEMDASLexer) ignoreNextRunesIf(predicate runePredicateFunc) {
	for lexer.ignoreNextRuneIf(predicate) {
	}
}

// peekRune gets the next rune from the input without updating location information.
func (lexer *PEMDASLexer) peekRune() (rune, int) {
	r, runeWidth := utf8.DecodeRuneInString(lexer.inputText[lexer.tokenLocation.ByteOffset:])
	return r, runeWidth
}

// readRune gets the next rune from the input and updates location information.
func (lexer *PEMDASLexer) readRune() rune {
	r, runeWidth := lexer.peekRune()
	lexer.tokenLocation.LocateRune(r, runeWidth)
	return r
}
