package tokens

import (
	"fmt"
)

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

func (token Token) String() string {
	return fmt.Sprintf(
		"token=<<%s>> type=%s line=%d column=%d",
		string(token.Lexeme),
		token.Type,
		token.Location.LineNumber,
		token.Location.ColumnNumber,
	)
}

func (token Token) LexemeText() string {
	return string(token.Lexeme)
}
