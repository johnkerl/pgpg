package lexers

import (
	"fmt"

	"github.com/johnkerl/pgpg/manual/pkg/tokens"
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
	EBNFLexerTypeDash       tokens.TokenType = "-"
	EBNFLexerTypeDot        tokens.TokenType = "."
	EBNFLexerTypeArrow      tokens.TokenType = "->"
	EBNFLexerTypeColon      tokens.TokenType = ":"
	EBNFLexerTypeComma      tokens.TokenType = ","
	EBNFLexerTypeInteger    tokens.TokenType = "integer"
)

// EBNFLexer tokenizes a common EBNF dialect with identifiers, string literals,
// ::= assignments (or =), alternation, and grouping operators.
type EBNFLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
	sourceName    string
}

func NewEBNFLexer(inputText string) AbstractLexer {
	return NewEBNFLexerWithSourceName(inputText, "")
}

func NewEBNFLexerWithSourceName(inputText string, sourceName string) AbstractLexer {
	return &EBNFLexer{
		inputText:     inputText,
		inputLength:   len(inputText),
		tokenLocation: tokens.NewTokenLocation(),
		sourceName:    sourceName,
	}
}

func (lexer *EBNFLexer) Scan() (token *tokens.Token) {
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	for {
		lexer.ignoreNextRunesIf(unicode.IsSpace)
		if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
			return tokens.NewEOFToken(lexer.tokenLocation)
		}
		r, runeWidth := lexer.peekRune()
		if r != '#' {
			break
		}
		lexer.tokenLocation.LocateRune(r, runeWidth)
		for {
			if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
				return tokens.NewEOFToken(lexer.tokenLocation)
			}
			r, runeWidth = lexer.peekRune()
			lexer.tokenLocation.LocateRune(r, runeWidth)
			if r == '\n' {
				break
			}
		}
	}
	if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	startLocation := *lexer.tokenLocation

	r, runeWidth := lexer.peekRune()

	if r == ':' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
			return tokens.NewToken([]rune{':'}, EBNFLexerTypeColon, &startLocation)
		}
		nextRune, nextWidth := lexer.peekRune()
		if nextRune != ':' {
			return tokens.NewToken([]rune{':'}, EBNFLexerTypeColon, &startLocation)
		}
		lexer.tokenLocation.LocateRune(nextRune, nextWidth)
		nextRune, nextWidth = lexer.peekRune()
		if nextRune != '=' {
			return tokens.NewErrorToken(
				fmt.Sprintf(
					"EBNF lexer: expected '::=' but found '::%c' at %s",
					nextRune,
					lexer.formatLocation(&startLocation),
				),
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

	} else if r == '-' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		if lexer.tokenLocation.ByteOffset < lexer.inputLength {
			nextR, nextW := lexer.peekRune()
			if nextR == '>' {
				lexer.tokenLocation.LocateRune(nextR, nextW)
				return tokens.NewToken([]rune{'-', '>'}, EBNFLexerTypeArrow, &startLocation)
			}
		}
		return tokens.NewToken([]rune{r}, EBNFLexerTypeDash, &startLocation)

	} else if r == ',' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, EBNFLexerTypeComma, &startLocation)

	} else if r == '.' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		return tokens.NewToken([]rune{r}, EBNFLexerTypeDot, &startLocation)

	} else if unicode.IsDigit(r) {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		runes := []rune{r}
		for lexer.tokenLocation.ByteOffset < lexer.inputLength {
			nextR, nextW := lexer.peekRune()
			if unicode.IsDigit(nextR) {
				lexer.tokenLocation.LocateRune(nextR, nextW)
				runes = append(runes, nextR)
			} else {
				break
			}
		}
		return tokens.NewToken(runes, EBNFLexerTypeInteger, &startLocation)

	} else if r == '"' || r == '\'' {
		return lexer.scanStringLiteral(r, runeWidth, &startLocation)

	} else if isEBNFIdentifierStart(r) {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		runes := make([]rune, 0, ebnfLexerInitialCapacity)
		runes = append(runes, r)

		for {
			r, runeWidth := lexer.peekRune()
			if isEBNFIdentifierContinue(r) {
				lexer.tokenLocation.LocateRune(r, runeWidth)
				runes = append(runes, r)
			} else {
				break
			}
		}
		return tokens.NewToken(runes, EBNFLexerTypeIdentifier, &startLocation)

	} else {
		return tokens.NewErrorToken(
			fmt.Sprintf("EBNF lexer: unrecognized token %q (%U) at %s", r, r, lexer.formatLocation(&startLocation)),
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
			return tokens.NewErrorToken(
				fmt.Sprintf("EBNF lexer: unterminated string literal at %s", lexer.formatLocation(startLocation)),
				lexer.tokenLocation,
			)
		}
		r, runeWidth := lexer.peekRune()
		lexer.tokenLocation.LocateRune(r, runeWidth)
		runes = append(runes, r)

		if r == '\\' {
			if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
				return tokens.NewErrorToken(
					fmt.Sprintf("EBNF lexer: unterminated escape in string literal at %s", lexer.formatLocation(startLocation)),
					lexer.tokenLocation,
				)
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

func (lexer *EBNFLexer) formatLocation(location *tokens.TokenLocation) string {
	if location == nil {
		return "unknown location"
	}
	if lexer.sourceName != "" {
		return fmt.Sprintf(
			"%s, line %d, column %d",
			lexer.sourceName,
			location.LineNumber,
			location.ColumnNumber,
		)
	}
	return fmt.Sprintf("line %d, column %d", location.LineNumber, location.ColumnNumber)
}

func isEBNFIdentifierStart(r rune) bool {
	return r == '!' || isVICIdentifierStart(r)
}

func isEBNFIdentifierContinue(r rune) bool {
	return isVICIdentifierContinue(r)
}
