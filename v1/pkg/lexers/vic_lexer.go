package lexers

import (
	"fmt"

	"github.com/johnkerl/pgpg/pkg/tokens"
	"unicode"
	"unicode/utf8"
)

const vicLexerInitialCapacity = 1024

const (
	VICLexerTypeNumber     tokens.TokenType = "number"
	VICLexerTypeIdentifier tokens.TokenType = "identifier"
	VICLexerTypePlus       tokens.TokenType = "+"
	VICLexerTypeMinus      tokens.TokenType = "-"
	VICLexerTypeTimes      tokens.TokenType = "*"
	VICLexerTypeDivide     tokens.TokenType = "/"
	VICLexerTypePower      tokens.TokenType = "**"
	VICLexerTypeAssign     tokens.TokenType = "="
	VICLexerTypeLParen     tokens.TokenType = "("
	VICLexerTypeRParen     tokens.TokenType = ")"
)

// VICLexer is for arithmetic with parentheses, exponentiation, multiplication/division, and
// addition/subtraction, plus identifiers and assignments. It only tokenizes; the parser enforces
// precedence and associativity.
type VICLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
}

func NewVICLexer(inputText string) AbstractLexer {
	return &VICLexer{
		inputText:     inputText,
		inputLength:   len(inputText),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

func (lexer *VICLexer) Scan() (token *tokens.Token) {
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	lexer.ignoreNextRunesIf(unicode.IsSpace)
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	startLocation := *lexer.tokenLocation

	// Look for: single operators/parens, or one or more consecutive digits/identifiers.
	r, runeWidth := lexer.peekRune()

	if r == '+' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, VICLexerTypePlus, &startLocation)

	} else if r == '-' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, VICLexerTypeMinus, &startLocation)

	} else if r == '*' {
		nextRune, nextWidth := utf8.DecodeRuneInString(lexer.inputText[lexer.tokenLocation.ByteOffset+runeWidth:])
		if nextRune == '*' {
			lexer.tokenLocation.LocateRune(r, runeWidth)
			lexer.tokenLocation.LocateRune(nextRune, nextWidth)
			return tokens.NewToken([]rune{r, nextRune}, VICLexerTypePower, &startLocation)
		}
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, VICLexerTypeTimes, &startLocation)

	} else if r == '/' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, VICLexerTypeDivide, &startLocation)

	} else if r == '=' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, VICLexerTypeAssign, &startLocation)

	} else if r == '(' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, VICLexerTypeLParen, &startLocation)

	} else if r == ')' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, VICLexerTypeRParen, &startLocation)

	} else if isVICIdentifierStart(r) {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		runes := make([]rune, 0, vicLexerInitialCapacity)
		runes = append(runes, r)

		for {
			r, runeWidth := lexer.peekRune()
			if isVICIdentifierContinue(r) {
				lexer.tokenLocation.LocateRune(r, runeWidth)
				runes = append(runes, r)
			} else {
				break
			}
		}
		return tokens.NewToken(runes, VICLexerTypeIdentifier, &startLocation)

	} else if r >= '0' && r <= '9' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		runes := make([]rune, 0, vicLexerInitialCapacity)
		runes = append(runes, r)

		for {
			r, runeWidth := lexer.peekRune()
			if r >= '0' && r <= '9' {
				lexer.tokenLocation.LocateRune(r, runeWidth)
				runes = append(runes, r)
			} else {
				break
			}
		}
		return tokens.NewToken(runes, VICLexerTypeNumber, &startLocation)

	} else {
		return tokens.NewErrorToken(
			fmt.Sprintf("VIC lexer: unrecognized token %q (%U)", r, r),
			lexer.tokenLocation,
		)
	}
}

func isVICIdentifierStart(r rune) bool {
	return r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func isVICIdentifierContinue(r rune) bool {
	return isVICIdentifierStart(r) || (r >= '0' && r <= '9')
}

func (lexer *VICLexer) ignoreNextRuneIf(predicate runePredicateFunc) bool {
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

func (lexer *VICLexer) ignoreNextRunesIf(predicate runePredicateFunc) {
	for lexer.ignoreNextRuneIf(predicate) {
	}
}

// peekRune gets the next rune from the input without updating location information.
func (lexer *VICLexer) peekRune() (rune, int) {
	r, runeWidth := utf8.DecodeRuneInString(lexer.inputText[lexer.tokenLocation.ByteOffset:])
	return r, runeWidth
}

// readRune gets the next rune from the input and updates location information.
func (lexer *VICLexer) readRune() rune {
	r, runeWidth := lexer.peekRune()
	lexer.tokenLocation.LocateRune(r, runeWidth)
	return r
}
