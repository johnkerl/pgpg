package lexers

import (
	"github.com/johnkerl/pgpg/pkg/tokens"
)

type AbstractLexer interface {
	// On EOF, the token-type will be EOF.
	// On error, the token-type will be Error, with Lexeme slot having the errortext.
	Scan() (token *tokens.Token)

	// For efficiency we carry around integers in bulk. But for human-readable displays
	// we want human-friendly strings.
	DecodeType(tokenType tokens.TokenType) (string, error)
}
