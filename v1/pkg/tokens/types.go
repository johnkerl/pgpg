package tokens

type TokenType int

// These token-types are common to all lexers. Then, any given lexer impl will have its own
// (nominally non-negative) types in addition.
const TokenTypeEOF = -1
const TokenTypeError = -2

// Token tracks a single lexeme and its type (as determined by the lexer) as well as where it was
// found within the source text.
type Token struct {
	Lexeme   []rune
	Type     TokenType
	Location TokenLocation
}
