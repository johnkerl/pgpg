package lexer

import "github.com/johnkerl/pgpg/v2/internal/errutil"

// TokenType distinguishes lexeme categories.
type TokenType struct {
	Name string
}

// Token is a single lexeme with source position.
type Token struct {
	Type   TokenType
	Value  string
	Offset int
	Line   int
	Column int
}

// Lexer turns input bytes into tokens.
type Lexer interface {
	Next() (Token, error)
	Peek() (Token, error)
}

// New returns a new lexer for the given input.
// TODO: implement NFA->DFA with longest-match and rule priority.
func New(_ []byte) (Lexer, error) {
	return nil, errutil.NotImplemented("lexer.New")
}
