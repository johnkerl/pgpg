package parsers

import (
	"github.com/johnkerl/pgpg/pkg/asts"
	"github.com/johnkerl/pgpg/pkg/lexers"
)

type AMEParser struct {
	lexer lexers.AbstractLexer
	ast   *asts.AST
}

// My goal (not the only possible goal): map input string -> tokens -> AST

// Abstraction: AST as ctor arg (I think not) or as impl struct state?

func NewAMEParser() AbstractParser {
	return &AMEParser{}
}

func (parser *AMEParser) Parse(inputText string) (*asts.AST, error) {
	parser.lexer = lexers.NewAMLexer(inputText)
	node, err := asts.NewASTNodeZary(nil) // TODO: type
	parser.ast = asts.NewAST(node)
	if err != nil {
		return nil, err
	}

	return parser.ast, nil
}

// _decdig : '0'-'9' ;
// !whitespace : ' ' | '\t' | '\n' | '\r' ;
// !comment : '#'  {.} '\n' ;
//
// int_literal : _decdig { _decdig };
// plus : '+';
// times : '*';
//
// Root
//   : int_literal
//   | int_literal plus Root
//   | int_literal times Root
// ;
