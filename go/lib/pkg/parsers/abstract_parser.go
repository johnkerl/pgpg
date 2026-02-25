package parsers

import "github.com/johnkerl/pgpg/go/lib/pkg/asts"

type AbstractParser interface {
	Parse(inputText string) (*asts.AST, error)
}
