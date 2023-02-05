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

// NewToken constructs a new token, nominally for a lexer to use while scanning.
// The location is copied. The idea is that a lexer can keep a TokenLocation in its
// object state, updated with the LocateRune method, and then on producing a token
// we can copy that.
func NewToken(lexeme []rune, tokenType TokenType, location *TokenLocation) *Token {
	return &Token{
		Lexeme:   lexeme,
		Type:     tokenType,
		Location: *location, // does a copy
	}
}

// NewEOFToken is a keystroke-saver for constructing a token of type EOF.
func NewEOFToken(location *TokenLocation) *Token {
	return &Token{
		Lexeme:   nil,
		Type:     TokenTypeEOF,
		Location: *location, // does a copy
	}
}

// NewErrorToken is a keystroke-saver for constructing a token of type EOF.
func NewErrorToken(errorText string, location *TokenLocation) *Token {
	return &Token{
		Lexeme:   []rune(errorText),
		Type:     TokenTypeError,
		Location: *location, // does a copy
	}
}

// IsEOF is a keystroke-saver for determining if a token's type is EOF.
func (token *Token) IsEOF() bool {
	return token.Type == TokenTypeEOF
}

// IsEOF is a keystroke-saver for determining if a token's type is Error.
func (token *Token) IsError() bool {
	return token.Type == TokenTypeError
}
