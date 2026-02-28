package lexers

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/johnkerl/pgpg/go/lib/pkg/tokens"
)

const ebnfLexerInitialCapacity = 1024

const (
	EBNFLexerTypeIdentifier tokens.TokenType = "identifier"
	EBNFLexerTypeString     tokens.TokenType = "string"
	EBNFLexerTypeAssign      tokens.TokenType = "::="
	EBNFLexerTypeOr          tokens.TokenType = "|"
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
// It reads from an io.Reader (streaming); use NewEBNFLexerFromString for string input.
type EBNFLexer struct {
	reader        *bufio.Reader
	tokenLocation *tokens.TokenLocation
	sourceName    string
	// One-rune peek for lookahead without consuming.
	hasPeek bool
	peekR   rune
	peekW   int
	atEOF   bool
}

// NewEBNFLexer returns a lexer that reads from r (streaming). For string input use NewEBNFLexerFromString.
func NewEBNFLexer(r io.Reader) AbstractLexer {
	return NewEBNFLexerWithSourceName(r, "")
}

// NewEBNFLexerWithSourceName returns a lexer that reads from r with a source name for error messages.
func NewEBNFLexerWithSourceName(r io.Reader, sourceName string) AbstractLexer {
	reader, ok := r.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(r)
	}
	return &EBNFLexer{
		reader:        reader,
		tokenLocation: tokens.NewTokenLocation(),
		sourceName:    sourceName,
	}
}

// NewEBNFLexerFromString returns a lexer over s (convenience for tests and -e mode).
func NewEBNFLexerFromString(s string) AbstractLexer {
	return NewEBNFLexer(strings.NewReader(s))
}

// NewEBNFLexerFromStringWithSourceName is like NewEBNFLexerFromString with a source name.
func NewEBNFLexerFromStringWithSourceName(s string, sourceName string) AbstractLexer {
	return NewEBNFLexerWithSourceName(strings.NewReader(s), sourceName)
}

func (lexer *EBNFLexer) isAtEOF() bool {
	return lexer.atEOF && !lexer.hasPeek
}

// peekRune returns the next rune and its byte width without consuming. Returns (0, 0) at EOF.
func (lexer *EBNFLexer) peekRune() (rune, int) {
	if lexer.atEOF && !lexer.hasPeek {
		return 0, 0
	}
	if lexer.hasPeek {
		return lexer.peekR, lexer.peekW
	}
	r, size, err := lexer.reader.ReadRune()
	if err == io.EOF {
		lexer.atEOF = true
		return 0, 0
	}
	if err != nil {
		lexer.atEOF = true
		return 0, 0
	}
	if size == 0 {
		return 0, 0
	}
	lexer.hasPeek = true
	lexer.peekR = r
	lexer.peekW = size
	return r, size
}

// consumePeek clears the one-rune peek after the rune has been consumed (LocateRune called).
func (lexer *EBNFLexer) consumePeek() {
	lexer.hasPeek = false
}

