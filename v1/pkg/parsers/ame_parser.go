package parsers

import (
	"errors"
	"fmt"

	"github.com/johnkerl/pgpg/pkg/asts"
	"github.com/johnkerl/pgpg/pkg/lexers"
)

// Grammar:
//
// Root
//   : int_literal
//   | int_literal plus Root
//   | int_literal times Root
// ;

type AMEParser struct {
	lexer *lexers.LookaheadLexer
}

func NewAMEParser() AbstractParser {
	return &AMEParser{}
}

// My goal (not the only possible goal): map input string -> tokens -> AST
func (parser *AMEParser) Parse(inputText string) (*asts.AST, error) {
	parser.lexer = lexers.NewLookaheadLexer(lexers.NewAMLexer(inputText))
	rootNode, err := parser.parseSumOrProduct()
	if err != nil {
		return nil, err
	}
	return asts.NewAST(rootNode), nil
}

func (parser *AMEParser) parseSumOrProduct() (*asts.ASTNode, error) {
	lookaheadToken := parser.lexer.LookAhead()

	// TODO: methodize
	if lookaheadToken.IsError() {
		return nil, errors.New(string(lookaheadToken.Lexeme))
	}
	if lookaheadToken.IsEOF() {
		return nil, errors.New("AMEParser: no token found in input")
	}

	if lookaheadToken.Type != lexers.AMLexerTypeNumber {
		actualDesc, err := parser.lexer.DecodeType(lookaheadToken.Type)
		if err != nil {
			return nil, err
		}
		expectedDesc, err := parser.lexer.DecodeType(lexers.AMLexerTypeNumber)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf(
			"AMEParser: initial token was of type %s; expected %s",
			actualDesc,
			expectedDesc,
		)
	}

	acceptedToken := parser.lexer.Advance()
	opToken := parser.lexer.Advance()
	if opToken.IsError() {
		return nil, errors.New(string(opToken.Lexeme))
	}

	if opToken.IsEOF() {
		// The entire expression is a single number
		node := asts.NewASTNodeZaryNestable(acceptedToken) // TODO: type
		return node, nil
	}

	// Make a binary node with parent being the plus or times operator, the left child being the
	// previous token, and the right child being TBD.
	parent := asts.NewASTNodeZaryNestable(opToken)   // TODO: type
	leftChild := asts.NewASTNodeZaryNestable(acceptedToken) // TODO: type
	rightChild, err := parser.parseSumOrProduct()
	if err != nil {
		return nil, err
	}
	parent.AddChild(leftChild)
	parent.AddChild(rightChild)

	return parent, nil
}
