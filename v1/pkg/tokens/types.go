package tokens

type TokenType string

// These token-types are common to all lexers. Then, any given lexer impl will have its own
// (nominally non-negative) types in addition.
const TokenTypeEOF TokenType = "EOF"
const TokenTypeError TokenType = "ERROR"

// Token tracks a single lexeme and its type (as determined by the lexer) as well as where it was
// found within the source text.
type Token struct {
	Lexeme   []rune
	Type     TokenType
	Location TokenLocation
}
