package parsers

import "github.com/johnkerl/pgpg/manual/pkg/asts"

type AbstractParser interface {
	Parse(inputText string) (*asts.AST, error)
}
