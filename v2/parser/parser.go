package parser

import (
	"github.com/johnkerl/pgpg/v2/internal/errutil"
	"github.com/johnkerl/pgpg/v2/lexer"
)

// Result captures parser output and diagnostics.
type Result struct {
	AST  any
	Warn []string
}

// Parser consumes a token stream.
type Parser interface {
	Parse() (Result, error)
}

// New returns a parser for the provided lexer.
// TODO: implement LR/LALR item sets and table-driven parsing.
func New(_ lexer.Lexer) (Parser, error) {
	return nil, errutil.NotImplemented("parser.New")
}
