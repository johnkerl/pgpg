package parsers

import "github.com/johnkerl/pgpg/manual/go/pkg/asts"

type AbstractParser interface {
	Parse(inputText string) (*asts.AST, error)
}
