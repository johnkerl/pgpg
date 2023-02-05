package lexers

import (
	"github.com/johnkerl/pgpg/pkg/tokens"
)

type AbstractLexer interface {
	// On EOF, this should return (nil, nil)
	Scan() (token *tokens.Token, err error)
}
