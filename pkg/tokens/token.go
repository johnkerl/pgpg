package tokens

type TokenType int

// These token-types are common to all lexers. Then, any given lexer impl will have its own
// (nominally non-negative) types in addition.
const TokenTypeEOF = -1
const TokenTypeError = -2

// Token tracks a single lexeme as well as where it was found within the source text.
type Token struct {
	Lexeme   []rune
	Type     TokenType
	Location TokenLocation
}

func NewToken(lexeme []rune, tokenType TokenType, location *TokenLocation) *Token {
	return &Token{
		Lexeme:   lexeme,
		Type:     tokenType,
		Location: *location, // does a copy
	}
}
