package parsers

import (
	"github.com/johnkerl/pgpg/pkg/asts"
)

type AbstractParser[T asts.TokenLike] interface {
	Parse(inputText string) (*asts.AST[T], error)
}
