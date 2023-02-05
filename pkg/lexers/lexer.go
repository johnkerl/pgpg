package lexers

import (
	"github.com/johnkerl/pgpg/pkg/tokens"
)

type AbstractLexer interface {
	Scan() (token *tokens.Token, err error)
}
