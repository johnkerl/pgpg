package parsers

import "github.com/johnkerl/pgpg/pkg/asts"

type AbstractParser interface {
	Parse(inputText string) (*asts.AST, error)
}
