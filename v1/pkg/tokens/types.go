package tokens

type TokenType string

// These token-types are common to all lexers. Then, any given lexer impl will have its own
// (nominally non-negative) types in addition.
const TokenTypeEOF TokenType = "EOF"
const TokenTypeError TokenType = "ERROR"

// TokenLocation contains information for lexer state, namely the ByteOffset, as well
// as user-facing information in the form of LineNumber and ColumnNumber.
// A string FileName is not included -- I feel like this is too bulky to keep this for every single
// token from a given input file. The filename information should be tracked up a level.
type TokenLocation struct {
	LineNumber   int
	ColumnNumber int
	ByteOffset   int
}

// Token tracks a single lexeme and its type (as determined by the lexer) as well as where it was
// found within the source text.
type Token struct {
	Lexeme   []rune
	Type     TokenType
	Location TokenLocation
}
