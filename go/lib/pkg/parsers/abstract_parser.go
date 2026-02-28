package parsers

import (
	"io"

	"github.com/johnkerl/pgpg/go/lib/pkg/asts"
)

type AbstractParser interface {
	Parse(r io.Reader) (*asts.AST, error)
}
