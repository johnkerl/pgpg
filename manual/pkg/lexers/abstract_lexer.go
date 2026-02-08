package lexers

import (
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type AbstractLexer interface {
	// On EOF, the token-type will be EOF.
	// On error, the token-type will be Error, with Lexeme slot having the errortext.
	Scan() (token *tokens.Token)
}
