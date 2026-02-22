package lexers

import (
	"fmt"
	"strings"

	"github.com/johnkerl/pgpg/lib/go/pkg/tokens"
	"unicode"
	"unicode/utf8"
)

const vbcLexerInitialCapacity = 1024

const (
	VBCLexerTypeIdentifier tokens.TokenType = "identifier"
	VBCLexerTypeAnd        tokens.TokenType = "AND"
	VBCLexerTypeOr         tokens.TokenType = "OR"
	VBCLexerTypeNot        tokens.TokenType = "NOT"
	VBCLexerTypeLParen     tokens.TokenType = "("
	VBCLexerTypeRParen     tokens.TokenType = ")"
)

// VBCLexer is for boolean expressions with identifiers, AND/OR/NOT, and parentheses.
type VBCLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
}

func NewVBCLexer(inputText string) AbstractLexer {
	return &VBCLexer{
		inputText:     inputText,
		inputLength:   len(inputText),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

func (lexer *VBCLexer) Scan() (token *tokens.Token) {
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	lexer.ignoreNextRunesIf(unicode.IsSpace)
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	startLocation := *lexer.tokenLocation

	r, runeWidth := lexer.peekRune()

	if r == '(' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, VBCLexerTypeLParen, &startLocation)

	} else if r == ')' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, VBCLexerTypeRParen, &startLocation)

	} else if isVICIdentifierStart(r) {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		runes := make([]rune, 0, vbcLexerInitialCapacity)
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
		lexeme := string(runes)
		if strings.EqualFold(lexeme, "AND") {
			return tokens.NewToken(runes, VBCLexerTypeAnd, &startLocation)
		}
		if strings.EqualFold(lexeme, "OR") {
			return tokens.NewToken(runes, VBCLexerTypeOr, &startLocation)
		}
		if strings.EqualFold(lexeme, "NOT") {
			return tokens.NewToken(runes, VBCLexerTypeNot, &startLocation)
		}
		return tokens.NewToken(runes, VBCLexerTypeIdentifier, &startLocation)
	}
	return tokens.NewErrorToken(
		fmt.Sprintf("VBC lexer: unrecognized token %q (%U)", r, r),
		lexer.tokenLocation,
	)
}

func (lexer *VBCLexer) ignoreNextRuneIf(predicate RunePredicateFunc) bool {
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
	}
	return false
}

func (lexer *VBCLexer) ignoreNextRunesIf(predicate RunePredicateFunc) {
	for lexer.ignoreNextRuneIf(predicate) {
	}
}

func (lexer *VBCLexer) peekRune() (rune, int) {
	r, runeWidth := utf8.DecodeRuneInString(lexer.inputText[lexer.tokenLocation.ByteOffset:])
	return r, runeWidth
}

func (lexer *VBCLexer) readRune() rune {
	r, runeWidth := lexer.peekRune()
	lexer.tokenLocation.LocateRune(r, runeWidth)
	return r
}
