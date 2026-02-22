package lexers

import (
	"fmt"

	"github.com/johnkerl/pgpg/lib/go/pkg/tokens"
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
}

func (lexer *AMLexer) ignoreNextRuneIf(predicate RunePredicateFunc) bool {
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

func (lexer *AMLexer) ignoreNextRunesIf(predicate RunePredicateFunc) {
	for lexer.ignoreNextRuneIf(predicate) {
	}
}

func (lexer *AMLexer) peekRune() (rune, int) {
	r, runeWidth := utf8.DecodeRuneInString(lexer.inputText[lexer.tokenLocation.ByteOffset:])
	return r, runeWidth
}

func (lexer *AMLexer) readRune() rune {
	r, runeWidth := lexer.peekRune()
	lexer.tokenLocation.LocateRune(r, runeWidth)
	return r
}
