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
	currentToken *tokens.Token
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
	parser.currentToken = parser.lexer.Scan()
	if parser.currentToken.IsError() {
		return nil, errors.New(string(parser.currentToken.Lexeme))
	}
	if parser.currentToken.IsEOF() {
		return nil, errors.New("AMEParser: no token found in input")
	}

	actualDesc, err := parser.lexer.DecodeType(parser.currentToken.Type)
	if err != nil {
		return nil, err
	}
	expectedDesc, err := parser.lexer.DecodeType(lexers.AMLexerTypeNumber)
	if err != nil {
		return nil, err
	}

	if parser.currentToken.Type != lexers.AMLexerTypeNumber {
		return nil, fmt.Errorf(
			"AMEParser: initial token was of type %s; expected %s",
			actualDesc,
			expectedDesc,
		)
	}

	// TODO: helper somehow
	previousToken := parser.currentToken
	parser.currentToken = parser.lexer.Scan()
	if parser.currentToken.IsError() {
		return nil, errors.New(string(parser.currentToken.Lexeme))
	}

	// The entire expression is a single number
	if parser.currentToken.IsEOF() {
		node := asts.NewASTNodeZaryNestable(previousToken) // TODO: type
		if err != nil {
			return nil, err
		}
		return node, nil
	}

	// Make a binary node with parent being the plus or times operator, the left child being the
	// previous token, and the right child being TBD.
	parent := asts.NewASTNodeZaryNestable(parser.currentToken) // TODO: type
	leftChild := asts.NewASTNodeZaryNestable(previousToken)    // TODO: type
	rightChild, err := parser.parseAux()
	if err != nil {
		return nil, err
	}
	parent.AddChild(leftChild)
	parent.AddChild(rightChild)

	return parent, nil
}