func (lexer *EBNFLexer) Scan() (token *tokens.Token) {
	if lexer.isAtEOF() {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	for {
		lexer.ignoreNextRunesIf(unicode.IsSpace)
		if lexer.isAtEOF() {
			return tokens.NewEOFToken(lexer.tokenLocation)
		}
		r, runeWidth := lexer.peekRune()
		if r != '#' {
			break
		}
		lexer.tokenLocation.LocateRune(r, runeWidth)
		lexer.consumePeek()
		for {
			if lexer.isAtEOF() {
				return tokens.NewEOFToken(lexer.tokenLocation)
			}
			r, runeWidth = lexer.peekRune()
			lexer.tokenLocation.LocateRune(r, runeWidth)
			lexer.consumePeek()
			if r == '\n' {
				break
			}
		}
	}
	if lexer.isAtEOF() {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	startLocation := *lexer.tokenLocation

	r, runeWidth := lexer.peekRune()

	if r == ':' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		lexer.consumePeek()
		if lexer.isAtEOF() {
			return tokens.NewToken([]rune{':'}, EBNFLexerTypeColon, &startLocation)
		}
		nextRune, nextWidth := lexer.peekRune()
		if nextRune != ':' {
			return tokens.NewToken([]rune{':'}, EBNFLexerTypeColon, &startLocation)
		}
		lexer.tokenLocation.LocateRune(nextRune, nextWidth)
		lexer.consumePeek()
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
		lexer.consumePeek()
		return tokens.NewToken([]rune{':', ':', '='}, EBNFLexerTypeAssign, &startLocation)

	} else if r == '=' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		lexer.consumePeek()
		return tokens.NewToken([]rune{r}, EBNFLexerTypeAssign, &startLocation)

	} else if r == '|' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		lexer.consumePeek()
		return tokens.NewToken([]rune{r}, EBNFLexerTypeOr, &startLocation)

	} else if r == '(' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		lexer.consumePeek()
		return tokens.NewToken([]rune{r}, EBNFLexerTypeLParen, &startLocation)

	} else if r == ')' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		lexer.consumePeek()
		return tokens.NewToken([]rune{r}, EBNFLexerTypeRParen, &startLocation)

	} else if r == '[' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		lexer.consumePeek()
		return tokens.NewToken([]rune{r}, EBNFLexerTypeLBracket, &startLocation)

	} else if r == ']' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		lexer.consumePeek()
		return tokens.NewToken([]rune{r}, EBNFLexerTypeRBracket, &startLocation)

	} else if r == '{' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		lexer.consumePeek()
		return tokens.NewToken([]rune{r}, EBNFLexerTypeLBrace, &startLocation)

	} else if r == '}' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		lexer.consumePeek()
		return tokens.NewToken([]rune{r}, EBNFLexerTypeRBrace, &startLocation)

	} else if r == ';' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		lexer.consumePeek()
		return tokens.NewToken([]rune{r}, EBNFLexerTypeSemicolon, &startLocation)

	} else if r == '-' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		lexer.consumePeek()
		if !lexer.isAtEOF() {
			nextR, nextW := lexer.peekRune()
			if nextR == '>' {
				lexer.tokenLocation.LocateRune(nextR, nextW)
				lexer.consumePeek()
				return tokens.NewToken([]rune{'-', '>'}, EBNFLexerTypeArrow, &startLocation)
			}
		}
		return tokens.NewToken([]rune{r}, EBNFLexerTypeDash, &startLocation)

	} else if r == ',' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		lexer.consumePeek()
		return tokens.NewToken([]rune{r}, EBNFLexerTypeComma, &startLocation)

	} else if r == '.' {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		lexer.consumePeek()
		return tokens.NewToken([]rune{r}, EBNFLexerTypeDot, &startLocation)

	} else if unicode.IsDigit(r) {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		lexer.consumePeek()
		runes := []rune{r}
		for !lexer.isAtEOF() {
			nextR, nextW := lexer.peekRune()
			if unicode.IsDigit(nextR) {
				lexer.tokenLocation.LocateRune(nextR, nextW)
				lexer.consumePeek()
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
		lexer.consumePeek()
		runes := make([]rune, 0, ebnfLexerInitialCapacity)
		runes = append(runes, r)

		for {
			r, runeWidth := lexer.peekRune()
			if isEBNFIdentifierContinue(r) {
				lexer.tokenLocation.LocateRune(r, runeWidth)
				lexer.consumePeek()
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
	lexer.consumePeek()
	runes := make([]rune, 0, ebnfLexerInitialCapacity)
	runes = append(runes, quote)

	for {
		if lexer.isAtEOF() {
			return tokens.NewErrorToken(
				fmt.Sprintf("EBNF lexer: unterminated string literal at %s", lexer.formatLocation(startLocation)),
				lexer.tokenLocation,
			)
		}
		r, runeWidth := lexer.peekRune()
		lexer.tokenLocation.LocateRune(r, runeWidth)
		lexer.consumePeek()
		runes = append(runes, r)

		if r == '\\' {
			if lexer.isAtEOF() {
				return tokens.NewErrorToken(
					fmt.Sprintf("EBNF lexer: unterminated escape in string literal at %s", lexer.formatLocation(startLocation)),
					lexer.tokenLocation,
				)
			}
			r, runeWidth = lexer.peekRune()
			lexer.tokenLocation.LocateRune(r, runeWidth)
			lexer.consumePeek()
			runes = append(runes, r)
			continue
		}
		if r == quote {
			break
		}
	}

	return tokens.NewToken(runes, EBNFLexerTypeString, startLocation)
}

func (lexer *EBNFLexer) ignoreNextRuneIf(predicate RunePredicateFunc) bool {
	if lexer.isAtEOF() {
		return false
	}
	r, runeWidth := lexer.peekRune()
	if runeWidth == 0 {
		return false
	}

	if predicate(r) {
		lexer.tokenLocation.LocateRune(r, runeWidth)
		lexer.consumePeek()
		return true
	}
	return false
}

func (lexer *EBNFLexer) ignoreNextRunesIf(predicate RunePredicateFunc) {
	for lexer.ignoreNextRuneIf(predicate) {
	}
}

// readRune gets the next rune from the input and updates location information.
func (lexer *EBNFLexer) readRune() rune {
	r, runeWidth := lexer.peekRune()
	lexer.tokenLocation.LocateRune(r, runeWidth)
	lexer.consumePeek()
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
	return r == '!' || r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func isEBNFIdentifierContinue(r rune) bool {
	return isEBNFIdentifierStart(r) || (r >= '0' && r <= '9')
}
