package lexers

import (
	"fmt"

	"github.com/johnkerl/pgpg/pkg/tokens"
	"unicode"
	"unicode/utf8"
)

const ebnfLexerInitialCapacity = 1024

const (
	EBNFLexerTypeIdentifier tokens.TokenType = "identifier"
	EBNFLexerTypeString     tokens.TokenType = "string"
	EBNFLexerTypeAssign     tokens.TokenType = "::="
	EBNFLexerTypeOr         tokens.TokenType = "|"
	EBNFLexerTypeLParen     tokens.TokenType = "("
	EBNFLexerTypeRParen     tokens.TokenType = ")"
	EBNFLexerTypeLBracket   tokens.TokenType = "["
	EBNFLexerTypeRBracket   tokens.TokenType = "]"
	EBNFLexerTypeLBrace     tokens.TokenType = "{"
	EBNFLexerTypeRBrace     tokens.TokenType = "}"
	EBNFLexerTypeSemicolon  tokens.TokenType = ";"
)

// EBNFLexer tokenizes a common EBNF dialect with identifiers, string literals,
// ::= assignments (or =), alternation, and grouping operators.
type EBNFLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
}

func NewEBNFLexer(inputText string) AbstractLexer {
	return &EBNFLexer{
		inputText:     inputText,
		inputLength:   len(inputText),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

func (lexer *EBNFLexer) Scan() (token *tokens.Token) {
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	lexer.ignoreNextRunesIf(unicode.IsSpace)
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	startLocation := *lexer.tokenLocation

	r, runeWidth := lexer.peekRune()

	if r == ':' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		nextRune, nextWidth := lexer.peekRune()
		if nextRune != ':' {
			return tokens.NewErrorToken(
				fmt.Sprintf("EBNF lexer: expected '::=' but found ':%c'", nextRune),
				lexer.tokenLocation,
			)
		}
		lexer.tokenLocation.LocateRune(nextRune, nextWidth)
		nextRune, nextWidth = lexer.peekRune()
		if nextRune != '=' {
			return tokens.NewErrorToken(
				fmt.Sprintf("EBNF lexer: expected '::=' but found '::%c'", nextRune),
				lexer.tokenLocation,
			)
		}
		lexer.tokenLocation.LocateRune(nextRune, nextWidth)
		return tokens.NewToken([]rune{':', ':', '='}, EBNFLexerTypeAssign, &startLocation)

	} else if r == '=' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, EBNFLexerTypeAssign, &startLocation)

	} else if r == '|' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, EBNFLexerTypeOr, &startLocation)

	} else if r == '(' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, EBNFLexerTypeLParen, &startLocation)

	} else if r == ')' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, EBNFLexerTypeRParen, &startLocation)

	} else if r == '[' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, EBNFLexerTypeLBracket, &startLocation)

	} else if r == ']' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, EBNFLexerTypeRBracket, &startLocation)

	} else if r == '{' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, EBNFLexerTypeLBrace, &startLocation)

	} else if r == '}' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, EBNFLexerTypeRBrace, &startLocation)

	} else if r == ';' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, EBNFLexerTypeSemicolon, &startLocation)

	} else if r == '"' || r == '\'' {
		return lexer.scanStringLiteral(r, runeWidth, &startLocation)

	} else if isVICIdentifierStart(r) {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		runes := make([]rune, 0, ebnfLexerInitialCapacity)
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
		return tokens.NewToken(runes, EBNFLexerTypeIdentifier, &startLocation)

	} else {
		return tokens.NewErrorToken(
			fmt.Sprintf("EBNF lexer: unrecognized token %q (%U)", r, r),
			lexer.tokenLocation,
		)
	}
}

func (lexer *EBNFLexer) scanStringLiteral(
	quote rune,
	quoteWidth int,
	startLocation *tokens.TokenLocation,
) *tokens.Token {
	lexer.tokenLocation.LocateRune(quote, quoteWidth)
	runes := make([]rune, 0, ebnfLexerInitialCapacity)
	runes = append(runes, quote)

	for {
		if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
			return tokens.NewErrorToken("EBNF lexer: unterminated string literal", lexer.tokenLocation)
		}
		r, runeWidth := lexer.peekRune()
		lexer.tokenLocation.LocateRune(r, runeWidth)
		runes = append(runes, r)

		if r == '\\' {
			if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
				return tokens.NewErrorToken("EBNF lexer: unterminated escape in string literal", lexer.tokenLocation)
			}
			r, runeWidth = lexer.peekRune()
			lexer.tokenLocation.LocateRune(r, runeWidth)
			runes = append(runes, r)
			continue
		}
		if r == quote {
			break
		}
	}

	return tokens.NewToken(runes, EBNFLexerTypeString, startLocation)
}

func (lexer *EBNFLexer) ignoreNextRuneIf(predicate runePredicateFunc) bool {
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

func (lexer *EBNFLexer) ignoreNextRunesIf(predicate runePredicateFunc) {
	for lexer.ignoreNextRuneIf(predicate) {
	}
}

// peekRune gets the next rune from the input without updating location information.
func (lexer *EBNFLexer) peekRune() (rune, int) {
	r, runeWidth := utf8.DecodeRuneInString(lexer.inputText[lexer.tokenLocation.ByteOffset:])
	return r, runeWidth
}

// readRune gets the next rune from the input and updates location information.
func (lexer *EBNFLexer) readRune() rune {
	r, runeWidth := lexer.peekRune()
	lexer.tokenLocation.LocateRune(r, runeWidth)
	return r
}
