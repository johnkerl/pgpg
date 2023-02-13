package parsers

import (
	"errors"
	"fmt"

	"github.com/johnkerl/pgpg/pkg/asts"
	"github.com/johnkerl/pgpg/pkg/lexers"
	"github.com/johnkerl/pgpg/pkg/tokens"
)

// Grammar:
//
// Root
//   : int_literal
//   | int_literal plus Root
//   | int_literal times Root
// ;

type AMEParser struct {
	lexer        lexers.AbstractLexer
	lookaheadToken *tokens.Token
}

func NewAMEParser() AbstractParser {
	return &AMEParser{}
}

// My goal (not the only possible goal): map input string -> tokens -> AST
func (parser *AMEParser) Parse(inputText string) (*asts.AST, error) {
	parser.lexer = lexers.NewAMLexer(inputText)
	rootNode, err := parser.parseAux()
	if err != nil {
		return nil, err
	}
	return asts.NewAST(rootNode), nil
}

func (parser *AMEParser) parseAux() (*asts.ASTNode, error) {
	// TODO: helper somehow
	parser.lookaheadToken = parser.lexer.Scan()
	if parser.lookaheadToken.IsError() {
		return nil, errors.New(string(parser.lookaheadToken.Lexeme))
	}
	if parser.lookaheadToken.IsEOF() {
		return nil, errors.New("AMEParser: no token found in input")
	}

	actualDesc, err := parser.lexer.DecodeType(parser.lookaheadToken.Type)
	if err != nil {
		return nil, err
	}
	expectedDesc, err := parser.lexer.DecodeType(lexers.AMLexerTypeNumber)
	if err != nil {
		return nil, err
	}

	if parser.lookaheadToken.Type != lexers.AMLexerTypeNumber {
		return nil, fmt.Errorf(
			"AMEParser: initial token was of type %s; expected %s",
			actualDesc,
			expectedDesc,
		)
	}

	// TODO: helper somehow
	acceptedToken := parser.lookaheadToken
	parser.lookaheadToken = parser.lexer.Scan()
	if parser.lookaheadToken.IsError() {
		return nil, errors.New(string(parser.lookaheadToken.Lexeme))
	}

	if parser.lookaheadToken.IsEOF() {
		// The entire expression is a single number
		node := asts.NewASTNodeZaryNestable(acceptedToken) // TODO: type
		if err != nil {
			return nil, err
		}
		return node, nil
	}

	// Make a binary node with parent being the plus or times operator, the left child being the
	// previous token, and the right child being TBD.
	parent := asts.NewASTNodeZaryNestable(parser.lookaheadToken) // TODO: type
	leftChild := asts.NewASTNodeZaryNestable(acceptedToken)    // TODO: type
	rightChild, err := parser.parseAux()
	if err != nil {
		return nil, err
	}
	parent.AddChild(leftChild)
	parent.AddChild(rightChild)

	return parent, nil
}
